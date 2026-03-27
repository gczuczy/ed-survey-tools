package vsds

import (
	"encoding/json"
	"fmt"

	"github.com/gczuczy/ed-survey-tools/pkg/db"
)

// bundleSurveyPoint is the JSON output type for survey-point bundles.
// Internal IDs are excluded; corrected_n is exposed as "syscount".
type bundleSurveyPoint struct {
	SysName     string  `db:"sysname"     json:"sysname"`
	ZSample     int     `db:"zsample"     json:"zsample"`
	X           float32 `db:"x"           json:"x"`
	Y           float32 `db:"y"           json:"y"`
	Z           float32 `db:"z"           json:"z"`
	GCX         float32 `db:"gc_x"        json:"gc_x"`
	GCY         float32 `db:"gc_y"        json:"gc_y"`
	GCZ         float32 `db:"gc_z"        json:"gc_z"`
	SysCount    int     `db:"corrected_n" json:"syscount"`
	MaxDistance float32 `db:"maxdistance" json:"maxdistance"`
	Rho         float64 `db:"rho"         json:"rho"`
}

// bundleSurvey is the JSON output type for survey bundles.
// Internal IDs are excluded; points is preserved as raw JSON.
type bundleSurvey struct {
	ProjectName string          `db:"projectname" json:"projectname"`
	RhoMax      float64         `db:"rho_max"     json:"rho_max"`
	X           float64         `db:"x"           json:"x"`
	Z           float64         `db:"z"           json:"z"`
	RhoStddev   *float64        `db:"rho_stddev"  json:"rho_stddev"`
	GCX         float64         `db:"gc_x"        json:"gc_x"`
	GCZ         float64         `db:"gc_z"        json:"gc_z"`
	Points      json.RawMessage `db:"points"      json:"points"`
}

// VSDSBundleRunner implements bundles.BundleRunner for VSDS data.
type VSDSBundleRunner struct{}

// NewBundleRunner returns a new VSDSBundleRunner.
func NewBundleRunner() *VSDSBundleRunner {
	return &VSDSBundleRunner{}
}

// LoadConfig fetches the VSDS-specific configuration for bundleID.
func (r *VSDSBundleRunner) LoadConfig(
	bundleID int,
) (any, error) {
	return db.Pool.GetVSDBundleConfig(bundleID)
}

// Generate builds the bundle payload from the database.
// config must be a *db.VSDBundleConfig returned by LoadConfig.
func (r *VSDSBundleRunner) Generate(config any) (any, error) {
	cfg, ok := config.(*db.VSDBundleConfig)
	if !ok {
		return nil, fmt.Errorf(
			"unexpected config type %T", config)
	}
	switch cfg.Subtype {
	case "surveypoints":
		return r.generateSurveyPoints(cfg)
	case "surveys":
		return r.generateSurveys(cfg)
	default:
		return nil, fmt.Errorf(
			"unknown vsds subtype: %s", cfg.Subtype)
	}
}

func (r *VSDSBundleRunner) generateSurveyPoints(
	cfg *db.VSDBundleConfig,
) (any, error) {
	if cfg.AllProjects {
		return db.QueryRows[bundleSurveyPoint](
			db.Pool, "vsds_bundle_surveypts_all",
		)
	}
	pids := toInt32Slice(cfg.ProjectIDs)
	return db.QueryRows[bundleSurveyPoint](
		db.Pool, "vsds_bundle_surveypts_proj", pids,
	)
}

func (r *VSDSBundleRunner) generateSurveys(
	cfg *db.VSDBundleConfig,
) (any, error) {
	if cfg.AllProjects {
		return db.QueryRows[bundleSurvey](
			db.Pool, "vsds_bundle_surveys_all",
		)
	}
	pids := toInt32Slice(cfg.ProjectIDs)
	return db.QueryRows[bundleSurvey](
		db.Pool, "vsds_bundle_surveys_proj", pids,
	)
}

func toInt32Slice(ids []int) []int32 {
	result := make([]int32, len(ids))
	for i, id := range ids {
		result[i] = int32(id)
	}
	return result
}
