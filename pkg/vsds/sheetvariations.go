package vsds

import (
	"fmt"
	"strings"
	"strconv"
	"github.com/gczuczy/ed-survey-tools/pkg/gcp"
)

var (
	variantDW3sds = sheetVariant{
		Name:      "DW3 SDS",
		Project:   `DW3 Stellar Density Scans`,
		HeaderRow: 4,
		HeaderChecks: []sheetHeaderCheck{
			{Column: 0, Row: 4, Value: "System"},
			{Column: 2, Row: 4, Value: "System Count"},
			{Column: 1, Row: 5, Value: "0"},
			{Column: 6, Row: 4, Value: "X"},
			{Column: 7, Row: 4, Value: "Z"},
			{Column: 8, Row: 4, Value: "Y"},
		},
		SampleIndicatorColumn: 1,
		SysNameColumn:         0,
		ZSampleColumn:         1,
		SystemCountColumn:     2,
		MaxDistanceColumn:     4,
		MinSampleRatio:        0.45,
	}

	variantDW3log = sheetVariant{
		Name:      "DW3 Logarithmic",
		Project:   `DW3 Logarithmic Density Scans`,
		HeaderRow: 4,
		HeaderChecks: []sheetHeaderCheck{
			{Column: 0, Row: 4, Value: "System"},
			{Column: 2, Row: 4, Value: "System Count"},
			{Column: 1, Row: 5, Value: "-250"},
			{Column: 6, Row: 4, Value: "X"},
			{Column: 7, Row: 4, Value: "Z"},
			{Column: 8, Row: 4, Value: "Y"},
		},
		SampleIndicatorColumn: 1,
		SysNameColumn:         0,
		ZSampleColumn:         1,
		SystemCountColumn:     2,
		MaxDistanceColumn:     4,
		MinSampleRatio:        0.45,
	}

	variantA15X = sheetVariant{
		Name:      "A15X A",
		Project:   `A15X CW Density Scans`,
		HeaderRow: 4,
		HeaderChecks: []sheetHeaderCheck{
			{Column: 0, Row: 4, Value: "System"},
			{Column: 2, Row: 4, Value: "n"},
			{Column: 5, Row: 4, Value: "X"},
			{Column: 6, Row: 4, Value: "Z"},
			{Column: 7, Row: 4, Value: "Y"},
		},
		SampleIndicatorColumn: 1,
		SysNameColumn:         0,
		ZSampleColumn:         1,
		SystemCountColumn:     2,
		MaxDistanceColumn:     3,
		MinSampleRatio:        0.9,
	}

	variantA15Xv1 = sheetVariant{
		Name:      "A15X B",
		Project:   `A15X CW Density Scans`,
		HeaderRow: 5,
		HeaderChecks: []sheetHeaderCheck{
			{Column: 0, Row: 5, Value: "System"},
			{Column: 2, Row: 5, Value: "n"},
			{Column: 5, Row: 5, Value: "X"},
			{Column: 6, Row: 5, Value: "Z"},
			{Column: 7, Row: 5, Value: "Y"},
		},
		SampleIndicatorColumn: 1,
		SysNameColumn:         0,
		ZSampleColumn:         1,
		SystemCountColumn:     2,
		MaxDistanceColumn:     3,
		MinSampleRatio:        0.9,
	}

	sheetVariants = []*sheetVariant{
		&variantDW3sds, &variantDW3log, &variantA15X, &variantA15Xv1,
	}
)

type sheetVariant struct {
	// debug
	Name string
	// project name
	Project string
	// row orientations
	HeaderRow int

	// column header names
	HeaderChecks []sheetHeaderCheck

	// column orientations (0-indexed, A=0)
	SampleIndicatorColumn int
	SysNameColumn         int
	ZSampleColumn         int
	SystemCountColumn     int
	MaxDistanceColumn     int
	// 0..1, minimum ratio of samples filled in the survey sheet
	MinSampleRatio float32
}

// Column and Rows are on the 0-indexed result set, not cell designations
type sheetHeaderCheck struct {
	Column int
	Row int
	Value string
}
func (shc sheetHeaderCheck) String() string {
	return fmt.Sprintf(`(%d,%d,%s)"`, shc.Column, shc.Row, shc.Value)
}

type VariantError struct {
	VariantName string
	Check *sheetHeaderCheck
	Value *string
	Message string
}
func (ve *VariantError) setCheck(check *sheetHeaderCheck) *VariantError {
	ve.Check = check
	return ve
}
func (ve *VariantError) setValue(v string) *VariantError {
	x := v
	ve.Value = &x
	return ve
}
func (ve *VariantError) msgf(format string, args ...any) *VariantError {
	ve.Message = fmt.Sprintf(format, args...)
	return ve
}

func (ve VariantError) Error() string {
	parts := []string{}
	if ve.Check != nil {
		parts = append(parts,
			fmt.Sprintf("Check:%s", ve.Check.String()))
	}
	if ve.Value != nil {
		parts = append(parts, fmt.Sprintf("Value:%s", *ve.Value))
	}
	if len(ve.Message) != 0 {
		parts = append(parts, ve.Message)
	}
	return fmt.Sprintf("Variant Mismatch: %s(%s)", strings.Join(parts, " "))
}

func (sv *sheetVariant) Eval(sheet gcp.Sheet) error {
	reterr := &VariantError{
		VariantName: sv.Name,
	}
	for _, check := range sv.HeaderChecks {
		if sheet.Rows() <= check.Row {
			return reterr.setCheck(&check).
				msgf("Not enough rows has:%d needs:%d",	sheet.Rows(),	check.Row)
		}
		val := sheet.Get(check.Row, check.Column)
		if val != check.Value {
			return reterr.setCheck(&check).setValue(val)
		}
	}

	nsamples := 0
	nzsamples := 0
	for i := sv.HeaderRow + 1; i < sheet.Rows(); i++ {
		if len(sheet.Get(i, sv.ZSampleColumn)) == 0 {
			break
		}
		nzsamples++

		hasSysCount := false
		hasMaxDistance := false

		if syscount, err := strconv.Atoi(
			sheet.Get(i, sv.SystemCountColumn)); err == nil &&
			syscount >= 0 && syscount < 50 {
			hasSysCount = true
		}
		if maxdst, err := parseFloat(
			sheet.Get(i, sv.MaxDistanceColumn), 32); err == nil &&
			maxdst >= 0 && maxdst <= 20 {
			hasMaxDistance = true
		}

		if hasSysCount || hasMaxDistance {
			nsamples++
		}
	}
	// if it has any data, that's good enough for now
	//return float32(nzsamples)*sv.MinSampleRatio < float32(nsamples)
	if nsamples == 0 {
		return reterr.msgf("Sheet has no data")
	}
	return nil
}
