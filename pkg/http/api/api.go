package api

import (
	"fmt"
	"errors"

	"github.com/gorilla/mux"

	"github.com/gczuczy/ed-survey-tools/pkg/config"
	"github.com/gczuczy/ed-survey-tools/pkg/http/api/auth"
)

func Init(r *mux.Router, cfg *config.OAuth2Config) error {
	var err error
	sr := r.PathPrefix("/auth").Subrouter()
	if err = auth.Init(sr, cfg); err != nil {
		return errors.Join(err, fmt.Errorf("Unable init auth endpoints"))
	}

	return nil
}
