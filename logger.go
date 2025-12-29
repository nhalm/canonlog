// Package canonlog provides structured logging with context accumulation.
//
// Canonlog collects context throughout a unit of work's lifecycle and outputs everything
// in a single, parseable log line. This approach reduces log noise, improves performance,
// and makes debugging easier by keeping all related data together.
//
// The package is built on Go's standard log/slog and provides middleware for both
// standard library HTTP and chi routers.
package canonlog

import (
	"log/slog"
	"os"
	"strings"

	"github.com/google/uuid"
)

// logLevel stores the configured log level for filtering accumulation.
var logLevel = slog.LevelInfo

// RequestIDGenerator is the function used to generate request IDs.
// It can be overridden globally to customize ID generation.
//
// By default, it uses GenerateRequestID which produces UUIDv7 identifiers.
// You can override this to use custom ID formats:
//
//	canonlog.RequestIDGenerator = func() string {
//		return fmt.Sprintf("req_%d", time.Now().UnixNano())
//	}
var RequestIDGenerator = GenerateRequestID

// GenerateRequestID generates a new unique request ID using UUIDv7.
// UUIDv7 identifiers are time-ordered and sortable, making them ideal for
// distributed tracing and log correlation.
func GenerateRequestID() string {
	return uuid.Must(uuid.NewV7()).String()
}

// SetupGlobalLogger configures the global slog logger with the specified level and format.
//
// Valid log levels: "debug", "info", "warn", "warning", "error" (default: "info")
// Valid formats: "json", "text" (default: "text")
//
// Example:
//
//	canonlog.SetupGlobalLogger("debug", "json")
func SetupGlobalLogger(levelStr, logFormat string) {
	// Parse log level
	var level slog.Level
	switch strings.ToLower(levelStr) {
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

	// Store the level for accumulation filtering
	logLevel = level

	// Set the global logger
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
