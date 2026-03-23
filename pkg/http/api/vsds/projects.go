package vsds

import (
	"fmt"
	"errors"
	"net/http"
	"strconv"
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
func getProject(r *wrappers.Request) wrappers.IResponse {
	idStr, ok := r.Vars["id"]
	if !ok {
		return wrappers.NewError(fmt.Errorf("Missing project ID"), http.StatusBadRequest)
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return wrappers.NewError(fmt.Errorf("Invalid project ID"), http.StatusBadRequest)
	}

	project, err := db.Pool.GetProject(id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return wrappers.NewError(fmt.Errorf("Project not found"), http.StatusNotFound)
		}
		r.L.Error().Err(err).Msg("Error while querying project")
		return wrappers.NewError(
			errors.Join(err, fmt.Errorf("Error while querying project")),
			http.StatusInternalServerError)
	}
	return wrappers.Success(project)
}

/*
   Sets all the zsamples associated with a project. Input is a JSON list of
   zsample values. Project ID is in the route. This will be the complete set of
   zsamples, overriding the previous defaults.
 */
func setZSamples(r *wrappers.Request) wrappers.IResponse {
	idStr, ok := r.Vars["id"]
	if !ok {
		return wrappers.NewError(fmt.Errorf("Missing project ID"), http.StatusBadRequest)
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return wrappers.NewError(fmt.Errorf("Invalid project ID"), http.StatusBadRequest)
	}

	var zsamples []int
	if err := json.NewDecoder(r.R.Body).Decode(&zsamples); err != nil {
		return wrappers.NewError(
			fmt.Errorf("Invalid request body: %v", err),
			http.StatusBadRequest)
	}

	project, err := db.Pool.SetZSamples(id, zsamples)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return wrappers.NewError(fmt.Errorf("Project not found"), http.StatusNotFound)
		}
		r.L.Error().Err(err).Msg("Error while setting zsamples")
		return wrappers.NewError(
			errors.Join(err, fmt.Errorf("Error while setting zsamples")),
			http.StatusInternalServerError)
	}
	return wrappers.Success(project)
}

/*
   Adds a single zsample to the project. ProjectID and the ZSample are in
   the routing parameters.
 */
func addZSample(r *wrappers.Request) wrappers.IResponse {
	idStr, ok := r.Vars["id"]
	if !ok {
		return wrappers.NewError(fmt.Errorf("Missing project ID"), http.StatusBadRequest)
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return wrappers.NewError(fmt.Errorf("Invalid project ID"), http.StatusBadRequest)
	}

	zsampleStr, ok := r.Vars["zsample"]
	if !ok {
		return wrappers.NewError(fmt.Errorf("Missing zsample"), http.StatusBadRequest)
	}
	zsample, err := strconv.Atoi(zsampleStr)
	if err != nil {
		return wrappers.NewError(fmt.Errorf("Invalid zsample value"), http.StatusBadRequest)
	}

	project, err := db.Pool.AddZSample(id, zsample)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return wrappers.NewError(fmt.Errorf("Project not found"), http.StatusNotFound)
		}
		if errors.Is(err, db.ErrDuplicate) {
			return wrappers.NewError(fmt.Errorf("ZSample already exists on this project"), http.StatusConflict)
		}
		r.L.Error().Err(err).Msg("Error while adding zsample")
		return wrappers.NewError(
			errors.Join(err, fmt.Errorf("Error while adding zsample")),
			http.StatusInternalServerError)
	}
	return wrappers.Success(project)
}

/*
   Deletes the called zsample on the project. ProjectID and ZSample are
   in the routing parameters.
 */
func deleteZSample(r *wrappers.Request) wrappers.IResponse {
	idStr, ok := r.Vars["id"]
	if !ok {
		return wrappers.NewError(fmt.Errorf("Missing project ID"), http.StatusBadRequest)
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return wrappers.NewError(fmt.Errorf("Invalid project ID"), http.StatusBadRequest)
	}

	zsampleStr, ok := r.Vars["zsample"]
	if !ok {
		return wrappers.NewError(fmt.Errorf("Missing zsample"), http.StatusBadRequest)
	}
	zsample, err := strconv.Atoi(zsampleStr)
	if err != nil {
		return wrappers.NewError(fmt.Errorf("Invalid zsample value"), http.StatusBadRequest)
	}

	project, err := db.Pool.DeleteZSample(id, zsample)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return wrappers.NewError(fmt.Errorf("Project or zsample not found"), http.StatusNotFound)
		}
		r.L.Error().Err(err).Msg("Error while deleting zsample")
		return wrappers.NewError(
			errors.Join(err, fmt.Errorf("Error while deleting zsample")),
			http.StatusInternalServerError)
	}
	return wrappers.Success(project)
}
