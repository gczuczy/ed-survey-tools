package cli

import (
	"os"
	"fmt"
	"github.com/knadh/koanf/v2"

	"github.com/gczuczy/ed-survey-tools/pkg/google"
	"github.com/gczuczy/ed-survey-tools/pkg/config"
	"github.com/gczuczy/ed-survey-tools/pkg/db"
	ds "github.com/gczuczy/ed-survey-tools/pkg/densitysurvey"
)

func Run() {
	var cfg *config.Config

	k := koanf.New(".")

	err := parseArgs(k)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(1)
	}

	if cfg, err = config.ParseConfig(k); err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(1)
	}

	if err = db.Init(&cfg.DB); err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(1)
	}

	creds := k.String(`sa-creds`)
	ss, err := google.NewSheets(creds)
	if err != nil {
		fmt.Printf("Credentials error: %s", creds)
		return
	}

	entry, err := ds.NewEntrySheet(k.String(`sheetid`), ss)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	ids, err := entry.GetSheetIDs()
	if err != nil && len(ids) == 0 {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, sheetid := range ids {
		fmt.Printf("SheetID: %s\n", sheetid)
		dss, err := ds.NewDensitySpreadsheet(sheetid, ss)
		if err != nil {
			fmt.Printf("Error in sheet %s: %v\n", sheetid, err)
			continue
		}

		ms, err := dss.GetSurveys()
		if err != nil {
			fmt.Printf("Measurement error in sheet %s: %v\n", sheetid, err)
			continue
		}
		for _, m := range ms {
			if err = m.LookupNames(); err != nil {
				fmt.Printf(" !! Lookupnames failed: %v\n", err)
			}
		}
		//fmt.Printf("Measurements in %s: %d\n", sheetid, len(ms))
		//fmt.Printf("M: %+v\n", ms)
		//os.Exit(0)

		for _, m := range ms {
			if err = db.Pool.AddSurvey(&m); err != nil {
				fmt.Printf("AddMeasurement (%s): %v\n%+v\n\n", sheetid, err, m)
			}
		}
	}
}
