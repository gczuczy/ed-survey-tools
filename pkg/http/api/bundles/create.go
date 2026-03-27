package bundles

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gczuczy/ed-survey-tools/pkg/bundles"
	"github.com/gczuczy/ed-survey-tools/pkg/db"
	w "github.com/gczuczy/ed-survey-tools/pkg/http/wrappers"
)

type createBundleBody struct {
	MeasurementType string          `json:"measurementtype"`
	Name            string          `json:"name"`
	AutoRegen       bool            `json:"autoregen"`
	VSDS            *vsdsCreateBody `json:"vsds"`
}

type vsdsCreateBody struct {
	Subtype     string `json:"subtype"`
	AllProjects bool   `json:"allprojects"`
	Projects    []int  `json:"projects"`
}

func createBundle(r *w.Request) w.IResponse {
	var body createBundleBody
	if err := json.NewDecoder(r.R.Body).Decode(&body); err != nil {
		return w.NewError(
			fmt.Errorf("Invalid request body: %v", err),
			http.StatusBadRequest)
	}

	if body.MeasurementType == "" {
		return w.NewError(
			fmt.Errorf("measurementtype is required"),
			http.StatusBadRequest)
	}
	if body.MeasurementType != "vsds" {
		return w.NewError(
			fmt.Errorf(
				"unsupported measurementtype: %s",
				body.MeasurementType),
			http.StatusBadRequest)
	}
	if body.Name == "" {
		return w.NewError(
			fmt.Errorf("name is required"),
			http.StatusBadRequest)
	}
	if body.VSDS == nil {
		return w.NewError(
			fmt.Errorf("vsds config is required"),
			http.StatusBadRequest)
	}
	if body.VSDS.Subtype != "surveypoints" &&
		body.VSDS.Subtype != "surveys" {
		return w.NewError(
			fmt.Errorf(
				"vsds.subtype must be surveypoints or surveys"),
			http.StatusBadRequest)
	}
	if !body.VSDS.AllProjects &&
		len(body.VSDS.Projects) == 0 {
		return w.NewError(
			fmt.Errorf(
				"vsds.projects required when allprojects is false"),
			http.StatusBadRequest)
	}

	bundle, err := db.Pool.CreateVSDSBundle(
		body.Name,
		body.AutoRegen,
		body.VSDS.Subtype,
		body.VSDS.AllProjects,
		body.VSDS.Projects,
	)
	if err != nil {
		r.L.Error().Err(err).Msg("Error creating bundle")
		return w.NewError(
			errors.Join(err,
				fmt.Errorf("Error creating bundle")),
			http.StatusInternalServerError)
	}
	bundles.Signal()
	return w.Success(bundle)
}
