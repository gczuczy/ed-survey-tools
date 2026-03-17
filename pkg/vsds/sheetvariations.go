package vsds



var (
	variantDW3sds = sheetVariant{
		Name:      "DW3",
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
		Name:      "DW3",
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
		Name:      "A15X",
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
		Name:      "A15X",
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
