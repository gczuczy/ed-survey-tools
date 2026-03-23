package cmdrs

import (
	"github.com/gorilla/mux"

	w "github.com/gczuczy/ed-survey-tools/pkg/http/wrappers"
)

func Init(r *mux.Router) error {
	r.Handle("",
		w.NewAPIHandler().AuthGet(listCmdrs, w.IsOwner),
	)
	r.Handle("/{id:[0-9]+}",
		w.NewAPIHandler().AuthPatch(setCmdrAdmin, w.IsOwner),
	)
	return nil
}
