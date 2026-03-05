package vsds

import (
	"time"

	"github.com/gczuczy/ed-survey-tools/pkg/config"
	"github.com/gczuczy/ed-survey-tools/pkg/db"
	"github.com/gczuczy/ed-survey-tools/pkg/gcp"
	"github.com/gczuczy/ed-survey-tools/pkg/log"
	vsdstypes "github.com/gczuczy/ed-survey-tools/pkg/vsds/types"
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
	p.logger.Info().
		Int("procid", job.ProcID).
		Int("folderid", job.FolderID).
		Str("gcpid", job.GCPID).
		Msg("Starting folder processing")

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

	for _, item := range items {
		p.logger.Info().
			Int("procid", job.ProcID).
			Str("file_id", item.ID).
			Str("name", item.Name).
			Str("content_type", item.MimeType).
			Str("created_at", item.CreatedTime).
			Str("modified_at", item.ModifiedTime).
			Msg("Folder item")
	}
}
