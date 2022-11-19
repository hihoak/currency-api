package logger

import (
	"os"

	"github.com/hihoak/currency-api/internal/pkg/config"
	"github.com/rs/zerolog"
)

type Logger struct {
	logg zerolog.Logger
}

func New(cfg config.LoggerSection) *Logger {
	return &Logger{
		logg: zerolog.New(os.Stdout).Level(convertToLevel(cfg.LogLevel)).With().Timestamp().Logger(),
	}
}

func convertToLevel(logLevel string) zerolog.Level {
	switch logLevel {
	case zerolog.LevelDebugValue:
		return zerolog.DebugLevel
	case zerolog.LevelErrorValue:
		return zerolog.ErrorLevel
	case zerolog.LevelFatalValue:
		return zerolog.FatalLevel
	case zerolog.LevelInfoValue:
		return zerolog.InfoLevel
	case zerolog.LevelWarnValue:
		return zerolog.WarnLevel
	default:
		panic("wrong info level value")
	}
}

func (l Logger) Info() *zerolog.Event {
	return l.logg.Info()
}

func (l Logger) Warn() *zerolog.Event {
	return l.logg.Warn()
}

func (l Logger) Debug() *zerolog.Event {
	return l.logg.Debug()
}

func (l Logger) Error() *zerolog.Event {
	return l.logg.Error()
}

func (l Logger) Fatal() *zerolog.Event {
	return l.logg.Fatal()
}
