package vsds

import (
	"fmt"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"encoding/json"

	"github.com/gczuczy/ed-survey-tools/pkg/http/wrappers"
	"github.com/gczuczy/ed-survey-tools/pkg/db"
)

func listVariants(r *wrappers.Request) wrappers.IResponse {
	idStr, ok := r.Vars["id"]
	if !ok {
		return wrappers.NewError(
			fmt.Errorf("Missing project ID"),
			http.StatusBadRequest)
	}
	projectID, err := strconv.Atoi(idStr)
	if err != nil || projectID <= 0 {
		return wrappers.NewError(
			fmt.Errorf("Invalid project ID"),
			http.StatusBadRequest)
	}

	variants, err := db.Pool.ListVariants(projectID)
	if err != nil {
		r.L.Error().Err(err).Msg("Error while listing variants")
		return wrappers.NewError(
			errors.Join(
				err,
				fmt.Errorf("Error while listing variants"),
			),
			http.StatusInternalServerError)
	}
	return wrappers.Success(variants)
}

func addVariant(r *wrappers.Request) wrappers.IResponse {
	idStr, ok := r.Vars["id"]
	if !ok {
		return wrappers.NewError(
			fmt.Errorf("Missing project ID"),
			http.StatusBadRequest)
	}
	projectID, err := strconv.Atoi(idStr)
	if err != nil || projectID <= 0 {
		return wrappers.NewError(
			fmt.Errorf("Invalid project ID"),
			http.StatusBadRequest)
	}

	var body struct {
		Name              string `json:"name"`
		HeaderRow         int    `json:"header_row"`
		SysNameColumn     int    `json:"sysname_column"`
		ZSampleColumn     int    `json:"zsample_column"`
		SystemCountColumn int    `json:"syscount_column"`
		MaxDistanceColumn int    `json:"maxdistance_column"`
	}
	if err := json.NewDecoder(r.R.Body).Decode(&body); err != nil {
		return wrappers.NewError(
			fmt.Errorf("Invalid request body: %v", err),
			http.StatusBadRequest)
	}

	name := strings.TrimSpace(body.Name)
	if len(name) < 1 || len(name) > 64 {
		return wrappers.NewError(
			fmt.Errorf("Name must be 1-64 characters"),
			http.StatusBadRequest)
	}
	if body.HeaderRow < 0 {
		return wrappers.NewError(
			fmt.Errorf("header_row must be >= 0"),
			http.StatusBadRequest)
	}
	if body.SysNameColumn < 0 || body.ZSampleColumn < 0 ||
		body.SystemCountColumn < 0 || body.MaxDistanceColumn < 0 {
		return wrappers.NewError(
			fmt.Errorf("Column values must be >= 0"),
			http.StatusBadRequest)
	}

	sv := db.DBSheetVariant{
		ProjectID:         projectID,
		Name:              name,
		HeaderRow:         body.HeaderRow,
		SysNameColumn:     body.SysNameColumn,
		ZSampleColumn:     body.ZSampleColumn,
		SystemCountColumn: body.SystemCountColumn,
		MaxDistanceColumn: body.MaxDistanceColumn,
	}
	result, err := db.Pool.AddVariant(sv)
	if err != nil {
		r.L.Error().Err(err).Msg("Error while adding variant")
		return wrappers.NewError(
			errors.Join(
				err,
				fmt.Errorf("Error while adding variant"),
			),
			http.StatusInternalServerError)
	}
	return wrappers.Success(result)
}

func updateVariant(r *wrappers.Request) wrappers.IResponse {
	idStr, ok := r.Vars["id"]
	if !ok {
		return wrappers.NewError(
			fmt.Errorf("Missing project ID"),
			http.StatusBadRequest)
	}
	projectID, err := strconv.Atoi(idStr)
	if err != nil || projectID <= 0 {
		return wrappers.NewError(
			fmt.Errorf("Invalid project ID"),
			http.StatusBadRequest)
	}

	vidStr, ok := r.Vars["vid"]
	if !ok {
		return wrappers.NewError(
			fmt.Errorf("Missing variant ID"),
			http.StatusBadRequest)
	}
	variantID, err := strconv.Atoi(vidStr)
	if err != nil || variantID <= 0 {
		return wrappers.NewError(
			fmt.Errorf("Invalid variant ID"),
			http.StatusBadRequest)
	}

	var body struct {
		Name              string `json:"name"`
		HeaderRow         int    `json:"header_row"`
		SysNameColumn     int    `json:"sysname_column"`
		ZSampleColumn     int    `json:"zsample_column"`
		SystemCountColumn int    `json:"syscount_column"`
		MaxDistanceColumn int    `json:"maxdistance_column"`
	}
	if err := json.NewDecoder(r.R.Body).Decode(&body); err != nil {
		return wrappers.NewError(
			fmt.Errorf("Invalid request body: %v", err),
			http.StatusBadRequest)
	}

	name := strings.TrimSpace(body.Name)
	if len(name) < 1 || len(name) > 64 {
		return wrappers.NewError(
			fmt.Errorf("Name must be 1-64 characters"),
			http.StatusBadRequest)
	}
	if body.HeaderRow < 0 {
		return wrappers.NewError(
			fmt.Errorf("header_row must be >= 0"),
			http.StatusBadRequest)
	}
	if body.SysNameColumn < 0 || body.ZSampleColumn < 0 ||
		body.SystemCountColumn < 0 || body.MaxDistanceColumn < 0 {
		return wrappers.NewError(
			fmt.Errorf("Column values must be >= 0"),
			http.StatusBadRequest)
	}

	sv := db.DBSheetVariant{
		ID:                variantID,
		ProjectID:         projectID,
		Name:              name,
		HeaderRow:         body.HeaderRow,
		SysNameColumn:     body.SysNameColumn,
		ZSampleColumn:     body.ZSampleColumn,
		SystemCountColumn: body.SystemCountColumn,
		MaxDistanceColumn: body.MaxDistanceColumn,
	}
	result, err := db.Pool.UpdateVariant(sv)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return wrappers.NewError(
				fmt.Errorf("Variant not found"),
				http.StatusNotFound)
		}
		r.L.Error().Err(err).Msg("Error while updating variant")
		return wrappers.NewError(
			errors.Join(
				err,
				fmt.Errorf("Error while updating variant"),
			),
			http.StatusInternalServerError)
	}
	return wrappers.Success(result)
}

func deleteVariant(r *wrappers.Request) wrappers.IResponse {
	idStr, ok := r.Vars["id"]
	if !ok {
		return wrappers.NewError(
			fmt.Errorf("Missing project ID"),
			http.StatusBadRequest)
	}
	projectID, err := strconv.Atoi(idStr)
	if err != nil || projectID <= 0 {
		return wrappers.NewError(
			fmt.Errorf("Invalid project ID"),
			http.StatusBadRequest)
	}

	vidStr, ok := r.Vars["vid"]
	if !ok {
		return wrappers.NewError(
			fmt.Errorf("Missing variant ID"),
			http.StatusBadRequest)
	}
	variantID, err := strconv.Atoi(vidStr)
	if err != nil || variantID <= 0 {
		return wrappers.NewError(
			fmt.Errorf("Invalid variant ID"),
			http.StatusBadRequest)
	}

	err = db.Pool.DeleteVariant(projectID, variantID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return wrappers.NewError(
				fmt.Errorf("Variant not found"),
				http.StatusNotFound)
		}
		r.L.Error().Err(err).Msg("Error while deleting variant")
		return wrappers.NewError(
			errors.Join(
				err,
				fmt.Errorf("Error while deleting variant"),
			),
			http.StatusInternalServerError)
	}
	return wrappers.Success(nil)
}

func addVariantCheck(r *wrappers.Request) wrappers.IResponse {
	idStr, ok := r.Vars["id"]
	if !ok {
		return wrappers.NewError(
			fmt.Errorf("Missing project ID"),
			http.StatusBadRequest)
	}
	projectID, err := strconv.Atoi(idStr)
	if err != nil || projectID <= 0 {
		return wrappers.NewError(
			fmt.Errorf("Invalid project ID"),
			http.StatusBadRequest)
	}

	vidStr, ok := r.Vars["vid"]
	if !ok {
		return wrappers.NewError(
			fmt.Errorf("Missing variant ID"),
			http.StatusBadRequest)
	}
	variantID, err := strconv.Atoi(vidStr)
	if err != nil || variantID <= 0 {
		return wrappers.NewError(
			fmt.Errorf("Invalid variant ID"),
			http.StatusBadRequest)
	}

	var body struct {
		Col   int    `json:"col"`
		Row   int    `json:"row"`
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.R.Body).Decode(&body); err != nil {
		return wrappers.NewError(
			fmt.Errorf("Invalid request body: %v", err),
			http.StatusBadRequest)
	}

	if body.Col < 0 || body.Row < 0 {
		return wrappers.NewError(
			fmt.Errorf("col and row must be >= 0"),
			http.StatusBadRequest)
	}
	value := strings.TrimSpace(body.Value)
	if len(value) < 1 || len(value) > 64 {
		return wrappers.NewError(
			fmt.Errorf("value must be 1-64 characters"),
			http.StatusBadRequest)
	}

	check := db.DBSheetVariantCheck{
		Col:   body.Col,
		Row:   body.Row,
		Value: value,
	}
	result, err := db.Pool.AddVariantCheck(projectID, variantID, check)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return wrappers.NewError(
				fmt.Errorf("Variant not found"),
				http.StatusNotFound)
		}
		if errors.Is(err, db.ErrDuplicate) {
			return wrappers.NewError(
				fmt.Errorf(
					"A check for this cell already exists",
				),
				http.StatusConflict)
		}
		r.L.Error().Err(err).
			Msg("Error while adding variant check")
		return wrappers.NewError(
			errors.Join(
				err,
				fmt.Errorf("Error while adding variant check"),
			),
			http.StatusInternalServerError)
	}
	return wrappers.Success(result)
}

func deleteVariantCheck(r *wrappers.Request) wrappers.IResponse {
	idStr, ok := r.Vars["id"]
	if !ok {
		return wrappers.NewError(
			fmt.Errorf("Missing project ID"),
			http.StatusBadRequest)
	}
	projectID, err := strconv.Atoi(idStr)
	if err != nil || projectID <= 0 {
		return wrappers.NewError(
			fmt.Errorf("Invalid project ID"),
			http.StatusBadRequest)
	}

	vidStr, ok := r.Vars["vid"]
	if !ok {
		return wrappers.NewError(
			fmt.Errorf("Missing variant ID"),
			http.StatusBadRequest)
	}
	variantID, err := strconv.Atoi(vidStr)
	if err != nil || variantID <= 0 {
		return wrappers.NewError(
			fmt.Errorf("Invalid variant ID"),
			http.StatusBadRequest)
	}

	cidStr, ok := r.Vars["cid"]
	if !ok {
		return wrappers.NewError(
			fmt.Errorf("Missing check ID"),
			http.StatusBadRequest)
	}
	checkID, err := strconv.Atoi(cidStr)
	if err != nil || checkID <= 0 {
		return wrappers.NewError(
			fmt.Errorf("Invalid check ID"),
			http.StatusBadRequest)
	}

	result, err := db.Pool.DeleteVariantCheck(
		projectID, variantID, checkID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return wrappers.NewError(
				fmt.Errorf("Check not found"),
				http.StatusNotFound)
		}
		r.L.Error().Err(err).
			Msg("Error while deleting variant check")
		return wrappers.NewError(
			errors.Join(
				err,
				fmt.Errorf(
					"Error while deleting variant check",
				),
			),
			http.StatusInternalServerError)
	}
	return wrappers.Success(result)
}
