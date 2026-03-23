package gcp

import (
	"fmt"
	"regexp"
	"time"
	"errors"
	"context"
	"google.golang.org/api/sheets/v4"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
)

type Sheet interface {
	GetName() string
	Get(row, col int) string
	Rows() int
	Cols() int
}

type GSheet struct {
	name string
	data [][]interface{}
}

func (s *GSheet) GetName() string {
	return s.name
}

func (s *GSheet) Rows() int {
	return len(s.data)
}

func (s *GSheet) Cols() int {
	max := 0
	for _, r := range s.data {
		if len(r) > max {
			max = len(r)
		}
	}
	return max
}

func (s *GSheet) Get(row, col int) string {
	if row >= len(s.data) {
		return ""
	}
	r := s.data[row]
	if col >= len(r) {
		return ""
	}
	v := r[col]
	if v == nil {
		return ""
	}
	if str, ok := v.(string); ok {
		return str
	}
	return fmt.Sprintf("%v", v)
}

type GSpreadsheetsService struct {
	Token         *oauth2.Token
	SheetsService *sheets.Service
}

type GSpreadsheet struct {
	meta          *sheets.Spreadsheet
	SheetsService *sheets.Service
	ID            string
	sheets        []Sheet
}

func NewSheets() (*GSpreadsheetsService, error) {
	var err error
	ctx := context.Background()

	gs := &GSpreadsheetsService{}

	scopes := option.WithScopes(sheets.SpreadsheetsScope)

	if gs.SheetsService, err = sheets.NewService(
		ctx, authOption, scopes); err != nil {
		return nil, err
	}

	return gs, nil
}

// gridDataToValues converts GridData (from IncludeGridData fetch) into
// the same [][]interface{} format produced by Values.Get. The API
// omits trailing empty cells/rows, matching FORMATTED_VALUE behaviour.
func gridDataToValues(gd *sheets.GridData) [][]interface{} {
	rows := make([][]interface{}, 0, len(gd.RowData))
	for _, row := range gd.RowData {
		cells := make([]interface{}, 0, len(row.Values))
		for _, cell := range row.Values {
			cells = append(cells, cell.FormattedValue)
		}
		rows = append(rows, cells)
	}
	return rows
}

func (s *GSpreadsheetsService) Sheet(id string) (*GSpreadsheet, error) {
	var err error
	sheet := &GSpreadsheet{
		ID:            id,
		SheetsService: s.SheetsService,
	}

	// IncludeGridData fetches metadata and all cell values in one call,
	// eliminating per-tab round-trips in GetSheets.
	sheet.meta, err = RateLimit(func() (*sheets.Spreadsheet, error) {
		return s.SheetsService.Spreadsheets.Get(id).
			IncludeGridData(true).Do()
	}, 30*time.Second)
	if err != nil {
		return nil, err
	}

	sheet.sheets = make([]Sheet, 0, len(sheet.meta.Sheets))
	for _, sh := range sheet.meta.Sheets {
		gs := &GSheet{name: sh.Properties.Title}
		if len(sh.Data) > 0 {
			gs.data = gridDataToValues(sh.Data[0])
		}
		sheet.sheets = append(sheet.sheets, gs)
	}

	return sheet, nil
}

func (s *GSpreadsheet) readSheet(
	name string) (*sheets.ValueRange, error) {
	f := func() (*sheets.ValueRange, error) {
		return s.SheetsService.Spreadsheets.Values.
			Get(s.ID, name).Do()
	}
	return RateLimit(f, 30*time.Second)
}

func (s *GSpreadsheet) GetSheets() ([]Sheet, error) {
	// Return pre-loaded data from the Sheet() fetch when available.
	if s.sheets != nil {
		return s.sheets, nil
	}

	// Fallback: fetch each tab individually.
	result := make([]Sheet, 0, len(s.meta.Sheets))
	for _, sh := range s.meta.Sheets {
		name := sh.Properties.Title
		data, err := s.readSheet(name)
		if err != nil {
			return nil, errors.Join(err,
				fmt.Errorf("GetSheets(%s)", name))
		}
		result = append(result, &GSheet{
			name: name,
			data: data.Values,
		})
	}
	return result, nil
}

func (s *GSpreadsheet) ReadRange(
	sheet string, start string, end string,
) (ret *sheets.ValueRange, err error) {
	rangestr := fmt.Sprintf("%s!%s:%s", sheet, start, end)
	f := func() (*sheets.ValueRange, error) {
		return s.SheetsService.Spreadsheets.Values.
			Get(s.ID, rangestr).Do()
	}
	ret, err = RateLimit(f, 30*time.Second)
	if err != nil {
		fmt.Printf("ReadCell error: %T/%v", err, err)
		err = errors.Join(err,
			fmt.Errorf("ReadCell(%s!%s:%s)", sheet, start, end))
	}
	return ret, err
}

var sheetIDRe = regexp.MustCompile(
	`/spreadsheets/d/([A-Za-z0-9_-]+)`)

// ExtractSheetID parses a Google Sheets URL and returns the
// spreadsheet ID, or an error if the URL does not contain one.
func ExtractSheetID(rawURL string) (string, error) {
	m := sheetIDRe.FindStringSubmatch(rawURL)
	if len(m) < 2 {
		return "", fmt.Errorf(
			"cannot extract sheet ID from URL")
	}
	return m[1], nil
}
