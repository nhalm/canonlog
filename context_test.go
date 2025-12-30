package canonlog

import (
	"context"
	"errors"
	"log/slog"
	"testing"
)

// setTestLogLevel sets the log level for testing and returns a cleanup function.
func setTestLogLevel(level slog.Level) func() {
	old := logLevel.Load()
	logLevel.Store(int32(level))
	return func() { logLevel.Store(old) }
}

func TestNew(t *testing.T) {
	l := New()

	if l == nil {
		t.Fatal("New returned nil")
	}

	if l.fields == nil {
		t.Error("fields map not initialized")
	}

	if l.level != slog.LevelInfo {
		t.Errorf("Expected default level Info, got %v", l.level)
	}

	if l.message != "Completed" {
		t.Errorf("Expected default message 'Completed', got %s", l.message)
	}
}

func TestLoggerDebugAdd(t *testing.T) {
	// Set level to debug so fields are accumulated
	defer setTestLogLevel(slog.LevelDebug)()

	l := New()
	l.DebugAdd("key1", "value1")

	if l.fields["key1"] != "value1" {
		t.Errorf("Expected field key1=value1, got %v", l.fields["key1"])
	}
}

func TestLoggerDebugAddIgnoredWhenLevelHigher(t *testing.T) {
	// Set level to info so debug fields are ignored
	defer setTestLogLevel(slog.LevelInfo)()

	l := New()
	l.DebugAdd("key1", "value1")

	if _, exists := l.fields["key1"]; exists {
		t.Error("Debug field should be ignored when level is Info")
	}
}

func TestLoggerInfoAdd(t *testing.T) {
	defer setTestLogLevel(slog.LevelInfo)()

	l := New()
	l.InfoAdd("key1", "value1")

	if l.fields["key1"] != "value1" {
		t.Errorf("Expected field key1=value1, got %v", l.fields["key1"])
	}
}

func TestLoggerWarnAdd(t *testing.T) {
	defer setTestLogLevel(slog.LevelWarn)()

	l := New()
	l.WarnAdd("key1", "value1")

	if l.fields["key1"] != "value1" {
		t.Errorf("Expected field key1=value1, got %v", l.fields["key1"])
	}

	if l.level != slog.LevelWarn {
		t.Errorf("Expected level Warn after WarnAdd, got %v", l.level)
	}
}

func TestLoggerErrorAdd(t *testing.T) {
	defer setTestLogLevel(slog.LevelError)()

	l := New()
	err := errors.New("test error")
	l.ErrorAdd(err)

	if len(l.errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(l.errors))
	}

	if l.errors[0] != "test error" {
		t.Errorf("Expected error 'test error', got %v", l.errors[0])
	}

	if l.level != slog.LevelError {
		t.Errorf("Expected level Error after ErrorAdd, got %v", l.level)
	}
}

func TestLoggerErrorAddMultiple(t *testing.T) {
	defer setTestLogLevel(slog.LevelError)()

	l := New()
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")

	l.ErrorAdd(err1).ErrorAdd(err2)

	if len(l.errors) != 2 {
		t.Fatalf("Expected 2 errors, got %d", len(l.errors))
	}

	if l.errors[0] != "error 1" {
		t.Errorf("Expected first error 'error 1', got %v", l.errors[0])
	}

	if l.errors[1] != "error 2" {
		t.Errorf("Expected second error 'error 2', got %v", l.errors[1])
	}
}

func TestLoggerErrorAddNil(t *testing.T) {
	defer setTestLogLevel(slog.LevelError)()

	l := New()
	l.ErrorAdd(nil)

	if len(l.errors) != 0 {
		t.Errorf("Expected 0 errors after adding nil, got %d", len(l.errors))
	}

	if l.level != slog.LevelInfo {
		t.Errorf("Expected level to remain Info after nil error, got %v", l.level)
	}
}

func TestLoggerAddMany(t *testing.T) {
	defer setTestLogLevel(slog.LevelInfo)()

	l := New()
	fields := map[string]any{
		"key1": "value1",
		"key2": 123,
		"key3": true,
	}
	l.InfoAddMany(fields)

	for k, v := range fields {
		if l.fields[k] != v {
			t.Errorf("Expected field %s=%v, got %v", k, v, l.fields[k])
		}
	}
}

func TestLoggerSetMessage(t *testing.T) {
	l := New()
	msg := "Custom message"
	l.SetMessage(msg)

	if l.message != msg {
		t.Errorf("Expected message '%s', got '%s'", msg, l.message)
	}
}

func TestLoggerChaining(t *testing.T) {
	defer setTestLogLevel(slog.LevelDebug)()

	l := New()
	result := l.DebugAdd("key1", "value1").
		InfoAdd("key2", "value2").
		SetMessage("Test message")

	if result != l {
		t.Error("Methods should return the same logger instance for chaining")
	}
}

func TestNewContext(t *testing.T) {
	ctx := NewContext(context.Background())

	l := ctx.Value(loggerKey)
	if l == nil {
		t.Fatal("Logger not stored in context")
	}

	if _, ok := l.(*Logger); !ok {
		t.Error("Context value is not a *Logger")
	}
}

func TestGetLogger(t *testing.T) {
	ctx := NewContext(context.Background())
	l := GetLogger(ctx)

	if l == nil {
		t.Fatal("GetLogger returned nil")
	}
}

func TestGetLoggerPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("GetLogger should panic when no logger in context")
		}
	}()

	emptyCtx := context.Background()
	GetLogger(emptyCtx)
}

func TestTryGetLogger(t *testing.T) {
	ctx := NewContext(context.Background())
	l, ok := TryGetLogger(ctx)

	if !ok {
		t.Fatal("TryGetLogger should return true when logger exists")
	}

	if l == nil {
		t.Fatal("TryGetLogger returned nil logger")
	}
}

func TestTryGetLoggerNoLogger(t *testing.T) {
	emptyCtx := context.Background()
	l, ok := TryGetLogger(emptyCtx)

	if ok {
		t.Error("TryGetLogger should return false when no logger in context")
	}

	if l != nil {
		t.Error("TryGetLogger should return nil when no logger in context")
	}
}

func TestInfoAdd_ContextHelper(t *testing.T) {
	defer setTestLogLevel(slog.LevelInfo)()

	ctx := NewContext(context.Background())
	InfoAdd(ctx, "test_key", "test_value")

	l := GetLogger(ctx)
	if l.fields["test_key"] != "test_value" {
		t.Errorf("Expected field test_key=test_value, got %v", l.fields["test_key"])
	}
}

func TestInfoAddMany_ContextHelper(t *testing.T) {
	defer setTestLogLevel(slog.LevelInfo)()

	ctx := NewContext(context.Background())
	fields := map[string]any{
		"key1": "value1",
		"key2": 456,
	}
	InfoAddMany(ctx, fields)

	l := GetLogger(ctx)
	for k, v := range fields {
		if l.fields[k] != v {
			t.Errorf("Expected field %s=%v, got %v", k, v, l.fields[k])
		}
	}
}

func TestErrorAdd_ContextHelper(t *testing.T) {
	defer setTestLogLevel(slog.LevelError)()

	ctx := NewContext(context.Background())
	err := errors.New("context error")
	ErrorAdd(ctx, err)

	l := GetLogger(ctx)
	if len(l.errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(l.errors))
	}

	if l.errors[0] != "context error" {
		t.Errorf("Expected error 'context error', got %v", l.errors[0])
	}

	if l.level != slog.LevelError {
		t.Errorf("Expected level Error, got %v", l.level)
	}
}

func TestHighestLevelTracking(t *testing.T) {
	defer setTestLogLevel(slog.LevelDebug)()

	l := New()
	l.DebugAdd("debug", "value")
	if l.level != slog.LevelInfo {
		t.Errorf("Expected level Info after DebugAdd, got %v", l.level)
	}

	l.InfoAdd("info", "value")
	if l.level != slog.LevelInfo {
		t.Errorf("Expected level Info after InfoAdd, got %v", l.level)
	}

	l.WarnAdd("warn", "value")
	if l.level != slog.LevelWarn {
		t.Errorf("Expected level Warn after WarnAdd, got %v", l.level)
	}

	l.ErrorAdd(errors.New("error"))
	if l.level != slog.LevelError {
		t.Errorf("Expected level Error after ErrorAdd, got %v", l.level)
	}
}
