package vsds

import (
	"github.com/gorilla/mux"

	w "github.com/gczuczy/ed-survey-tools/pkg/http/wrappers"
)

func Init(r *mux.Router) error {
	r.Handle("/folders",
		w.NewAPIHandler().AuthGet(listFolders, w.IsAdmin).AuthPost(addFolder, w.IsAdmin),
	)
	r.Handle("/folders/{id:[0-9]+}",
		w.NewAPIHandler().AuthDelete(deleteFolder, w.IsAdmin),
	)
	r.Handle("/folders/{id:[0-9]+}/process",
		w.NewAPIHandler().AuthPost(processFolder, w.IsAdmin),
	)
	r.Handle("/folders/{id:[0-9]+}/extraction",
		w.NewAPIHandler().AuthGet(getFolderExtractionSummary, w.IsAdmin),
	)

	r.Handle("/projects",
		w.NewAPIHandler().Get(listProjects).AuthPut(addProject, w.IsAdmin),
	)
	r.Handle("/projects/{id:[0-9]+}",
		w.NewAPIHandler().Get(getProject),
	)
	// zsamples
	r.Handle("/projects/{id:[0-9]+}/zsamples",
		w.NewAPIHandler().AuthPost(setZSamples, w.IsAdmin),
	)

	r.Handle("/projects/{id:[0-9]+}/zsamples/{zsample:-?[0-9]+}",
		w.NewAPIHandler().AuthPut(addZSample, w.IsAdmin).
			AuthDelete(deleteZSample, w.IsAdmin),
	)
	return nil
}
