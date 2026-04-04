package config

import (
	"github.com/gorilla/mux"

	w "github.com/gczuczy/ed-survey-tools/pkg/http/wrappers"
)

func InitRoutes(r *mux.Router) error {
	r.Handle("", w.NewAPIHandler().Get(getConfig))
	return nil
}
