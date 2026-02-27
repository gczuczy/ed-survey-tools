package vsds

import (
	"fmt"
	"errors"
	"net/http"
	"encoding/json"

	"github.com/gczuczy/ed-survey-tools/pkg/http/wrappers"
	"github.com/gczuczy/ed-survey-tools/pkg/db"
)

func listProjects(r *wrappers.Request) wrappers.IResponse {
	projects, err := db.Pool.ListProjects()
	if err != nil {
		r.L.Error().Err(err).Msg("Error while querying projects")
		return wrappers.NewError(
			errors.Join(err, fmt.Errorf("Error while querying projects")),
			http.StatusInternalServerError)
	}
	return wrappers.Success(projects)
}

func addProject(r *wrappers.Request) wrappers.IResponse {
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.R.Body).Decode(&body); err != nil {
		return wrappers.NewError(
			fmt.Errorf("Invalid request body: %v", err),
			http.StatusBadRequest)
	}

	if len(body.Name) < 5 {
		return wrappers.NewError(
			fmt.Errorf("Name is required, with a minimum length of 5 characters"),
			http.StatusBadRequest)
	}
	if len(body.Name) > 64 {
		return wrappers.NewError(
			fmt.Errorf("Name exceeds maximum length of 64 characters"),
			http.StatusBadRequest)
	}

	project, err := db.Pool.AddProject(body.Name)
	if err != nil {
		r.L.Error().Err(err).Msg("Error while adding project")
		return wrappers.NewError(
			errors.Join(err, fmt.Errorf("Error while adding project")),
			http.StatusInternalServerError)
	}

	return wrappers.Success(project)
}

/*
   Returns a project by its ID. The ID is the database ID and passed
   in the endpoint route.
 */
func listProjects(r *wrappers.Request) wrappers.IResponse {
}

/*
   Sets all the zsamples associated with a project. Input is a JSON list of
   zsample values. Project ID is in the route. This will be the complete set of
   zsamples, overriding the previous defaults.
 */
func setZSamples(r *wrappers.Request) wrappers.IResponse {
}

/*
   Adds a single zsample to the project. ProjectID and the ZSample are in
   the routing parameters.
 */
func addZSample(r *wrappers.Request) wrappers.IResponse {
}

/*
   Deletes the called zsample on the project. ProjectID and ZSample are
   in the routing parameters.
 */
func deleteZSample(r *wrappers.Request) wrappers.IResponse {
}
