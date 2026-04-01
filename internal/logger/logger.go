package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/rs/zerolog"
	slogzerolog "github.com/samber/slog-zerolog"
)

type Config struct {
	Format string `env:"FORMAT" env-default:"json" env-description:"Log output format (text or json)" json:"format"`
	Level  string `env:"LEVEL" env-default:"info" env-description:"Log level (debug, info, warn, error)" json:"level"`
}

const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

const (
	LogFormatJSON = "json"
	LogFormatText = "text"
)

func New(conf Config) Logger {
	confLevel := strings.ToLower(conf.Level)
	var level slog.Level
	switch confLevel {
	case LogLevelDebug:
		level = slog.LevelDebug
	case LogLevelWarn:
		level = slog.LevelWarn
	case LogLevelError:
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	var zerologLogger zerolog.Logger
	format := strings.ToLower(conf.Format)
	switch format {
	case LogFormatText:
		zerologLogger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Caller().CallerWithSkipFrameCount(5).Logger()
	default:
		zerologLogger = zerolog.New(os.Stderr)
	}

	loggerConfig := slogzerolog.Option{
		Level:  level,
		Logger: &zerologLogger,
	}.NewZerologHandler()

	logger := slog.New(loggerConfig)
	slog.SetDefault(logger)

	return logger
}

func NewNoop() Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

// Logger is a logger interface that restricts the slog.Logger to the methods we use.
type Logger interface {
	DebugContext(ctx context.Context, msg string, args ...any)
	InfoContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
}
