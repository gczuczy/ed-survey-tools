package vsds

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/gczuczy/ed-survey-tools/pkg/gcp"
)

// OdsSheet implements gcp.Sheet for ODS spreadsheets.
type OdsSheet struct {
	name string
	rows [][]string
}

func (s *OdsSheet) GetName() string {
	return s.name
}

func (s *OdsSheet) Rows() int {
	return len(s.rows)
}

func (s *OdsSheet) Get(row, col int) string {
	if row >= len(s.rows) {
		return ""
	}
	r := s.rows[row]
	if col >= len(r) {
		return ""
	}
	return r[col]
}

func openOdsSheets(
	item gcp.FolderItem,
) ([]gcp.Sheet, error) {
	data, err := gcp.DownloadFile(item.ID)
	if err != nil {
		return nil, err
	}

	zr, err := zip.NewReader(
		bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf(
			"failed to open ods %s as zip: %w", item.Name, err)
	}

	var contentFile *zip.File
	for _, f := range zr.File {
		if f.Name == "content.xml" {
			contentFile = f
			break
		}
	}
	if contentFile == nil {
		return nil, fmt.Errorf(
			"ods %s: content.xml not found", item.Name)
	}

	rc, err := contentFile.Open()
	if err != nil {
		return nil, fmt.Errorf(
			"ods %s: open content.xml: %w", item.Name, err)
	}
	defer rc.Close()

	xmlData, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf(
			"ods %s: read content.xml: %w", item.Name, err)
	}

	tables, err := parseOdsContent(xmlData)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to parse ods %s: %w", item.Name, err)
	}

	sheets := make([]gcp.Sheet, 0, len(tables))
	for i := range tables {
		sheets = append(sheets, &OdsSheet{
			name: tables[i].name,
			rows: tables[i].rows,
		})
	}
	return sheets, nil
}

// --- ODS XML structures ---

// odsTableData is the parsed form of a single ODS table.
type odsTableData struct {
	name string
	rows [][]string
}

// odsDoc is the top-level ODS content.xml structure.
// Element names are matched by local name (namespace-agnostic).
type odsDoc struct {
	Tables []odsXMLTable `xml:"body>spreadsheet>table"`
}

type odsXMLTable struct {
	Name string      `xml:"name,attr"`
	Rows []odsXMLRow `xml:"table-row"`
}

type odsXMLRow struct {
	Repeated int          `xml:"number-rows-repeated,attr"`
	Cells    []odsXMLCell `xml:"table-cell"`
}

type odsXMLCell struct {
	Repeated int        `xml:"number-columns-repeated,attr"`
	Text     odsXMLText `xml:"p"`
}

// odsXMLText captures text:p content including nested elements
// such as text:span used for formatted labels.
type odsXMLText struct {
	Value string
}

func (t *odsXMLText) UnmarshalXML(
	d *xml.Decoder, start xml.StartElement,
) error {
	var buf strings.Builder
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch v := tok.(type) {
		case xml.CharData:
			buf.Write(v)
		case xml.StartElement:
			var nested odsXMLText
			if err := d.DecodeElement(&nested, &v); err != nil {
				return err
			}
			buf.WriteString(nested.Value)
		case xml.EndElement:
			t.Value = buf.String()
			return nil
		}
	}
}

func parseOdsContent(data []byte) ([]odsTableData, error) {
	var doc odsDoc
	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	result := make([]odsTableData, 0, len(doc.Tables))
	for _, t := range doc.Tables {
		result = append(result, expandOdsTable(t))
	}
	return result, nil
}

func expandOdsTable(t odsXMLTable) odsTableData {
	var rows [][]string
	for _, xmlrow := range t.Rows {
		repeat := xmlrow.Repeated
		if repeat == 0 {
			repeat = 1
		}
		cells := expandOdsRow(xmlrow.Cells)
		// A high-repeat empty row is ODS padding at the end
		// of the sheet — add one instance and stop expanding.
		if len(cells) == 0 && repeat > 1 {
			rows = append(rows, cells)
			continue
		}
		for i := 0; i < repeat; i++ {
			rows = append(rows, cells)
		}
	}
	// trim trailing empty rows
	for len(rows) > 0 && len(rows[len(rows)-1]) == 0 {
		rows = rows[:len(rows)-1]
	}
	return odsTableData{name: t.Name, rows: rows}
}

func expandOdsRow(cells []odsXMLCell) []string {
	var row []string
	for _, cell := range cells {
		repeat := cell.Repeated
		if repeat == 0 {
			repeat = 1
		}
		val := cell.Text.Value
		for i := 0; i < repeat; i++ {
			row = append(row, val)
		}
	}
	// trim trailing empty cells
	for len(row) > 0 && row[len(row)-1] == "" {
		row = row[:len(row)-1]
	}
	return row
}
