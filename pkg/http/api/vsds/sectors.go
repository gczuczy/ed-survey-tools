package vsds

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gczuczy/ed-survey-tools/pkg/db"
	w "github.com/gczuczy/ed-survey-tools/pkg/http/wrappers"
)

// SectorVoxel is the response element for
// GET /api/vsds/visualization/sectors.
type SectorVoxel struct {
	GCX    float64 `db:"gc_x"    json:"gc_x"`
	GCZ    float64 `db:"gc_z"    json:"gc_z"`
	YMin   float64 `db:"y_min"   json:"y_min"`
	YMax   float64 `db:"y_max"   json:"y_max"`
	RhoMin float64 `db:"rho_min" json:"rho_min"`
	RhoAvg float64 `db:"rho_avg" json:"rho_avg"`
	RhoMax float64 `db:"rho_max" json:"rho_max"`
	Count  int64   `db:"cnt"     json:"count"`
}

// SectorsResponse is the response type for
// GET /api/vsds/visualization/sectors.
type SectorsResponse = []SectorVoxel

func listSectors(r *w.Request) w.IResponse {
	q := r.R.URL.Query()

	xzStep, err := parseStepParam(q.Get("xz_step"), 50, 10, 5000)
	if err != nil {
		return w.NewError(
			fmt.Errorf("invalid xz_step: %v", err),
			http.StatusBadRequest)
	}

	yStep, err := parseStepParam(q.Get("y_step"), 20, 10, 50000)
	if err != nil {
		return w.NewError(
			fmt.Errorf("invalid y_step: %v", err),
			http.StatusBadRequest)
	}

	rows, err := db.QueryRows[SectorVoxel](
		db.Pool, "vsds_sectors",
		float64(xzStep), float64(yStep),
	)
	if err != nil {
		r.L.Error().Err(err).
			Int("xz_step", xzStep).Int("y_step", yStep).
			Msg("Error fetching sectors")
		return w.NewError(
			errors.Join(err,
				fmt.Errorf("Error fetching sectors")),
			http.StatusInternalServerError)
	}
	return w.Success(rows)
}

// parseStepParam parses an optional integer query parameter.
// If absent, defaultVal is used. Returns an error if out of
// [min, max].
func parseStepParam(
	raw string, defaultVal, min, max int,
) (int, error) {
	if raw == "" {
		return defaultVal, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}
	if v < min || v > max {
		return 0, fmt.Errorf(
			"must be between %d and %d", min, max)
	}
	return v, nil
}
