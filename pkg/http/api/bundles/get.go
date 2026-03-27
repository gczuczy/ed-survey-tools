package bundles

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gczuczy/ed-survey-tools/pkg/db"
	w "github.com/gczuczy/ed-survey-tools/pkg/http/wrappers"
)

// BundleResponse is the response type for GET /api/bundles/{id}
type BundleResponse = db.Bundle

func getBundle(r *w.Request) w.IResponse {
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

	bundle, err := db.Pool.GetBundle(id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return w.NewError(
				fmt.Errorf("Bundle not found"),
				http.StatusNotFound)
		}
		r.L.Error().Err(err).Msg("Error fetching bundle")
		return w.NewError(
			errors.Join(err,
				fmt.Errorf("Error fetching bundle")),
			http.StatusInternalServerError)
	}
	return w.Success(bundle)
}
