package http

import (
	"os"
	"fmt"
	"errors"
	"context"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/gczuczy/ed-survey-tools/pkg/http/api"
	"github.com/gczuczy/ed-survey-tools/pkg/http/sessions"
	"github.com/gczuczy/ed-survey-tools/pkg/config"
	"github.com/gczuczy/ed-survey-tools/pkg/http/wrappers"
	"github.com/gczuczy/ed-survey-tools/pkg/log"
)

var (
	l log.Logger
)

type HTTPService struct {
	config *config.HTTPConfig
	srv http.Server
}

func New(cfg *config.Config) (*HTTPService, error) {

	l = log.GetLogger("http")
	wrappers.SetLogger(l)

	if err := sessions.Init(&cfg.Sessions); err != nil {
		l.Error().Err(err).Msg("Unable to init sessions")
		return nil, errors.Join(err, fmt.Errorf("Unable to init http sessions"))
	}

	spa, err := newSPAHandler()
	if err != nil {
		l.Error().Err(err).Msg("Unable to init SPA tarball")
		return nil, errors.Join(err, fmt.Errorf("Unable to init SPA tarball"))
	}

	router := mux.NewRouter()
	apisr := router.PathPrefix("/api").Subrouter()
	if err = api.Init(apisr, &cfg.OAuth2, &cfg.Bundles); err != nil {
		l.Error().Err(err).Msg("API init error")
		return nil, err
	}
	if cfg.Bundles.Serve {
		router.PathPrefix("/static/").
			Handler(newStaticBundleHandler(cfg.Bundles.Path))
	}
	router.PathPrefix(`/`).Handler(spa)

	hs := HTTPService{
		config: &cfg.HTTP,
		srv: http.Server{
			Addr: fmt.Sprintf(":%d", cfg.HTTP.Port),
			Handler: router,
		},
	}

	l.Info().Msg("HTTP subsystem initialized")
	return &hs, nil
}

func Close() error {
	return sessions.Close()
}

func (hs *HTTPService) Run() error {

	go func() {
		err := hs.srv.ListenAndServe()
		if err != http.ErrServerClosed {
			l.Error().Err(err).Msg("http.Servce() error")
			os.Exit(2)
		}
	}()
	return nil
}

func (hs *HTTPService) Shutdown() error {
	ctx := context.Background()

	return hs.srv.Shutdown(ctx)
}

func (hs *HTTPService) Close() error {
	return hs.srv.Close()
}
