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
)

type HTTPService struct {
	config *config.HTTPConfig
	srv http.Server
}

func New(cfg *config.Config) (*HTTPService, error) {

	if err := sessions.Init(&cfg.Sessions); err != nil {
		return nil, errors.Join(err, fmt.Errorf("Unable top init http sessions"))
	}

	spa, err := newSPAHandler()
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("Unable to init SPA tarball"))
	}

	router := mux.NewRouter()
	apisr := router.PathPrefix("/api").Subrouter()
	if err = api.Init(apisr, &cfg.OAuth2); err != nil {
		return nil, err
	}
	router.PathPrefix(`/`).Handler(spa)

	hs := HTTPService{
		config: &cfg.HTTP,
		srv: http.Server{
			Addr: fmt.Sprintf(":%d", cfg.HTTP.Port),
			Handler: router,
		},
	}

	return &hs, nil
}

func Close() error {
	return sessions.Close()
}

func (hs *HTTPService) Run() error {

	go func() {
		err := hs.srv.ListenAndServe()
		if err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "http.Serve(): %v\n", err)
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
