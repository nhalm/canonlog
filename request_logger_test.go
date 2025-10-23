package canonlog

import (
	"context"
	"errors"
	"log/slog"
	"testing"
)

func TestNewRequestLogger(t *testing.T) {
	rl := NewRequestLogger()

	if rl == nil {
		t.Fatal("NewRequestLogger returned nil")
	}

	if rl.fields == nil {
		t.Error("fields map not initialized")
	}

	if rl.errors == nil {
		t.Error("errors slice not initialized")
	}

	if rl.level != slog.LevelInfo {
		t.Errorf("Expected default level Info, got %v", rl.level)
	}

	if rl.message != "Request completed" {
		t.Errorf("Expected default message 'Request completed', got %s", rl.message)
	}
}

func TestRequestLoggerWithField(t *testing.T) {
	rl := NewRequestLogger()
	rl.WithField("key1", "value1")

	if rl.fields["key1"] != "value1" {
		t.Errorf("Expected field key1=value1, got %v", rl.fields["key1"])
	}
}

func TestRequestLoggerWithFields(t *testing.T) {
	rl := NewRequestLogger()
	fields := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
		"key3": true,
	}
	rl.WithFields(fields)

	for k, v := range fields {
		if rl.fields[k] != v {
			t.Errorf("Expected field %s=%v, got %v", k, v, rl.fields[k])
		}
	}
}

func TestRequestLoggerWithError(t *testing.T) {
	rl := NewRequestLogger()
	err := errors.New("test error")
	rl.WithError(err)

	if len(rl.errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(rl.errors))
	}

	if rl.errors[0] != err {
		t.Errorf("Expected error %v, got %v", err, rl.errors[0])
	}

	if rl.level != slog.LevelError {
		t.Errorf("Expected level Error after WithError, got %v", rl.level)
	}

	if rl.message != "Request failed" {
		t.Errorf("Expected message 'Request failed', got %s", rl.message)
	}
}

func TestRequestLoggerWithMultipleErrors(t *testing.T) {
	rl := NewRequestLogger()
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")

	rl.WithError(err1).WithError(err2)

	if len(rl.errors) != 2 {
		t.Fatalf("Expected 2 errors, got %d", len(rl.errors))
	}
}

func TestRequestLoggerSetLevel(t *testing.T) {
	rl := NewRequestLogger()
	rl.SetLevel(slog.LevelWarn)

	if rl.level != slog.LevelWarn {
		t.Errorf("Expected level Warn, got %v", rl.level)
	}
}

func TestRequestLoggerSetMessage(t *testing.T) {
	rl := NewRequestLogger()
	msg := "Custom message"
	rl.SetMessage(msg)

	if rl.message != msg {
		t.Errorf("Expected message '%s', got '%s'", msg, rl.message)
	}
}

func TestRequestLoggerChaining(t *testing.T) {
	rl := NewRequestLogger()
	result := rl.WithField("key1", "value1").
		WithFields(map[string]interface{}{"key2": "value2"}).
		SetLevel(slog.LevelDebug).
		SetMessage("Test message")

	if result != rl {
		t.Error("Methods should return the same logger instance for chaining")
	}
}

func TestNewRequestContext(t *testing.T) {
	ctx := NewRequestContext(context.Background())

	rl := ctx.Value(requestLoggerKey)
	if rl == nil {
		t.Fatal("Request logger not stored in context")
	}

	if _, ok := rl.(*RequestLogger); !ok {
		t.Error("Context value is not a *RequestLogger")
	}
}

func TestGetRequestLogger(t *testing.T) {
	ctx := NewRequestContext(context.Background())
	rl := GetRequestLogger(ctx)

	if rl == nil {
		t.Fatal("GetRequestLogger returned nil")
	}

	emptyCtx := context.Background()
	rl2 := GetRequestLogger(emptyCtx)

	if rl2 == nil {
		t.Fatal("GetRequestLogger should return a new logger if none exists")
	}
}

func TestAddRequestField(t *testing.T) {
	ctx := NewRequestContext(context.Background())
	AddRequestField(ctx, "test_key", "test_value")

	rl := GetRequestLogger(ctx)
	if rl.fields["test_key"] != "test_value" {
		t.Errorf("Expected field test_key=test_value, got %v", rl.fields["test_key"])
	}
}

func TestAddRequestFields(t *testing.T) {
	ctx := NewRequestContext(context.Background())
	fields := map[string]interface{}{
		"key1": "value1",
		"key2": 456,
	}
	AddRequestFields(ctx, fields)

	rl := GetRequestLogger(ctx)
	for k, v := range fields {
		if rl.fields[k] != v {
			t.Errorf("Expected field %s=%v, got %v", k, v, rl.fields[k])
		}
	}
}

func TestAddRequestError(t *testing.T) {
	ctx := NewRequestContext(context.Background())
	err := errors.New("context error")
	AddRequestError(ctx, err)

	rl := GetRequestLogger(ctx)
	if len(rl.errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(rl.errors))
	}

	if rl.errors[0] != err {
		t.Errorf("Expected error %v, got %v", err, rl.errors[0])
	}
}
