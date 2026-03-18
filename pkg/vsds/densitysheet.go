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

func stripCmdrPrefix(name string) string {
	if strings.HasPrefix(strings.ToLower(name), "cmdr ") {
		return name[5:]
	}
	return name
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
	} else {
		m.CMDR = sheet.Get(0, 0)
	}
	// strip the `CMDR ` prefix
	m.CMDR = stripCmdrPrefix(m.CMDR)
	b1name := sheet.Get(0, 1)
	if strings.ToLower(m.CMDR) == "name" && b1name != "" {
		m.CMDR = stripCmdrPrefix(b1name)
	}

	var (
		variant *sheetVariant
		varianterr error
	)
	for _, sv := range sheetVariants {
		if err := sv.Eval(sheet); err != nil {
			varianterr = errors.Join(varianterr, err)
		} else {
			variant = sv
			break
		}
	}
	if variant == nil {
		return m, errors.Join(varianterr, fmt.Errorf(
			"Unable to identify sheet variant for %s", name))
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
		strz := sheet.Get(i, variant.ZSampleColumn)
		if z, err = strconv.Atoi(strz); err != nil {
			continue
		}
		strc := sheet.Get(i, variant.SystemCountColumn)
		if c, err = strconv.Atoi(strc); err != nil {
			continue
		}
		// no data
		if len(strz) == 0 && len(strc) == 0 {
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
		systemName := sheet.Get(i, variant.SysNameColumn)
		// skip on empty system name
		if len(systemName) == 0 || systemName == "-" {
			continue
		}
		// if both are 0, then md=20
		if c == 0 && md < 0.01 {
			md = 20.0
		}
		if md > 20.0 {
			// forgotten dot
			if md > 100 && md < 200 {
				md = md/10
			} else {
				md = 20.0
			}
		}
		dp := vsdstypes.SurveyPoint{
			SystemName:  systemName,
			ZSample:     z,
			Count:       c,
			MaxDistance: float32(md),
		}
		m.SurveyPoints = append(m.SurveyPoints, dp)
	}

	return m, nil
}

