package canonlog

import (
	"testing"

	"github.com/google/uuid"
)

func TestGenerateRequestID(t *testing.T) {
	id1 := GenerateRequestID()
	id2 := GenerateRequestID()

	if id1 == "" {
		t.Error("GenerateRequestID returned empty string")
	}

	if id1 == id2 {
		t.Error("GenerateRequestID returned duplicate IDs")
	}

	if _, err := uuid.Parse(id1); err != nil {
		t.Errorf("GenerateRequestID returned invalid UUID: %v", err)
	}

	parsed, _ := uuid.Parse(id1)
	if parsed.Version() != 7 {
		t.Errorf("Expected UUIDv7, got version %d", parsed.Version())
	}
}

func TestRequestIDGenerator(t *testing.T) {
	originalGenerator := RequestIDGenerator
	defer func() { RequestIDGenerator = originalGenerator }()

	customID := "custom-id-123"
	RequestIDGenerator = func() string {
		return customID
	}

	id := RequestIDGenerator()
	if id != customID {
		t.Errorf("Expected %s, got %s", customID, id)
	}
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
			SetupGlobalLogger(tt.logLevel, tt.logFormat)
		})
	}
}
