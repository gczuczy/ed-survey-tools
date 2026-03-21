package vsds

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/gczuczy/ed-survey-tools/pkg/gcp"
)

// CsvSheet implements gcp.Sheet for CSV files.
// A CSV has no tabs; the sheet name is derived from
// the filename without its extension.
type CsvSheet struct {
	name string
	rows [][]string
}

func (s *CsvSheet) GetName() string {
	return s.name
}

func (s *CsvSheet) Rows() int {
	return len(s.rows)
}

func (s *CsvSheet) Get(row, col int) string {
	if row >= len(s.rows) {
		return ""
	}
	r := s.rows[row]
	if col >= len(r) {
		return ""
	}
	return r[col]
}

func openCsvSheets(
	item gcp.FolderItem,
) ([]gcp.Sheet, error) {
	data, err := gcp.DownloadFile(item.ID)
	if err != nil {
		return nil, err
	}

	// Strip UTF-8 BOM if present so csv.Reader
	// does not treat it as part of the first field.
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))

	r := csv.NewReader(bytes.NewReader(data))
	r.FieldsPerRecord = -1 // allow variable column count
	rows, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf(
			"failed to parse csv %s: %w", item.Name, err)
	}

	// Trim trailing empty rows.
	for len(rows) > 0 {
		last := rows[len(rows)-1]
		empty := true
		for _, v := range last {
			if v != "" {
				empty = false
				break
			}
		}
		if !empty {
			break
		}
		rows = rows[:len(rows)-1]
	}

	// Derive sheet name from filename without extension.
	name := item.Name
	if i := strings.LastIndexByte(name, '.'); i >= 0 {
		name = name[:i]
	}

	return []gcp.Sheet{&CsvSheet{
		name: name,
		rows: rows,
	}}, nil
}
