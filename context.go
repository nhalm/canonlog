package canonlog

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// attrPool reduces allocations in Flush by reusing attribute slices.
var attrPool = sync.Pool{
	New: func() any {
		s := make([]slog.Attr, 0, 16)
		return &s
	},
}

type loggerKeyType string

const loggerKey loggerKeyType = "canonlogger"

// Logger accumulates context throughout a unit of work and logs once at the end.
// It collects fields and metadata as work is processed, then outputs
// everything in a single structured log line when Flush is called.
//
// Logger is NOT safe for concurrent use. Each unit of work (e.g., HTTP request,
// background job) should have its own Logger instance. The middleware and
// NewContext create a new Logger per request, so concurrent requests are safe.
//
// Example usage:
//
//	log := canonlog.New()
//	log.DebugAdd("cache", "hit")
//	log.InfoAdd("user_id", "123")
//	defer log.Flush(ctx)
type Logger struct {
	startTime time.Time
	fields    map[string]any
	errors    []string
	level     slog.Level
	message   string
}

// New creates a new logger with default settings.
// The logger starts with INFO level and "Completed" as the default message.
func New() *Logger {
	return &Logger{
		startTime: time.Now(),
		fields:    make(map[string]any, 8), // Pre-size for typical request (6-10 fields)
		level:     slog.LevelInfo,
		message:   "Completed",
	}
}

// DebugAdd adds a field if debug level is enabled.
func (l *Logger) DebugAdd(key string, value any) *Logger {
	if getLogLevel() <= slog.LevelDebug {
		l.fields[key] = value
	}
	return l
}

// DebugAddMany adds multiple fields if debug level is enabled.
func (l *Logger) DebugAddMany(fields map[string]any) *Logger {
	if getLogLevel() <= slog.LevelDebug {
		for k, v := range fields {
			l.fields[k] = v
		}
	}
	return l
}

// InfoAdd adds a field if info level is enabled.
func (l *Logger) InfoAdd(key string, value any) *Logger {
	if getLogLevel() <= slog.LevelInfo {
		l.fields[key] = value
	}
	return l
}

// InfoAddMany adds multiple fields if info level is enabled.
func (l *Logger) InfoAddMany(fields map[string]any) *Logger {
	if getLogLevel() <= slog.LevelInfo {
		for k, v := range fields {
			l.fields[k] = v
		}
	}
	return l
}

// WarnAdd adds a field if warn level is enabled and sets level to at least Warn.
func (l *Logger) WarnAdd(key string, value any) *Logger {
	if getLogLevel() <= slog.LevelWarn {
		l.fields[key] = value
		if l.level < slog.LevelWarn {
			l.level = slog.LevelWarn
		}
	}
	return l
}

// WarnAddMany adds multiple fields if warn level is enabled and sets level to at least Warn.
func (l *Logger) WarnAddMany(fields map[string]any) *Logger {
	if getLogLevel() <= slog.LevelWarn {
		for k, v := range fields {
			l.fields[k] = v
		}
		if l.level < slog.LevelWarn {
			l.level = slog.LevelWarn
		}
	}
	return l
}

// ErrorAdd appends an error to the errors slice and sets level to Error.
// All errors are output as an "errors" array in the final log entry.
func (l *Logger) ErrorAdd(err error) *Logger {
	if err != nil && getLogLevel() <= slog.LevelError {
		l.errors = append(l.errors, err.Error())
		if l.level < slog.LevelError {
			l.level = slog.LevelError
		}
	}
	return l
}

// SetMessage sets the message for the final log entry and returns the logger for chaining.
func (l *Logger) SetMessage(message string) *Logger {
	l.message = message
	return l
}

// Flush outputs the accumulated data in a single structured log line.
// It includes the duration since the logger was created, all accumulated fields,
// and any errors that were added.
//
// This method is typically called in a defer statement to ensure logging
// happens even if the handler panics.
func (l *Logger) Flush(ctx context.Context) {
	duration := time.Since(l.startTime)

	attrsPtr := attrPool.Get().(*[]slog.Attr)
	attrs := (*attrsPtr)[:0]

	attrs = append(attrs, slog.Duration("duration", duration))
	attrs = append(attrs, slog.Int64("duration_ms", duration.Milliseconds()))

	for k, v := range l.fields {
		attrs = append(attrs, slog.Any(k, v))
	}

	if len(l.errors) > 0 {
		attrs = append(attrs, slog.Any("errors", l.errors))
	}

	slog.LogAttrs(ctx, l.level, l.message, attrs...)

	// Return slice to pool
	*attrsPtr = attrs[:0]
	attrPool.Put(attrsPtr)
}

// NewContext creates a new context with a logger attached.
// This is typically called by middleware at the start of a request.
func NewContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, loggerKey, New())
}

// GetLogger retrieves the logger from context or panics if none exists.
// Use this when you want to chain multiple field additions:
//
//	canonlog.GetLogger(ctx).
//		InfoAdd("user_id", "123").
//		InfoAdd("action", "login")
func GetLogger(ctx context.Context) *Logger {
	if l, ok := ctx.Value(loggerKey).(*Logger); ok {
		return l
	}
	panic("canonlog: no logger in context")
}

// TryGetLogger retrieves the logger from context without panicking.
// Returns (logger, true) if found, or (nil, false) if no logger exists.
func TryGetLogger(ctx context.Context) (*Logger, bool) {
	l, ok := ctx.Value(loggerKey).(*Logger)
	return l, ok
}

// DebugAdd adds a field to the logger in context if debug level is enabled.
func DebugAdd(ctx context.Context, key string, value any) {
	GetLogger(ctx).DebugAdd(key, value)
}

// DebugAddMany adds multiple fields to the logger in context if debug level is enabled.
func DebugAddMany(ctx context.Context, fields map[string]any) {
	GetLogger(ctx).DebugAddMany(fields)
}

// InfoAdd adds a field to the logger in context if info level is enabled.
func InfoAdd(ctx context.Context, key string, value any) {
	GetLogger(ctx).InfoAdd(key, value)
}

// InfoAddMany adds multiple fields to the logger in context if info level is enabled.
func InfoAddMany(ctx context.Context, fields map[string]any) {
	GetLogger(ctx).InfoAddMany(fields)
}

// WarnAdd adds a field to the logger in context if warn level is enabled.
func WarnAdd(ctx context.Context, key string, value any) {
	GetLogger(ctx).WarnAdd(key, value)
}

// WarnAddMany adds multiple fields to the logger in context if warn level is enabled.
func WarnAddMany(ctx context.Context, fields map[string]any) {
	GetLogger(ctx).WarnAddMany(fields)
}

// ErrorAdd appends an error to the logger in context and sets level to Error.
func ErrorAdd(ctx context.Context, err error) {
	GetLogger(ctx).ErrorAdd(err)
}

// Flush logs the accumulated data from the logger stored in context.
// This is typically called in a defer statement by middleware.
func Flush(ctx context.Context) {
	GetLogger(ctx).Flush(ctx)
}
