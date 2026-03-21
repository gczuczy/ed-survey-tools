package vsds

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/gczuczy/ed-survey-tools/pkg/db"
	"github.com/gczuczy/ed-survey-tools/pkg/gcp"
)

// sheetVariant describes the column layout of one recognised
// spreadsheet format. All coordinates are 0-indexed.
type sheetVariant struct {
	// display / debug name
	Name string
	// resolved project name
	Project string
	// index of the header row
	HeaderRow int
	// cell assertions used to fingerprint the variant
	HeaderChecks []sheetHeaderCheck
	// column indices
	SysNameColumn     int
	ZSampleColumn     int
	SystemCountColumn int
	MaxDistanceColumn int
}

// sheetHeaderCheck is a single cell-value assertion used to
// fingerprint a sheet variant.
// Column and Row are 0-indexed result-set coordinates.
type sheetHeaderCheck struct {
	Column int
	Row    int
	Value  string
}

func (shc sheetHeaderCheck) String() string {
	return fmt.Sprintf(`(%d,%d,%s)"`, shc.Column, shc.Row, shc.Value)
}

// VariantError is returned when a sheet cannot be matched to a
// known variant.
type VariantError struct {
	VariantName string
	Check       *sheetHeaderCheck
	Value       *string
	Message     string
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
		parts = append(parts, fmt.Sprintf("Value:%v", *ve.Value))
	}
	if len(ve.Message) != 0 {
		parts = append(parts, ve.Message)
	}
	return fmt.Sprintf("Variant Mismatch: %s(%s)", ve.VariantName,
		strings.Join(parts, " "))
}

func (sv *sheetVariant) Eval(sheet gcp.Sheet) error {
	reterr := &VariantError{
		VariantName: sv.Name,
	}
	for _, check := range sv.HeaderChecks {
		if sheet.Rows() <= check.Row {
			return reterr.setCheck(&check).
				msgf("Not enough rows has:%d needs:%d",
					sheet.Rows(), check.Row)
		}
		val := sheet.Get(check.Row, check.Column)
		if val != check.Value {
			return reterr.setCheck(&check).setValue(val)
		}
	}

	nsamples := 0
	for i := sv.HeaderRow + 1; i < sheet.Rows(); i++ {
		if len(sheet.Get(i, sv.ZSampleColumn)) == 0 {
			break
		}

		systemName := sheet.Get(i, sv.SysNameColumn)
		if slices.Contains(skipSysNames, systemName) {
			continue
		}

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
	if nsamples == 0 {
		return reterr.msgf("Sheet has no data")
	}
	return nil
}

// VariantService holds the set of known sheet variants loaded from
// the database at the start of a processing run.
type VariantService struct {
	variants []*sheetVariant
}

// NewVariantService loads all sheet variant definitions from the
// database within the given transaction and returns a ready service.
func NewVariantService(txn *db.Transaction) (*VariantService, error) {
	dbVariants, err := txn.FetchVariants()
	if err != nil {
		return nil, err
	}

	variants := make([]*sheetVariant, len(dbVariants))
	for i, dv := range dbVariants {
		sv := &sheetVariant{
			Name:              dv.Name,
			Project:           dv.ProjectName,
			HeaderRow:         dv.HeaderRow,
			SysNameColumn:     dv.SysNameColumn,
			ZSampleColumn:     dv.ZSampleColumn,
			SystemCountColumn: dv.SystemCountColumn,
			MaxDistanceColumn: dv.MaxDistanceColumn,
			HeaderChecks:      make([]sheetHeaderCheck, len(dv.Checks)),
		}
		for j, c := range dv.Checks {
			sv.HeaderChecks[j] = sheetHeaderCheck{
				Column: c.Col,
				Row:    c.Row,
				Value:  c.Value,
			}
		}
		variants[i] = sv
	}

	return &VariantService{variants: variants}, nil
}

// Identify returns the first variant that matches sheet, or an error
// aggregating all mismatch details when no variant matches.
func (vs *VariantService) Identify(sheet gcp.Sheet) (*sheetVariant, error) {
	var varianterr error
	for _, sv := range vs.variants {
		if err := sv.Eval(sheet); err != nil {
			varianterr = errors.Join(varianterr, err)
		} else {
			return sv, nil
		}
	}
	return nil, errors.Join(varianterr, fmt.Errorf(
		"Unable to identify sheet variant for %s", sheet.GetName()))
}
