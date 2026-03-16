package vsds

import (
	"fmt"
	"errors"
	"strings"
	"strconv"

	"github.com/gczuczy/ed-survey-tools/pkg/gcp"
	vsdstypes "github.com/gczuczy/ed-survey-tools/pkg/vsds/types"
)

func parseFloat(s string, bitSize int) (float64, error) {
	return strconv.ParseFloat(
		strings.ReplaceAll(s, ",", "."), bitSize)
}

func ParseSheet(sheet gcp.Sheet) (vsdstypes.Survey, error) {
	name := sheet.GetName()
	m := vsdstypes.Survey{
		Name:         name,
		SurveyPoints: make([]vsdstypes.SurveyPoint, 0, 32),
	}

	parts := strings.Split(sheet.Get(0, 0), " - ")
	if len(parts) == 2 {
		m.CMDR = parts[0]
		m.Project = parts[1]
	}

	var variant *sheetVariant
	for _, sv := range sheetVariants {
		if evalSheetVariant(sv, sheet) {
			variant = sv
			break
		}
	}
	if variant == nil {
		return m, fmt.Errorf(
			"Unable to identify sheet variant for %s", name)
	}

	var (
		z     int
		c     int
		mdstr string
		md    float64
		err   error
	)
	for i := variant.HeaderRow + 1; i < sheet.Rows(); i++ {
		if sheet.Get(i, variant.ZSampleColumn) == "" {
			break
		}
		if z, err = strconv.Atoi(
			sheet.Get(i, variant.ZSampleColumn)); err != nil {
			continue
		}
		if c, err = strconv.Atoi(
			sheet.Get(i, variant.SystemCountColumn)); err != nil {
			continue
		}
		mdstr = sheet.Get(i, variant.MaxDistanceColumn)
		if mdstr == "" {
			md = 20.0
		} else if md, err = parseFloat(mdstr, 32); err != nil {
			return m, errors.Join(err, fmt.Errorf(
				"Conversion MaxDst error %d/%d '%v': %s",
				i, variant.MaxDistanceColumn, mdstr, name))
		}
		dp := vsdstypes.SurveyPoint{
			SystemName:  sheet.Get(i, variant.SysNameColumn),
			ZSample:     z,
			Count:       c,
			MaxDistance: float32(md),
		}
		m.SurveyPoints = append(m.SurveyPoints, dp)
	}

	return m, nil
}

func evalSheetVariant(sv *sheetVariant, sheet gcp.Sheet) bool {
	for _, check := range sv.HeaderChecks {
		if sheet.Rows() <= check.Row {
			return false
		}
		if sheet.Get(check.Row, check.Column) != check.Value {
			return false
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
	return float32(nzsamples)*sv.MinSampleRatio < float32(nsamples)
}
