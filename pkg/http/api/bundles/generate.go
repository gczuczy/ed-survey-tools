package bundles

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gczuczy/ed-survey-tools/pkg/bundles"
	"github.com/gczuczy/ed-survey-tools/pkg/db"
	w "github.com/gczuczy/ed-survey-tools/pkg/http/wrappers"
)

func generateBundle(r *w.Request) w.IResponse {
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

	if err := db.Pool.QueueBundle(id); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return w.NewError(
				fmt.Errorf("Bundle not found"),
				http.StatusNotFound)
		}
		if errors.Is(err, db.ErrAlreadyQueued) {
			return w.NewError(
				fmt.Errorf(
					"Bundle is currently generating"),
				http.StatusConflict)
		}
		r.L.Error().Err(err).
			Msg("Error queuing bundle")
		return w.NewError(
			errors.Join(err,
				fmt.Errorf("Error queuing bundle")),
			http.StatusInternalServerError)
	}

	bundles.Signal()
	return w.Success(nil)
}
