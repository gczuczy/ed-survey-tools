package bundles

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gczuczy/ed-survey-tools/pkg/db"
	w "github.com/gczuczy/ed-survey-tools/pkg/http/wrappers"
)

// BundleListResponse is the response type for GET /api/bundles
type BundleListResponse = []db.Bundle

func listBundles(r *w.Request) w.IResponse {
	bundles, err := db.Pool.ListBundles()
	if err != nil {
		r.L.Error().Err(err).Msg("Error listing bundles")
		return w.NewError(
			errors.Join(err,
				fmt.Errorf("Error listing bundles")),
			http.StatusInternalServerError)
	}
	return w.Success(bundles)
}
