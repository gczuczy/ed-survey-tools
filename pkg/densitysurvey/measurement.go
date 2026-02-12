package densitysurvey

import (
	"github.com/gczuczy/ed-survey-tools/pkg/edsm"
)

var (
	// in cli-mode having a global state is a workaround
	edsms *edsm.EDSM
)

type Survey struct {
	CMDR string
	Project string
	Name string
	SurveyPoints []SurveyPoint
}

type SurveyPoint struct {
	X float32
	Y float32
	Z float32
	EDSMID int64
	SystemName string
	ZSample int
	Count int
	MaxDistance float32
}

func (m *Survey) LookupNames() error {

	if edsms == nil {
		edsms = edsm.New()
	}

	names := make([]string, 0, len(m.SurveyPoints))
	for _, dp := range m.SurveyPoints {
		names = append(names, dp.SystemName)
	}

	lookupres, err := edsms.Systems(names)
	if err != nil {
		return err
	}

	// and correlate names
	for i, dp := range m.SurveyPoints {
		for _, sys := range lookupres {
			if sys.Name == dp.SystemName {
				m.SurveyPoints[i].EDSMID = sys.ID
				m.SurveyPoints[i].X = sys.Coords.X
				m.SurveyPoints[i].Y = sys.Coords.Z
				m.SurveyPoints[i].Z = sys.Coords.Y
				break
			}
		}
	}

	return nil
}
