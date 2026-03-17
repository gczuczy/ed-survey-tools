package vsds

import (
	"sort"
	"time"
	"slices"

	"github.com/gczuczy/ed-survey-tools/pkg/config"
	"github.com/gczuczy/ed-survey-tools/pkg/db"
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

type Processor struct {
	cfg    *config.VSDSConfig
	logger log.Logger
	stopCh chan struct{}
	doneCh chan struct{}
}

func NewProcessor(cfg *config.VSDSConfig) *Processor {
	return &Processor{
		cfg:    cfg,
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
	defer db.Pool.FinishFolderProcessing(job.ProcID)
	p.logger.Info().
		Int("procid", job.ProcID).
		Int("folderid", job.FolderID).
		Str("gcpid", job.GCPID).
		Msg("Starting folder processing")

	var (
		txn           *db.Transaction
		err           error
		gss           *gcp.GSpreadsheetsService
		ss            *gcp.GSpreadsheet
		survey        vsdstypes.Survey
		spreadsheetID int
		sheetID       int
		sheets        []gcp.Sheet
		surveyHash    uint64
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

	sysCache := db.NewSystemCache(txn)

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

		if item.MimeType != typeGoogleSheet {
			p.logger.Error().
				Int("procid", job.ProcID).
				Str("file_id", item.ID).
				Str("name", item.Name).
				Str("content_type", item.MimeType).
				Msg("Type not implemented")
			continue
		}

		if ss, err = gss.Sheet(item.ID); err != nil {
			p.logger.Error().Err(err).
				Int("procid", job.ProcID).
				Str("file_id", item.ID).
				Str("name", item.Name).
				Msg("Unable to open Google Spreadsheet")
			continue
		}

		spreadsheetID, err = txn.AddSpreadsheet(
			job.FolderID, item.ID, item.Name, typeGoogleSheet,
		)
		if err != nil {
			p.logger.Error().Err(err).
				Int("procid", job.ProcID).
				Str("file_id", item.ID).
				Str("name", item.Name).
				Msg("Error registering spreadsheet in DB")
			continue
		}

		if sheets, err = ss.GetSheets(); err != nil {
			p.logger.Error().Err(err).
				Int("procid", job.ProcID).
				Str("file_id", item.ID).
				Str("name", item.Name).
				Msg("Unable to load spreadsheet tabs")
			continue
		}

		excludedTabs := []string{
			"Blank", "Blank CW", "Summary", "Master", "MASTER",
		}
		for _, sheet := range sheets {
			tabName := sheet.GetName()
			if slices.Contains(excludedTabs, tabName) {
				continue
			}

			sheetID, err = txn.AddSheet(spreadsheetID, &tabName)
			if err != nil {
				p.logger.Error().Err(err).
					Int("procid", job.ProcID).
					Str("file_id", item.ID).
					Str("tab", tabName).
					Msg("Error registering sheet in DB")
				continue
			}

			survey, err = ParseSheet(sheet)
			if err != nil {
				p.logger.Error().Err(err).
					Int("procid", job.ProcID).
					Str("file_id", item.ID).
					Str("tab", tabName).
					Msg("Error parsing sheet")
				p.recordResult(txn, job.ProcID, sheetID,
					false, err.Error())
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
					false, lErr.Error())
				continue
			}

			// Lock in any newly inserted system rows.
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

			systemZ := make(map[string]float32, len(sysMap))
			for name, sys := range sysMap {
				systemZ[name] = sys.Y
			}
			for _, dp := range survey.Normalize(systemZ) {
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
					false, pErr.Error())
				continue
			}

			cmdrID, cErr := txn.UpsertCmdr(survey.CMDR)
			if cErr != nil {
				p.logger.Error().Err(cErr).
					Int("procid", job.ProcID).
					Str("tab", tabName).
					Str("cmdr", survey.CMDR).
					Msg("Error upserting CMDR")
				p.recordResult(txn, job.ProcID, sheetID,
					false, cErr.Error())
				continue
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
					false, sErr.Error())
				continue
			}

			p.recordResult(txn, job.ProcID, sheetID, true, "")
		}
	}
}

// recordResult writes a sheet_processing row. Errors are logged but
// do not alter the processing flow of the caller.
func (p *Processor) recordResult(
	txn *db.Transaction,
	procID, sheetID int,
	success bool,
	message string,
) {
	if err := txn.RecordSheetResult(
		procID, sheetID, success, message,
	); err != nil {
		p.logger.Error().Err(err).
			Int("procid", procID).
			Int("sheetid", sheetID).
			Msg("Error recording sheet processing result")
	}
}
