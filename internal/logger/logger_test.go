package logger_test

import (
	"casino/internal/logger"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name          string
		config        logger.Config
		expectedLevel slog.Level
	}{
		{
			name: "debug level",
			config: logger.Config{
				Level:  logger.LogLevelDebug,
				Format: logger.LogFormatJSON,
			},
			expectedLevel: slog.LevelDebug,
		},
		{
			name: "info level",
			config: logger.Config{
				Level:  logger.LogLevelInfo,
				Format: logger.LogFormatJSON,
			},
			expectedLevel: slog.LevelInfo,
		},
		{
			name: "warn level",
			config: logger.Config{
				Level:  logger.LogLevelWarn,
				Format: logger.LogFormatJSON,
			},
			expectedLevel: slog.LevelWarn,
		},
		{
			name: "error level",
			config: logger.Config{
				Level:  logger.LogLevelError,
				Format: logger.LogFormatJSON,
			},
			expectedLevel: slog.LevelError,
		},
		{
			name: "invalid level defaults to info",
			config: logger.Config{
				Level:  "unknown",
				Format: logger.LogFormatJSON,
			},
			expectedLevel: slog.LevelInfo,
		},
		{
			name: "text format",
			config: logger.Config{
				Level:  logger.LogLevelInfo,
				Format: logger.LogFormatText,
			},
			expectedLevel: slog.LevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := logger.New(tt.config)
			assert.NotNil(t, l)
			assert.Implements(t, (*logger.Logger)(nil), l)

			// Cast to *slog.Logger to verify level-related behavior
			sl, ok := l.(*slog.Logger)
			if !ok {
				t.Fatalf("expected *slog.Logger, got %T", l)
			}

			ctx := context.Background()
			assert.True(t, sl.Enabled(ctx, tt.expectedLevel), "Logger should be enabled for its own level")
			if tt.expectedLevel > slog.LevelDebug {
				assert.False(t, sl.Enabled(ctx, tt.expectedLevel-1), "Logger should not be enabled for lower levels")
			}
		})
	}
}

func TestDefaultLogger(t *testing.T) {
	config := logger.Config{
		Level:  logger.LogLevelInfo,
		Format: logger.LogFormatJSON,
	}
	l := logger.New(config)
	
	assert.Equal(t, l, slog.Default(), "New should set the default slog logger")
}
