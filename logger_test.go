package canonlog

import (
	"log/slog"
	"sync"
	"testing"
)

// resetSetupOnce resets the sync.Once for testing purposes.
// This allows testing SetupGlobalLogger idempotency in isolation.
func resetSetupOnce() {
	setupOnce = sync.Once{}
	logLevel.Store(int32(slog.LevelInfo)) // Reset to default
}

func TestSetupGlobalLogger(t *testing.T) {
	tests := []struct {
		name      string
		logLevel  string
		logFormat string
	}{
		{"debug json", "debug", "json"},
		{"info json", "info", "json"},
		{"warn json", "warn", "json"},
		{"error json", "error", "json"},
		{"info text", "info", "text"},
		{"default level", "unknown", "json"},
		{"default format", "info", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetSetupOnce()
			SetupGlobalLogger(tt.logLevel, tt.logFormat)
		})
	}
}

func TestSetupGlobalLoggerIdempotent(t *testing.T) {
	resetSetupOnce()

	// First call sets to debug
	SetupGlobalLogger("debug", "json")
	firstLevel := getLogLevel()

	if firstLevel != slog.LevelDebug {
		t.Errorf("Expected debug level after first call, got %v", firstLevel)
	}

	// Second call should be ignored (sync.Once)
	SetupGlobalLogger("error", "text")
	secondLevel := getLogLevel()

	if secondLevel != firstLevel {
		t.Errorf("SetupGlobalLogger should only execute once; level changed from %v to %v", firstLevel, secondLevel)
	}
}

func TestSetupGlobalLoggerWarningAlias(t *testing.T) {
	resetSetupOnce()

	SetupGlobalLogger("warning", "text")
	level := getLogLevel()

	if level != slog.LevelWarn {
		t.Errorf("Expected 'warning' to set Warn level, got %v", level)
	}
}
