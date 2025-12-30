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
		s := make([]slog.Attr, 0, 32)
		return &s
	},
}

type loggerKeyType string

const loggerKey loggerKeyType = "canonlogger"

// Logger accumulates context throughout a unit of work and logs once at the end.
// It collects fields and metadata as work is processed, then outputs
// everything in a single structured log line when Flush is called.
//
// Logger is safe for concurrent use within a single request. Multiple goroutines
// spawned from the same request can safely add fields to the same logger.
//
// Example usage:
//
//	log := canonlog.New()
//	log.DebugAdd("cache", "hit")
//	log.InfoAdd("user_id", "123")
//	defer log.Flush(ctx)
type Logger struct {
	mu        sync.Mutex
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
		l.mu.Lock()
		l.fields[key] = value
		l.mu.Unlock()
	}
	return l
}

// DebugAddMany adds multiple fields if debug level is enabled.
func (l *Logger) DebugAddMany(fields map[string]any) *Logger {
	if getLogLevel() <= slog.LevelDebug {
		l.mu.Lock()
		for k, v := range fields {
			l.fields[k] = v
		}
		l.mu.Unlock()
	}
	return l
}

// InfoAdd adds a field if info level is enabled.
func (l *Logger) InfoAdd(key string, value any) *Logger {
	if getLogLevel() <= slog.LevelInfo {
		l.mu.Lock()
		l.fields[key] = value
		l.mu.Unlock()
	}
	return l
}

// InfoAddMany adds multiple fields if info level is enabled.
func (l *Logger) InfoAddMany(fields map[string]any) *Logger {
	if getLogLevel() <= slog.LevelInfo {
		l.mu.Lock()
		for k, v := range fields {
			l.fields[k] = v
		}
		l.mu.Unlock()
	}
	return l
}

// WarnAdd adds a field if warn level is enabled and sets level to at least Warn.
func (l *Logger) WarnAdd(key string, value any) *Logger {
	if getLogLevel() <= slog.LevelWarn {
		l.mu.Lock()
		l.fields[key] = value
		if l.level < slog.LevelWarn {
			l.level = slog.LevelWarn
		}
		l.mu.Unlock()
	}
	return l
}

// WarnAddMany adds multiple fields if warn level is enabled and sets level to at least Warn.
func (l *Logger) WarnAddMany(fields map[string]any) *Logger {
	if getLogLevel() <= slog.LevelWarn {
		l.mu.Lock()
		for k, v := range fields {
			l.fields[k] = v
		}
		if l.level < slog.LevelWarn {
			l.level = slog.LevelWarn
		}
		l.mu.Unlock()
	}
	return l
}

// ErrorAdd appends an error to the errors slice and sets level to Error.
// All errors are output as an "errors" array in the final log entry.
func (l *Logger) ErrorAdd(err error) *Logger {
	if err != nil && getLogLevel() <= slog.LevelError {
		errStr := err.Error()
		l.mu.Lock()
		l.errors = append(l.errors, errStr)
		if l.level < slog.LevelError {
			l.level = slog.LevelError
		}
		l.mu.Unlock()
	}
	return l
}

// SetMessage sets the message for the final log entry and returns the logger for chaining.
func (l *Logger) SetMessage(message string) *Logger {
	l.mu.Lock()
	l.message = message
	l.mu.Unlock()
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

	l.mu.Lock()
	level := l.level
	message := l.message

	attrsPtr := attrPool.Get().(*[]slog.Attr)
	attrs := *attrsPtr
	attrs = attrs[:0]

	attrs = append(attrs, slog.Duration("duration", duration))
	attrs = append(attrs, slog.Int64("duration_ms", duration.Milliseconds()))

	for k, v := range l.fields {
		attrs = append(attrs, slog.Any(k, v))
	}

	if len(l.errors) > 0 {
		attrs = append(attrs, slog.Any("errors", l.errors))
	}
	l.mu.Unlock()

	slog.LogAttrs(ctx, level, message, attrs...)

	// Return slice to pool, preserving any capacity growth
	*attrsPtr = attrs
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
