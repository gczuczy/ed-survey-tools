package vsds

import (
	"slices"
	"sort"
	"time"

	"github.com/gczuczy/ed-survey-tools/pkg/bundles"
	"github.com/gczuczy/ed-survey-tools/pkg/config"
	"github.com/gczuczy/ed-survey-tools/pkg/db"
	"github.com/gczuczy/ed-survey-tools/pkg/edsm"
	"github.com/gczuczy/ed-survey-tools/pkg/gcp"
	"github.com/gczuczy/ed-survey-tools/pkg/log"
	"github.com/gczuczy/ed-survey-tools/pkg/types"
	vsdstypes "github.com/gczuczy/ed-survey-tools/pkg/vsds/types"
)

const (
	typeGoogleSheet = "application/vnd.google-apps.spreadsheet"
	typeXlsx        = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	typeCSV         = "text/csv"
	typeODS         = "application/vnd.oasis.opendocument.spreadsheet"
)

var excludedTabs = []string{
	"Blank", "Blank CW", "Summary",
	"Master", "MASTER", "Master CW",
}

func filterSheets(sheets []gcp.Sheet) []gcp.Sheet {
	result := make([]gcp.Sheet, 0, len(sheets))
	for _, s := range sheets {
		if !slices.Contains(excludedTabs, s.GetName()) {
			result = append(result, s)
		}
	}
	return result
}

type Processor struct {
	cfg    *config.VSDSConfig
	edsm   *edsm.EDSM
	logger log.Logger
	stopCh chan struct{}
	doneCh chan struct{}
}

func NewProcessor(cfg *config.VSDSConfig, edsm *edsm.EDSM) *Processor {
	return &Processor{
		cfg:    cfg,
		edsm:   edsm,
		logger: log.GetLogger("VSDSProcessor"),
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}
}

func (p *Processor) Start() {
	go p.run()
}

func (p *Processor) Stop() {
	close(p.stopCh)
	<-p.doneCh
}

func (p *Processor) run() {
	defer func() {
		p.logger.Info().Msg("Processor stopped")
		close(p.doneCh)
	}()
	p.logger.Info().Msg("Processor started")

	for {
		select {
		case <-p.stopCh:
			return
		default:
		}

		job, err := db.Pool.FetchPendingFolderProcessing()
		if err != nil {
			p.logger.Error().Err(err).Msg("Error fetching pending job")
			select {
			case <-p.stopCh:
				return
			case <-time.After(p.cfg.ProcessorInterval):
			}
			continue
		}

		if job == nil {
			select {
			case <-p.stopCh:
				return
			case <-time.After(p.cfg.ProcessorInterval):
			}
			continue
		}

		p.process(job)
		return
	}
}

func (p *Processor) process(job *vsdstypes.FolderProcessingJob) {
	affectedProjects := make(map[int]struct{})

	defer func() {
		if ferr := db.Pool.FinishFolderProcessing(
			job.ProcID); ferr != nil {
			p.logger.Error().Err(ferr).Caller().
				Int("procid", job.ProcID).
				Msg("Error finishing folder processing")
		}
		rerr := db.Pool.RefreshSurveyMaterializedViews()
		if rerr != nil {
			p.logger.Error().Err(rerr).Caller().
				Int("procid", job.ProcID).
				Msg("Error refreshing materialized views")
		}
		if len(affectedProjects) > 0 {
			pids := make([]int, 0, len(affectedProjects))
			for pid := range affectedProjects {
				pids = append(pids, pid)
			}
			if err := db.Pool.QueueAutoRegenBundles(pids); err != nil {
				p.logger.Error().Err(err).
					Msg("Error queuing auto-regen bundles")
			}
			bundles.Signal()
		}
	}()
	p.logger.Info().
		Int("procid", job.ProcID).
		Int("folderid", job.FolderID).
		Str("gcpid", job.GCPID).
		Msg("Starting folder processing")

	var (
		txn           *db.Transaction
		err           error
		gss           *gcp.GSpreadsheetsService
		survey        vsdstypes.Survey
		spreadsheetID int
		sheetID       int
		surveyHash    uint64
		variantSvc    *VariantService
	)

	surveyCache := types.NewSet[uint64]()

	if gss, err = gcp.NewSheets(); err != nil {
		p.logger.Error().Err(err).
			Int("procid", job.ProcID).
			Msg("Error getting GCP Spreadsheets service")
		return
	}

	if txn, err = db.Pool.StartLongTxn(); err != nil {
		p.logger.Error().Err(err).
			Int("procid", job.ProcID).
			Msg("Error starting long transaction")
		return
	}
	defer txn.Close()

	sysCache := db.NewSystemCache(txn, p.edsm)

	if variantSvc, err = NewVariantService(txn); err != nil {
		p.logger.Error().Err(err).
			Int("procid", job.ProcID).
			Msg("Error loading sheet variants")
		return
	}

	items, err := gcp.ListFolder(job.GCPID)
	if err != nil {
		p.logger.Error().Err(err).
			Int("procid", job.ProcID).
			Str("gcpid", job.GCPID).
			Msg("Error listing folder")
		return
	}

	p.logger.Info().
		Int("procid", job.ProcID).
		Int("count", len(items)).
		Msg("Folder contents fetched")

	// Delete & replace: wipe previous data for this folder so
	// re-runs start from a clean slate.
	if err = txn.DeleteFolderSpreadsheets(job.FolderID); err != nil {
		p.logger.Error().Err(err).
			Int("procid", job.ProcID).
			Int("folderid", job.FolderID).
			Msg("Error deleting existing folder spreadsheets")
		return
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedTime > items[j].CreatedTime
	})

	for _, item := range items {
		p.logger.Info().
			Int("procid", job.ProcID).
			Str("file_id", item.ID).
			Str("name", item.Name).
			Str("content_type", item.MimeType).
			Str("created_at", item.CreatedTime).
			Str("modified_at", item.ModifiedTime).
			Msg("Processing item")

		var sheets []gcp.Sheet
		switch item.MimeType {
		case typeGoogleSheet:
			sheets, err = openGoogleSheets(gss, item)
		case typeXlsx:
			sheets, err = openXlsxSheets(item)
		case typeODS:
			sheets, err = openOdsSheets(item)
		case typeCSV:
			sheets, err = openCsvSheets(item)
		default:
			p.logger.Error().
				Int("procid", job.ProcID).
				Str("file_id", item.ID).
				Str("name", item.Name).
				Str("content_type", item.MimeType).
				Msg("Type not implemented")
			continue
		}
		if err != nil {
			p.logger.Error().Err(err).
				Int("procid", job.ProcID).
				Str("file_id", item.ID).
				Str("name", item.Name).
				Msg("Unable to open spreadsheet")
			continue
		}

		spreadsheetID, err = txn.AddSpreadsheet(
			job.FolderID, item.ID, item.Name, item.MimeType,
		)
		if err != nil {
			p.logger.Error().Err(err).
				Int("procid", job.ProcID).
				Str("file_id", item.ID).
				Str("name", item.Name).
				Msg("Error registering spreadsheet in DB")
			continue
		}

		toProcess := filterSheets(sheets)
		if item.MimeType == typeCSV {
			toProcess = sheets
		}
		for _, sheet := range toProcess {
			tabName := sheet.GetName()

			sheetID, err = txn.AddSheet(spreadsheetID, &tabName)
			if err != nil {
				p.logger.Error().Err(err).
					Int("procid", job.ProcID).
					Str("file_id", item.ID).
					Str("tab", tabName).
					Msg("Error registering sheet in DB")
				continue
			}

			survey, err = ParseSheet(sheet, variantSvc)
			if err != nil {
				p.logger.Error().Err(err).
					Int("procid", job.ProcID).
					Str("file_id", item.ID).
					Str("tab", tabName).
					Msg("Error parsing sheet")
				p.recordResult(txn, job.ProcID, sheetID,
					false, err.Error(), survey.CMDR)
				continue
			}

			surveyHash = survey.Hash()
			if surveyCache.IsSet(surveyHash) {
				continue
			}
			surveyCache.Set(surveyHash)

			// Collect system names for bulk lookup.
			names := make([]string, len(survey.SurveyPoints))
			for i, sp := range survey.SurveyPoints {
				names[i] = sp.SystemName
			}

			// Savepoint before system lookups so that any
			// partial DB inserts can be rolled back cleanly
			// on failure.
			if err = txn.Savepoint(); err != nil {
				p.logger.Error().Err(err).
					Int("procid", job.ProcID).
					Str("tab", tabName).
					Msg("Error creating savepoint")
				continue
			}

			systems, lErr := sysCache.Lookup(names)
			if lErr != nil {
				if len(systems) == 0 {
					// Fatal error or every system
					// unresolvable; roll back any partial
					// inserts and skip this tab.
					p.logger.Error().Err(lErr).
						Int("procid", job.ProcID).
						Str("file_id", item.ID).
						Str("tab", tabName).
						Msg("Error resolving system names")
					if rerr := txn.Rollback(); rerr != nil {
						p.logger.Error().Err(rerr).
							Int("procid", job.ProcID).
							Msg("Error rolling back transaction")
					}
					p.recordResult(txn, job.ProcID, sheetID,
						false, lErr.Error(), survey.CMDR)
					continue
				}
				// Partial resolution: some systems found.
				// Keep partial inserts; errors recorded
				// in the final sheet result.
				p.logger.Warn().Err(lErr).
					Int("procid", job.ProcID).
					Str("file_id", item.ID).
					Str("tab", tabName).
					Int("resolved", len(systems)).
					Msg("Some system names could not be resolved")
			}

			// Lock in newly inserted system rows
			// (including the partial set on error).
			if err = txn.Savepoint(); err != nil {
				p.logger.Error().Err(err).
					Int("procid", job.ProcID).
					Str("tab", tabName).
					Msg("Error creating savepoint")
				continue
			}

			sysMap := make(map[string]db.System, len(systems))
			for _, s := range systems {
				sysMap[s.Name] = s
			}

			sysRealY := make(map[string]float32, len(sysMap))
			for name, sys := range sysMap {
				sysRealY[name] = sys.Y
			}
			for _, dp := range survey.Normalize(sysRealY) {
				p.logger.Warn().
					Int("procid", job.ProcID).
					Str("file_id", item.ID).
					Str("tab", tabName).
					Str("system", dp.SystemName).
					Int("zsample", dp.ZSample).
					Msg("Duplicate system dropped from survey")
			}

			projectID, pErr := txn.LookupProject(survey.Project)
			if pErr != nil {
				p.logger.Error().Err(pErr).
					Int("procid", job.ProcID).
					Str("tab", tabName).
					Str("project", survey.Project).
					Msg("Error looking up project")
				p.recordResult(txn, job.ProcID, sheetID,
					false, pErr.Error(), survey.CMDR)
				continue
			}

			rawCmdrID, cErr := txn.UpsertCmdr(survey.CMDR)
			if cErr != nil {
				p.logger.Error().Err(cErr).
					Int("procid", job.ProcID).
					Str("tab", tabName).
					Str("cmdr", survey.CMDR).
					Msg("Error upserting CMDR")
				p.recordResult(txn, job.ProcID, sheetID,
					false, cErr.Error(), survey.CMDR)
				continue
			}
			var cmdrID *int
			if rawCmdrID != 0 {
				cmdrID = &rawCmdrID
			}

			if sErr := txn.AddSurvey(
				sheetID, projectID, cmdrID,
				survey.SurveyPoints, sysMap,
			); sErr != nil {
				p.logger.Error().Err(sErr).
					Int("procid", job.ProcID).
					Str("tab", tabName).
					Msg("Error adding survey to DB")
				p.recordResult(txn, job.ProcID, sheetID,
					false, sErr.Error(), survey.CMDR)
				continue
			}

			affectedProjects[projectID] = struct{}{}

			if lErr != nil {
				p.recordResult(txn, job.ProcID, sheetID,
					false, lErr.Error(), survey.CMDR)
			} else {
				p.recordResult(txn, job.ProcID, sheetID,
					true, "", survey.CMDR)
			}
		}
	}
}

// recordResult writes a sheet_processing row. Errors are logged but
// do not alter the processing flow of the caller.  cmdrName is used
// for a best-effort CMDR lookup; pass survey.CMDR (may be empty).
func (p *Processor) recordResult(
	txn *db.Transaction,
	procID, sheetID int,
	success bool,
	message string,
	cmdrName string,
) {
	if err := txn.RecordSheetResult(
		procID, sheetID, success, message, cmdrName,
	); err != nil {
		p.logger.Error().Err(err).
			Int("procid", procID).
			Int("sheetid", sheetID).
			Msg("Error recording sheet processing result")
	}
}
