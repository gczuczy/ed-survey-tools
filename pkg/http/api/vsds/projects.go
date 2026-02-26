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
