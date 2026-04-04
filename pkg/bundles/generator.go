package bundles

// BundleRunner is implemented by each measurement-type subsystem.
// The framework calls LoadConfig then Generate; all I/O is handled
// by the framework.
type BundleRunner interface {
	// LoadConfig fetches type-specific configuration for the given
	// bundle ID from the database.
	LoadConfig(bundleID int) (any, error)

	// Generate builds the data structure to be serialised. The
	// config argument is the value returned by LoadConfig.
	Generate(config any) (any, error)
}

var registry = map[string]BundleRunner{}

// Register associates a runner with a measurement type. Call from
// service.go before starting the processor.
func Register(measurementType string, runner BundleRunner) {
	registry[measurementType] = runner
}

func get(measurementType string) (BundleRunner, bool) {
	r, ok := registry[measurementType]
	return r, ok
}
