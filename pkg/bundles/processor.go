package bundles

import (
	"path/filepath"
	"time"

	"github.com/gczuczy/ed-survey-tools/pkg/config"
	"github.com/gczuczy/ed-survey-tools/pkg/db"
	"github.com/gczuczy/ed-survey-tools/pkg/log"
)

// P is the package-level processor instance, set by service.go.
var P *Processor

// Signal wakes the processor immediately to check for queued bundles.
// Safe to call before Start(); the signal is discarded if P is nil.
func Signal() {
	if P == nil {
		return
	}
	select {
	case P.sigCh <- struct{}{}:
	default:
	}
}

// Processor polls for pending bundles and generates them sequentially.
type Processor struct {
	cfg    *config.BundlesConfig
	logger log.Logger
	stopCh chan struct{}
	doneCh chan struct{}
	sigCh  chan struct{}
}

// NewProcessor creates a Processor using the given configuration.
func NewProcessor(cfg *config.BundlesConfig) *Processor {
	return &Processor{
		cfg:    cfg,
		logger: log.GetLogger("BundleProcessor"),
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
		sigCh:  make(chan struct{}, 1),
	}
}

// Start launches the processor goroutine.
func (p *Processor) Start() {
	go p.run()
}

// Stop signals the processor to stop and waits for it to exit.
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

	p.process()

	ticker := time.NewTicker(p.cfg.CheckInterval)
	defer ticker.Stop()
	for {
		select {
		case <-p.stopCh:
			return
		case <-p.sigCh:
			p.process()
		case <-ticker.C:
			p.process()
		}
	}
}

func (p *Processor) process() {
	bundles, err := db.Pool.ListPendingBundles()
	if err != nil {
		p.logger.Error().Err(err).
			Msg("Error listing pending bundles")
		return
	}

	for _, b := range bundles {
		p.generate(b)
	}
}

func (p *Processor) generate(b db.Bundle) {
	if err := db.Pool.SetBundleGenerating(b.ID); err != nil {
		// Another instance claimed it; skip silently.
		return
	}

	runner, ok := get(b.MeasurementType)
	if !ok {
		p.logger.Error().
			Int("id", b.ID).
			Str("type", b.MeasurementType).
			Msg("No runner registered for measurement type")
		_ = db.Pool.SetBundleError(b.ID,
			"no runner for type: "+b.MeasurementType)
		return
	}

	cfg, err := runner.LoadConfig(b.ID)
	if err != nil {
		p.logger.Error().Err(err).
			Int("id", b.ID).
			Msg("Error loading bundle config")
		_ = db.Pool.SetBundleError(b.ID, err.Error())
		return
	}

	data, err := runner.Generate(cfg)
	if err != nil {
		p.logger.Error().Err(err).
			Int("id", b.ID).
			Msg("Error generating bundle")
		_ = db.Pool.SetBundleError(b.ID, err.Error())
		return
	}

	destPath := filepath.Join(p.cfg.Path, b.Filename)
	if err = write(data, destPath); err != nil {
		p.logger.Error().Err(err).
			Int("id", b.ID).
			Str("path", destPath).
			Msg("Error writing bundle file")
		_ = db.Pool.SetBundleError(b.ID, err.Error())
		return
	}

	if err = db.Pool.SetBundleReady(b.ID); err != nil {
		p.logger.Error().Err(err).
			Int("id", b.ID).
			Msg("Error marking bundle ready")
	}

	p.logger.Info().
		Int("id", b.ID).
		Str("file", b.Filename).
		Msg("Bundle generated")
}
