package gcp

import (
	"fmt"
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

func (s *GSpreadsheetsService) Sheet(id string) (*GSpreadsheet, error) {
	var err error
	sheet := &GSpreadsheet{
		ID:            id,
		SheetsService: s.SheetsService,
	}

	sheet.meta, err = RateLimit(func() (*sheets.Spreadsheet, error) {
		return s.SheetsService.Spreadsheets.Get(id).Do()
	}, 30*time.Second)
	if err != nil {
		return nil, err
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
