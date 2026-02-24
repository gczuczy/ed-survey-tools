package log

import (
	"os"
	"fmt"
	"errors"

	zl "github.com/rs/zerolog"

	"github.com/gczuczy/ed-survey-tools/pkg/config"
)

type Logger = zl.Logger

var (
	rootLogger zl.Logger
)

func Init(cfg *config.LoggingConfig) error {
	var (
		level zl.Level
		err error
	)

	if level, err = zl.ParseLevel(cfg.Level); err != nil {
		return errors.Join(err, fmt.Errorf("Unable to set loglevel to %s", cfg.Level))
	}
	rootLogger = zl.New(os.Stdout).With().Timestamp().Logger().Level(level)
	return nil
}

func GetLogger(system string) Logger {
	return rootLogger.With().Str("subsystem", system).Logger()
}
