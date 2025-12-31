package canonlog

import (
	"testing"
)

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
			SetupGlobalLogger(tt.logLevel, tt.logFormat)
		})
	}
}
