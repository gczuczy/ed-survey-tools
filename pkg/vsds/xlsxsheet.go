package vsds

import (
	"bytes"
	"fmt"

	"github.com/gczuczy/ed-survey-tools/pkg/gcp"
	"github.com/xuri/excelize/v2"
)

type XlsxSheet struct {
	name string
	rows [][]string
}

func (s *XlsxSheet) GetName() string {
	return s.name
}

func (s *XlsxSheet) Rows() int {
	return len(s.rows)
}

func (s *XlsxSheet) Cols() int {
	max := 0
	for _, r := range s.rows {
		if len(r) > max {
			max = len(r)
		}
	}
	return max
}

func (s *XlsxSheet) Get(row, col int) string {
	if row >= len(s.rows) {
		return ""
	}
	r := s.rows[row]
	if col >= len(r) {
		return ""
	}
	return r[col]
}

func openXlsxSheets(
	item gcp.FolderItem,
) ([]gcp.Sheet, error) {
	data, err := gcp.DownloadFile(item.ID)
	if err != nil {
		return nil, err
	}

	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf(
			"failed to parse xlsx %s: %w", item.Name, err)
	}
	defer f.Close()

	names := f.GetSheetList()
	sheets := make([]gcp.Sheet, 0, len(names))
	for _, name := range names {
		rows, rErr := f.GetRows(name)
		if rErr != nil {
			return nil, fmt.Errorf(
				"failed to read sheet %s in %s: %w",
				name, item.Name, rErr)
		}
		sheets = append(sheets, &XlsxSheet{
			name: name,
			rows: rows,
		})
	}
	return sheets, nil
}
