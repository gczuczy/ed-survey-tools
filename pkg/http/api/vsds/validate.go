package vsds

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gczuczy/ed-survey-tools/pkg/gcp"
	pvsds "github.com/gczuczy/ed-survey-tools/pkg/vsds"
	"github.com/gczuczy/ed-survey-tools/pkg/http/wrappers"
)

// VariantValidateCheckResult is the evaluation outcome for one header
// check cell.
type VariantValidateCheckResult struct {
	Col      int    `json:"col"`
	Row      int    `json:"row"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
	OK       bool   `json:"ok"`
}

// VariantValidateTabResult holds the serialised cell data and check
// results for one spreadsheet tab.
type VariantValidateTabResult struct {
	Name    string                       `json:"name"`
	Rows    [][]string                   `json:"rows"`
	Checks  []VariantValidateCheckResult `json:"checks"`
	Matched bool                         `json:"matched"`
}

// VariantValidateResponse is the top-level payload returned by the
// variant validation endpoint.
type VariantValidateResponse struct {
	Tabs []VariantValidateTabResult `json:"tabs"`
}

// tabToRows serialises sheet into a [][]string grid capped at
// 50 rows and 15 columns.
func tabToRows(sheet gcp.Sheet) [][]string {
	maxRows := sheet.Rows()
	if maxRows > 50 {
		maxRows = 50
	}
	maxCols := sheet.Cols()
	if maxCols > 15 {
		maxCols = 15
	}

	rows := make([][]string, maxRows)
	for r := 0; r < maxRows; r++ {
		row := make([]string, maxCols)
		for c := 0; c < maxCols; c++ {
			row[c] = sheet.Get(r, c)
		}
		rows[r] = row
	}
	return rows
}

func validateVariant(r *wrappers.Request) wrappers.IResponse {
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
		URL     string `json:"url"`
		Variant struct {
			Name              string `json:"name"`
			HeaderRow         int    `json:"header_row"`
			SysNameColumn     int    `json:"sysname_column"`
			ZSampleColumn     int    `json:"zsample_column"`
			SystemCountColumn int    `json:"syscount_column"`
			MaxDistanceColumn int    `json:"maxdistance_column"`
			Checks            []struct {
				Col   int    `json:"col"`
				Row   int    `json:"row"`
				Value string `json:"value"`
			} `json:"checks"`
		} `json:"variant"`
	}
	if err := json.NewDecoder(r.R.Body).Decode(&body); err != nil {
		return wrappers.NewError(
			fmt.Errorf("Invalid request body: %v", err),
			http.StatusBadRequest)
	}

	url := strings.TrimSpace(body.URL)
	if url == "" {
		return wrappers.NewError(
			fmt.Errorf("url is required"),
			http.StatusBadRequest)
	}
	if body.Variant.HeaderRow < 0 {
		return wrappers.NewError(
			fmt.Errorf("header_row must be >= 0"),
			http.StatusBadRequest)
	}
	if body.Variant.SysNameColumn < 0 ||
		body.Variant.ZSampleColumn < 0 ||
		body.Variant.SystemCountColumn < 0 ||
		body.Variant.MaxDistanceColumn < 0 {
		return wrappers.NewError(
			fmt.Errorf("Column values must be >= 0"),
			http.StatusBadRequest)
	}

	sheetID, err := gcp.ExtractSheetID(url)
	if err != nil {
		return wrappers.NewError(
			fmt.Errorf("Invalid Google Sheets URL"),
			http.StatusBadRequest)
	}

	gss, err := gcp.NewSheets()
	if err != nil {
		r.L.Error().Err(err).
			Msg("Error initialising Sheets service")
		return wrappers.NewError(
			fmt.Errorf("Google Sheets service unavailable"),
			http.StatusInternalServerError)
	}

	spreadsheet, err := gss.Sheet(sheetID)
	if err != nil {
		r.L.Error().Err(err).
			Str("sheetid", sheetID).
			Msg("Error fetching spreadsheet")
		return wrappers.NewError(
			fmt.Errorf("Could not fetch spreadsheet: %v", err),
			http.StatusUnprocessableEntity)
	}

	sheets, err := spreadsheet.GetSheets()
	if err != nil {
		r.L.Error().Err(err).
			Str("sheetid", sheetID).
			Msg("Error reading sheets")
		return wrappers.NewError(
			fmt.Errorf("Could not read spreadsheet tabs"),
			http.StatusInternalServerError)
	}

	// Build CheckInput slice once — reused for every tab.
	checks := make([]pvsds.CheckInput, len(body.Variant.Checks))
	for i, c := range body.Variant.Checks {
		checks[i] = pvsds.CheckInput{
			Col:   c.Col,
			Row:   c.Row,
			Value: c.Value,
		}
	}

	tabs := make([]VariantValidateTabResult, 0, len(sheets))
	for _, sheet := range sheets {
		checkResults := pvsds.EvalChecks(checks, sheet)

		matched := true
		for _, cr := range checkResults {
			if !cr.OK {
				matched = false
				break
			}
		}

		// Convert check results to response type.
		respChecks := make(
			[]VariantValidateCheckResult, len(checkResults))
		for i, cr := range checkResults {
			respChecks[i] = VariantValidateCheckResult{
				Col:      cr.Col,
				Row:      cr.Row,
				Expected: cr.Expected,
				Actual:   cr.Actual,
				OK:       cr.OK,
			}
		}

		tabs = append(tabs, VariantValidateTabResult{
			Name:    sheet.GetName(),
			Rows:    tabToRows(sheet),
			Checks:  respChecks,
			Matched: matched,
		})
	}

	return wrappers.Success(VariantValidateResponse{Tabs: tabs})
}
