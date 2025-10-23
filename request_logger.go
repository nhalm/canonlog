package canonlog

import (
	"context"
	"log/slog"
	"time"
)

type requestLoggerKeyType string

const requestLoggerKey requestLoggerKeyType = "requestLogger"

// RequestLogger accumulates context throughout a request and logs once at the end.
// It collects fields, errors, and metadata as the request is processed, then outputs
// everything in a single structured log line when Log is called.
//
// Example usage:
//
//	rl := canonlog.NewRequestLogger()
//	rl.WithField("user_id", "123")
//	rl.WithField("action", "create_order")
//	defer rl.Log(ctx)
type RequestLogger struct {
	startTime time.Time
	fields    map[string]interface{}
	errors    []error
	level     slog.Level
	message   string
}

// NewRequestLogger creates a new request logger with default settings.
// The logger starts with INFO level and "Request completed" as the default message.
func NewRequestLogger() *RequestLogger {
	return &RequestLogger{
		startTime: time.Now(),
		fields:    make(map[string]interface{}),
		errors:    make([]error, 0),
		level:     slog.LevelInfo,
		message:   "Request completed",
	}
}

// WithField adds a field to the request logger and returns the logger for chaining.
//
// Example:
//
//	rl.WithField("user_id", "123").WithField("action", "login")
func (rl *RequestLogger) WithField(key string, value interface{}) *RequestLogger {
	rl.fields[key] = value
	return rl
}

// WithFields adds multiple fields to the request logger and returns the logger for chaining.
//
// Example:
//
//	rl.WithFields(map[string]any{
//		"user_id": "123",
//		"role": "admin",
//	})
func (rl *RequestLogger) WithFields(fields map[string]interface{}) *RequestLogger {
	for k, v := range fields {
		rl.fields[k] = v
	}
	return rl
}

// WithError adds an error to the request logger, sets the log level to ERROR,
// and changes the message to "Request failed". Returns the logger for chaining.
//
// Multiple errors can be added and will all be logged together.
func (rl *RequestLogger) WithError(err error) *RequestLogger {
	if err != nil {
		rl.errors = append(rl.errors, err)
		rl.level = slog.LevelError
		rl.message = "Request failed"
	}
	return rl
}

// SetLevel sets the log level for the final log entry and returns the logger for chaining.
func (rl *RequestLogger) SetLevel(level slog.Level) *RequestLogger {
	rl.level = level
	return rl
}

// SetMessage sets the message for the final log entry and returns the logger for chaining.
func (rl *RequestLogger) SetMessage(message string) *RequestLogger {
	rl.message = message
	return rl
}

// Log outputs the accumulated request data in a single structured log line.
// It includes the duration since the logger was created, all accumulated fields,
// and any errors that were added.
//
// This method is typically called in a defer statement to ensure the request
// is logged even if the handler panics.
func (rl *RequestLogger) Log(ctx context.Context) {
	// Calculate request duration
	duration := time.Since(rl.startTime)

	// Build attributes
	attrs := make([]slog.Attr, 0, len(rl.fields)+3)

	// Add duration
	attrs = append(attrs, slog.Duration("duration", duration))
	attrs = append(attrs, slog.Int64("duration_ms", duration.Milliseconds()))

	// Add all accumulated fields
	for k, v := range rl.fields {
		attrs = append(attrs, slog.Any(k, v))
	}

	// Add errors if any
	if len(rl.errors) > 0 {
		if len(rl.errors) == 1 {
			attrs = append(attrs, slog.String("error", rl.errors[0].Error()))
		} else {
			errorStrings := make([]string, len(rl.errors))
			for i, err := range rl.errors {
				errorStrings[i] = err.Error()
			}
			attrs = append(attrs, slog.Any("errors", errorStrings))
		}
	}

	// Log everything in one line
	slog.LogAttrs(ctx, rl.level, rl.message, attrs...)
}

// NewRequestContext creates a new context with a request logger attached.
// This is typically called by middleware at the start of a request.
func NewRequestContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, requestLoggerKey, NewRequestLogger())
}

// GetRequestLogger retrieves the request logger from context.
// If no logger exists in the context, it returns a new logger as a fallback.
func GetRequestLogger(ctx context.Context) *RequestLogger {
	if rl, ok := ctx.Value(requestLoggerKey).(*RequestLogger); ok {
		return rl
	}
	return NewRequestLogger()
}

// AddRequestField adds a field to the request logger stored in context.
// This is the primary way to add context to a request during processing.
//
// Example:
//
//	canonlog.AddRequestField(ctx, "user_id", userID)
func AddRequestField(ctx context.Context, key string, value interface{}) {
	if rl := GetRequestLogger(ctx); rl != nil {
		rl.WithField(key, value)
	}
}

// AddRequestFields adds multiple fields to the request logger stored in context.
//
// Example:
//
//	canonlog.AddRequestFields(ctx, map[string]any{
//		"user_id": userID,
//		"org_id": orgID,
//	})
func AddRequestFields(ctx context.Context, fields map[string]interface{}) {
	if rl := GetRequestLogger(ctx); rl != nil {
		rl.WithFields(fields)
	}
}

// AddRequestError adds an error to the request logger stored in context.
// This automatically sets the log level to ERROR and changes the message to "Request failed".
//
// Example:
//
//	if err := db.Query(ctx, ...); err != nil {
//		canonlog.AddRequestError(ctx, err)
//		return err
//	}
func AddRequestError(ctx context.Context, err error) {
	if rl := GetRequestLogger(ctx); rl != nil {
		rl.WithError(err)
	}
}

// LogRequest logs the accumulated request data from the logger stored in context.
// This is typically called in a defer statement by middleware, but can be called
// manually if needed.
func LogRequest(ctx context.Context) {
	if rl := GetRequestLogger(ctx); rl != nil {
		rl.Log(ctx)
	}
}
