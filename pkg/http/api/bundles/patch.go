package bundles

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gczuczy/ed-survey-tools/pkg/db"
	w "github.com/gczuczy/ed-survey-tools/pkg/http/wrappers"
)

type patchBundleBody struct {
	Name      *string        `json:"name"`
	AutoRegen *bool          `json:"autoregen"`
	VSDS      *vsdsPatchBody `json:"vsds"`
}

type vsdsPatchBody struct {
	Subtype     *string `json:"subtype"`
	AllProjects *bool   `json:"allprojects"`
	Projects    []int   `json:"projects"`
}

func patchBundle(r *w.Request) w.IResponse {
	idStr, ok := r.Vars["id"]
	if !ok {
		return w.NewError(
			fmt.Errorf("Missing bundle ID"),
			http.StatusBadRequest)
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return w.NewError(
			fmt.Errorf("Invalid bundle ID"),
			http.StatusBadRequest)
	}

	var body patchBundleBody
	if err := json.NewDecoder(r.R.Body).Decode(&body); err != nil {
		return w.NewError(
			fmt.Errorf("Invalid request body: %v", err),
			http.StatusBadRequest)
	}

	var (
		vsdsSubtype     *string
		vsdsAllProjects *bool
		vsdsProjects    []int
	)
	if body.VSDS != nil {
		vsdsSubtype = body.VSDS.Subtype
		vsdsAllProjects = body.VSDS.AllProjects
		vsdsProjects = body.VSDS.Projects
	}

	bundle, err := db.Pool.UpdateVSDSBundle(
		id,
		body.Name,
		body.AutoRegen,
		vsdsSubtype,
		vsdsAllProjects,
		vsdsProjects,
	)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return w.NewError(
				fmt.Errorf("Bundle not found"),
				http.StatusNotFound)
		}
		r.L.Error().Err(err).Msg("Error updating bundle")
		return w.NewError(
			errors.Join(err,
				fmt.Errorf("Error updating bundle")),
			http.StatusInternalServerError)
	}
	return w.Success(bundle)
}
