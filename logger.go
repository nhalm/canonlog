package canonlog

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/google/uuid"
)

// RequestIDGenerator is the function used to generate request IDs.
// It can be overridden globally to customize ID generation.
var RequestIDGenerator = GenerateRequestID

// GenerateRequestID generates a new unique request ID using UUIDv7 (time-ordered, sortable)
func GenerateRequestID() string {
	return uuid.Must(uuid.NewV7()).String()
}

// SetupGlobalLogger configures the global slog logger based on provided parameters
func SetupGlobalLogger(logLevel, logFormat string) {
	// Parse log level
	var level slog.Level
	switch strings.ToLower(logLevel) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo // Default to info if unknown
	}

	// Create handler based on format
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: level,
	}

	switch strings.ToLower(logFormat) {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	case "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts) // Default to text
	}

	// Set the global logger
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Log the configuration for debugging
	slog.InfoContext(context.Background(), "Logger configured",
		"level", logLevel,
		"format", logFormat,
		"effective_level", level.String())
}
