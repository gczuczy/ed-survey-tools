package densitysurvey

import (
	"fmt"
	"errors"
	"strings"
	"net/url"
	"regexp"
	"github.com/gczuczy/ed-survey-tools/pkg/google"
)

type EntrySheet struct {
	spreadsheet *google.GSpreadsheet
}

func NewEntrySheet(sheetid string, ss *google.GSpreadsheetsService) (*EntrySheet, error) {
	s, err := ss.Sheet(sheetid)
	if err != nil {
		fmt.Printf("NewEntrySheet err: %v\n", err)
		return nil, errors.Join(err, fmt.Errorf("Unable to load sheet %s", sheetid))
	}

	return &EntrySheet{
		spreadsheet: s,
	}, nil
}

func (es *EntrySheet) GetSheetIDs() ([]string, error) {
	var reterr error = nil
	sheetids := []string{}

	sheetname := es.spreadsheet.GetSheets()[0].Properties.Title

	step := 1024
	pos := 1
	cont := true

	for {
		startcell := fmt.Sprintf("A%d", pos)
		endcell := fmt.Sprintf("A%d", pos+step)
		data, err := es.spreadsheet.ReadRange(sheetname, startcell, endcell)
		if err != nil {
			return []string{}, errors.Join(err, fmt.Errorf("Error while reading %s/%s!%s:%s",
			es.spreadsheet.ID, sheetname, startcell, endcell))
		}
		cont = len(data.Values)==step

		for _, row := range data.Values {
			if id, err := extractSpreadsheetID(row[0].(string)); err != nil {
				reterr = errors.Join(reterr, err)
			} else {
				sheetids = append(sheetids, id)
			}
		}
		pos += step
		if !cont {
			break
		}
	}

	return sheetids, reterr
}

// extractSpreadsheetID takes a string input, attempts to extract a Google Spreadsheet ID,
// and returns the ID along with an error status.
func extractSpreadsheetID(input string) (string, error) {
    // Regular expression to match Google Spreadsheet ID
    re := regexp.MustCompile(`^([a-zA-Z0-9_-]{25,})`)

    // Check if input is a valid Google Spreadsheet ID
    if re.MatchString(input) {
        return input, nil
    }

    // Attempt to parse input as a URL
    u, err := url.Parse(input)
    if err != nil {
        return "", fmt.Errorf("invalid input: %w", err)
    }

    // Check if URL is a Google Spreadsheet URL
    if u.Host != "docs.google.com" && u.Host != "drive.google.com" {
        return "", errors.New("input is not a Google Spreadsheet URL or ID")
    }

	// Extract spreadsheet ID from URL
    pathParts := strings.Split(u.Path, "/")
    for _, part := range pathParts {
        if re.MatchString(part) {
            return part, nil
        }
    }

	// If no ID is found in URL path, check query parameters
    query := u.Query()
    id := query.Get("id")
    if id != "" && re.MatchString(id) {
        return id, nil
    }

    return "", errors.New("unable to extract spreadsheet ID from input")
}
