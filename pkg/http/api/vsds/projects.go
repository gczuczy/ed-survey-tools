package vsds

import (
	"fmt"
	"errors"
	"net/http"

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
}
