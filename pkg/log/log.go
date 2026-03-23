package log

import (
	"errors"
	"fmt"
	"io"
	"log/syslog"
	"os"
	"strings"

	zl "github.com/rs/zerolog"

	"github.com/gczuczy/ed-survey-tools/pkg/config"
)

type Logger = zl.Logger

var (
	rootLogger zl.Logger
)

var facilityMap = map[string]syslog.Priority{
	"KERN":     syslog.LOG_KERN,
	"USER":     syslog.LOG_USER,
	"MAIL":     syslog.LOG_MAIL,
	"DAEMON":   syslog.LOG_DAEMON,
	"AUTH":     syslog.LOG_AUTH,
	"SYSLOG":   syslog.LOG_SYSLOG,
	"LPR":      syslog.LOG_LPR,
	"NEWS":     syslog.LOG_NEWS,
	"UUCP":     syslog.LOG_UUCP,
	"CRON":     syslog.LOG_CRON,
	"AUTHPRIV": syslog.LOG_AUTHPRIV,
	"FTP":      syslog.LOG_FTP,
	"LOCAL0":   syslog.LOG_LOCAL0,
	"LOCAL1":   syslog.LOG_LOCAL1,
	"LOCAL2":   syslog.LOG_LOCAL2,
	"LOCAL3":   syslog.LOG_LOCAL3,
	"LOCAL4":   syslog.LOG_LOCAL4,
	"LOCAL5":   syslog.LOG_LOCAL5,
	"LOCAL6":   syslog.LOG_LOCAL6,
	"LOCAL7":   syslog.LOG_LOCAL7,
}

type syslogLevelWriter struct {
	w *syslog.Writer
}

func (s *syslogLevelWriter) Write(p []byte) (n int, err error) {
	msg := strings.TrimRight(string(p), "\n")
	return len(p), s.w.Info(msg)
}

func (s *syslogLevelWriter) WriteLevel(
	l zl.Level, p []byte,
) (n int, err error) {
	msg := strings.TrimRight(string(p), "\n")
	switch l {
	case zl.TraceLevel, zl.DebugLevel:
		return len(p), s.w.Debug(msg)
	case zl.InfoLevel:
		return len(p), s.w.Info(msg)
	case zl.WarnLevel:
		return len(p), s.w.Warning(msg)
	case zl.ErrorLevel:
		return len(p), s.w.Err(msg)
	case zl.FatalLevel:
		return len(p), s.w.Crit(msg)
	case zl.PanicLevel:
		return len(p), s.w.Emerg(msg)
	default:
		return len(p), s.w.Info(msg)
	}
}

func initSyslogWriter(
	cfg *config.LoggingConfig,
) (*syslog.Writer, error) {
	sc := &cfg.Syslog
	facility, ok := facilityMap[strings.ToUpper(sc.Facility)]
	if !ok {
		return nil, fmt.Errorf(
			"unknown syslog facility: %s", sc.Facility,
		)
	}
	priority := facility | syslog.LOG_INFO
	if sc.Proto != "" {
		addr := fmt.Sprintf("%s:%d", sc.Host, sc.Port)
		return syslog.Dial(sc.Proto, addr, priority, "edst")
	}
	return syslog.New(priority, "edst")
}

func Init(cfg *config.LoggingConfig) error {
	var (
		level zl.Level
		err   error
	)

	if level, err = zl.ParseLevel(cfg.Level); err != nil {
		return errors.Join(
			err,
			fmt.Errorf("Unable to set loglevel to %s", cfg.Level),
		)
	}

	var w io.Writer
	switch cfg.Output {
	case "syslog":
		sw, err := initSyslogWriter(cfg)
		if err != nil {
			return fmt.Errorf("syslog init failed: %w", err)
		}
		w = &syslogLevelWriter{sw}
	default:
		w = os.Stdout
	}

	ctx := zl.New(w).With()
	if cfg.Timestamp {
		ctx = ctx.Timestamp()
	}
	rootLogger = ctx.Logger().Level(level)
	return nil
}

func GetLogger(system string) Logger {
	return rootLogger.With().Str("subsystem", system).Logger()
}
