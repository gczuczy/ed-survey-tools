package vsds

import (
	"fmt"
	"slices"
	"errors"
	"strings"
	"strconv"

	"github.com/gczuczy/ed-survey-tools/pkg/gcp"
	vsdstypes "github.com/gczuczy/ed-survey-tools/pkg/vsds/types"
)

var (
	skipSysNames = []string{
		"-", "N/A", "", "NO SYSTEM FOUND",
	}
)

func parseFloat(s string, bitSize int) (float64, error) {
	s = strings.ReplaceAll(s, ",", ".")
	s = strings.ReplaceAll(s, "..", ".")
	return strconv.ParseFloat(s, bitSize)
}

func stripCmdrPrefix(name string) string {
	if strings.HasPrefix(strings.ToLower(name), "cmdr ") {
		return name[5:]
	}
	return name
}

// cleanCmdrName extracts and normalises a CMDR name from a cell
// that should contain "CMDR Name - Project Name".  People often
// omit the leading space before the dash, wrap the name in asterisks
// or quotes, or leave the field blank.
func cleanCmdrName(cell string) string {
	// Try the canonical separator first, fall back to "- " (no
	// leading space) which is the most common mis-format.
	var name string
	if parts := strings.SplitN(cell, " - ", 2); len(parts) == 2 {
		name = parts[0]
	} else if parts := strings.SplitN(cell, "- ", 2); len(parts) == 2 {
		name = parts[0]
	} else {
		name = cell
	}
	name = stripCmdrPrefix(name)
	name = strings.Trim(name, "*")
	name = strings.Trim(name, "\"")
	name = strings.TrimSpace(name)
	return name
}

func ParseSheet(
	sheet gcp.Sheet, vs *VariantService,
) (vsdstypes.Survey, error) {
	name := sheet.GetName()
	m := vsdstypes.Survey{
		Name:         name,
		SurveyPoints: make([]vsdstypes.SurveyPoint, 0, 32),
	}

	cell00 := sheet.Get(0, 0)
	m.CMDR = cleanCmdrName(cell00)
	b1name := sheet.Get(0, 1)
	if strings.ToLower(m.CMDR) == "name" && b1name != "" {
		m.CMDR = cleanCmdrName(b1name)
	}

	variant, err := vs.Identify(sheet)
	if err != nil {
		return m, err
	}
	// fix the project name
	m.Project = variant.Project

	var (
		z     int
		c     int
		mdstr string
		md    float64
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
		if mdstr == "" || strings.EqualFold(mdstr, "n/a") {
			md = 20.0
		} else if md, err = parseFloat(mdstr, 32); err != nil {
			return m, errors.Join(err, fmt.Errorf(
				"Conversion MaxDst error %d/%d '%v': %s",
				i, variant.MaxDistanceColumn, mdstr, name))
		}
		systemName := sheet.Get(i, variant.SysNameColumn)
		// skip on empty system name
		if len(systemName) == 0 || slices.Contains(skipSysNames, systemName) {
			continue
		}
		// both zero means no real data — skip rather than fabricate
		if c == 0 && md < 0.01 {
			continue
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

