package vsds

import (
	"github.com/gorilla/mux"

	w "github.com/gczuczy/ed-survey-tools/pkg/http/wrappers"
)

func Init(r *mux.Router) error {
	r.Handle("/config",
		w.NewAPIHandler().AuthGet(getConfig, w.IsAdmin),
	)

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

	// sheet variants
	r.Handle("/projects/{id:[0-9]+}/variants",
		w.NewAPIHandler().
			AuthGet(listVariants, w.IsAdmin).
			AuthPut(addVariant, w.IsAdmin),
	)
	r.Handle(
		"/projects/{id:[0-9]+}/variants/validate",
		w.NewAPIHandler().AuthPost(validateVariant, w.IsAdmin),
	)
	r.Handle("/projects/{id:[0-9]+}/variants/{vid:[0-9]+}",
		w.NewAPIHandler().
			AuthPost(updateVariant, w.IsAdmin).
			AuthDelete(deleteVariant, w.IsAdmin),
	)
	r.Handle(
		"/projects/{id:[0-9]+}/variants/{vid:[0-9]+}/checks",
		w.NewAPIHandler().
			AuthPut(addVariantCheck, w.IsAdmin),
	)
	r.Handle(
		"/projects/{id:[0-9]+}/variants/{vid:[0-9]+}/checks/{cid:[0-9]+}",
		w.NewAPIHandler().
			AuthDelete(deleteVariantCheck, w.IsAdmin),
	)

	r.Handle("/contribution",
		w.NewAPIHandler().AuthGet(getContribution),
	)
	r.Handle("/contribution/errors",
		w.NewAPIHandler().AuthGet(getContributionErrors),
	)

	r.Handle("/visualization/sectors",
		w.NewAPIHandler().AuthGet(listSectors),
	)
	return nil
}
