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

type GSpreadsheetsService struct {
	Token *oauth2.Token
	SheetsService *sheets.Service
}

type GSpreadsheet struct {
	Sheet *sheets.Spreadsheet
	SheetsService *sheets.Service
	ID string
}

func NewSheets(credfile string) (*GSpreadsheetsService, error) {
	var err error
	ctx := context.Background()

	gs := &GSpreadsheetsService{}

	creds := option.WithCredentialsFile(credfile)
	scopes := option.WithScopes(sheets.SpreadsheetsScope)

	if gs.SheetsService, err = sheets.NewService(ctx, creds, scopes); err != nil {
		return nil, err
	}

	return gs, nil
}

func (s *GSpreadsheetsService) Sheet(id string) (*GSpreadsheet, error) {
	var err error
	sheet := &GSpreadsheet{
		ID: id,
		SheetsService: s.SheetsService,
	}

	sheet.Sheet, err = s.SheetsService.Spreadsheets.Get(id).Do()
	if err != nil {
		return nil, err
	}

	return sheet, nil
}

func (s *GSpreadsheet) GetSheets() []*sheets.Sheet {
	return s.Sheet.Sheets
}

func (s *GSpreadsheet) ReadRange(sheet string, start string, end string) (ret *sheets.ValueRange, err error) {
	rangestr := fmt.Sprintf("%s!%s:%s", sheet, start, end)
	f := func() (*sheets.ValueRange, error) {
		return s.SheetsService.Spreadsheets.Values.Get(s.ID, rangestr).Do()
	}
	ret, err = RateLimit(f, 30*time.Second)
	if err != nil {
		fmt.Printf("ReadCell error: %T/%v", err, err)
		err = errors.Join(err, fmt.Errorf("ReadCell(%s!%s:%s)", sheet, start, end))
	}
	return ret, err
}
