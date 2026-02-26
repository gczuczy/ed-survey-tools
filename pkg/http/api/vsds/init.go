package vsds

import (
	"github.com/gorilla/mux"

	w "github.com/gczuczy/ed-survey-tools/pkg/http/wrappers"
)

func Init(r *mux.Router) error {
	r.Handle("/projects",
		w.NewAPIHandler().Get(listProjects).AuthPut(addProject, w.IsAdmin),
	)
	return nil
}
