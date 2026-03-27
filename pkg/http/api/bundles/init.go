package bundles

import (
	"github.com/gorilla/mux"

	w "github.com/gczuczy/ed-survey-tools/pkg/http/wrappers"
)

func Init(r *mux.Router) error {
	r.Handle("",
		w.NewAPIHandler().
			Get(listBundles).
			AuthPut(createBundle, w.IsAdmin),
	)
	r.Handle("/{id:[0-9]+}",
		w.NewAPIHandler().
			Get(getBundle).
			AuthDelete(deleteBundle, w.IsAdmin).
			AuthPatch(patchBundle, w.IsAdmin),
	)
	r.Handle("/{id:[0-9]+}/generate",
		w.NewAPIHandler().AuthPost(generateBundle, w.IsAdmin),
	)
	return nil
}
