package api

import (
	"fmt"
	"errors"

	"github.com/gorilla/mux"

	"github.com/gczuczy/ed-survey-tools/pkg/config"
	"github.com/gczuczy/ed-survey-tools/pkg/http/api/auth"
	apibundles "github.com/gczuczy/ed-survey-tools/pkg/http/api/bundles"
	apicfg "github.com/gczuczy/ed-survey-tools/pkg/http/api/config"
	"github.com/gczuczy/ed-survey-tools/pkg/http/api/cmdrs"
	"github.com/gczuczy/ed-survey-tools/pkg/http/api/vsds"
)

func Init(
	r *mux.Router,
	cfg *config.OAuth2Config,
	bcfg *config.BundlesConfig,
) error {
	var err error
	sr := r.PathPrefix("/auth").Subrouter()
	if err = auth.Init(sr, cfg); err != nil {
		return errors.Join(err, fmt.Errorf("Unable init auth endpoints"))
	}

	sr = r.PathPrefix("/vsds").Subrouter()
	if err = vsds.Init(sr); err != nil {
		return errors.Join(err, fmt.Errorf("Unable init vsds endpoints"))
	}

	sr = r.PathPrefix("/cmdrs").Subrouter()
	if err = cmdrs.Init(sr); err != nil {
		return errors.Join(
			err, fmt.Errorf("Unable init cmdrs endpoints"))
	}

	apicfg.Init(bcfg)
	sr = r.PathPrefix("/config").Subrouter()
	if err = apicfg.InitRoutes(sr); err != nil {
		return errors.Join(
			err, fmt.Errorf("Unable init config endpoints"))
	}

	sr = r.PathPrefix("/bundles").Subrouter()
	if err = apibundles.Init(sr); err != nil {
		return errors.Join(
			err, fmt.Errorf("Unable init bundles endpoints"))
	}

	return nil
}
