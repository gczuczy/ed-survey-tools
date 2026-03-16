package vsds

import (
	"sort"
	"time"

	"github.com/gczuczy/ed-survey-tools/pkg/config"
	"github.com/gczuczy/ed-survey-tools/pkg/db"
	"github.com/gczuczy/ed-survey-tools/pkg/gcp"
	"github.com/gczuczy/ed-survey-tools/pkg/log"
	"github.com/gczuczy/ed-survey-tools/pkg/types"
	vsdstypes "github.com/gczuczy/ed-survey-tools/pkg/vsds/types"
)

const (
	typeGoogleSheet = "application/vnd.google-apps.spreadsheet"
	typeXlsx = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	typeCSV = "text/csv"
	typeODS = "application/vnd.oasis.opendocument.spreadsheet"
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
		txn *db.Transaction
		err error
		gss *gcp.GSpreadsheetsService
		ss gcp.GSpreadSheet
		survey vsdstypes.Survey
	)

	surveyCache := types.NewSet[uint64]()

	if gss, err = gcp.NewSheets(); err != nil {
		p.logger.Error().Err(err).
			Int("procid", job.ProcID).
			Msg("Error getting GCP Spreadsheets service")
		return
	}
	_ = gss

	if txn, err = db.Pool.StartLongTxn(); err != nil {
		p.logger.Error().Err(err).
			Int("procid", job.ProcID).
			Msg("Error starting long transaction")
		return
	}
	defer txn.Close()

	sysCache := db.NewSystemCache(txn)
	_ = sysCache

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
		if item.MimeType == typeGoogleSheet {
			// load the spreadsheet
			if ss, err = gss.Sheet(item.ID); err != nil {
				p.logger.Error().Err(err).
					Int("procid", job.ProcID).
					Str("file_id", item.ID).
					Str("name", item.Name).
					Msg("Unable to open Goolge Spreadsheet")
				continue
			}
			// iterate over the sheets
			if sheets, err = ss.GetSheets(); err != nil {
				p.logger.Error().Err(err).
					Int("procid", job.ProcID).
					Str("file_id", item.ID).
					Str("name", item.Name).
					Msg("Unable to load Goolge Spreadsheet's sheets")
				continue
			}

			for sheet := range sheets {
				if survey, err = vsds.ParseSheet(sheet); err != nil {
					p.logger.Error().Err(err).
						Int("procid", job.ProcID).
						Str("file_id", item.ID).
						Str("name", item.Name).
						Msg("Error parsing sheet")
				}
				// check whether we already have it
				surveyHash = survey.Hash()
				if surveyCache.IsSet(surveyHash) {
					continue
				}
				// we insert it
				surveyCache.Set(surveyHash)
				// TODO resolve systemnames
				// TODO txn-add
			}
		} else {
			p.logger.Error().
				Int("procid", job.ProcID).
				Str("file_id", item.ID).
				Str("name", item.Name).
				Str("content_type", item.MimeType).
				Msg("Type not implemented")
		}
	}
}
