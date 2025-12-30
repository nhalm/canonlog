package canonlog

import (
	"context"
	"log/slog"
	"testing"
)

func BenchmarkGenerateRequestID(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = GenerateRequestID()
	}
}

func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = New()
	}
}

func BenchmarkLoggerInfoAdd(b *testing.B) {
	oldLevel := logLevel
	logLevel = slog.LevelInfo
	defer func() { logLevel = oldLevel }()

	l := New()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		l.InfoAdd("key", "value")
	}
}

func BenchmarkLoggerInfoAddMany(b *testing.B) {
	oldLevel := logLevel
	logLevel = slog.LevelInfo
	defer func() { logLevel = oldLevel }()

	l := New()
	fields := map[string]any{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		l.InfoAddMany(fields)
	}
}

func BenchmarkInfoAdd(b *testing.B) {
	oldLevel := logLevel
	logLevel = slog.LevelInfo
	defer func() { logLevel = oldLevel }()

	ctx := NewContext(context.Background())
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		InfoAdd(ctx, "key", "value")
	}
}

func BenchmarkInfoAddMany(b *testing.B) {
	oldLevel := logLevel
	logLevel = slog.LevelInfo
	defer func() { logLevel = oldLevel }()

	ctx := NewContext(context.Background())
	fields := map[string]any{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		InfoAddMany(ctx, fields)
	}
}

func BenchmarkLoggerFlush(b *testing.B) {
	oldLevel := logLevel
	logLevel = slog.LevelInfo
	defer func() { logLevel = oldLevel }()

	ctx := context.Background()
	SetupGlobalLogger("info", "json")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		l := New()
		l.InfoAdd("user_id", "123")
		l.InfoAdd("action", "test")
		l.InfoAddMany(map[string]any{
			"key1": "value1",
			"key2": 123,
			"key3": true,
		})
		l.Flush(ctx)
	}
}

func BenchmarkGetLogger(b *testing.B) {
	ctx := NewContext(context.Background())
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = getLogger(ctx)
	}
}

func BenchmarkGetLoggerFallback(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = getLogger(ctx)
	}
}

func BenchmarkFullRequestCycle(b *testing.B) {
	oldLevel := logLevel
	logLevel = slog.LevelInfo
	defer func() { logLevel = oldLevel }()

	SetupGlobalLogger("info", "json")
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx := NewContext(context.Background())

		InfoAdd(ctx, "request_id", GenerateRequestID())
		InfoAdd(ctx, "method", "GET")
		InfoAdd(ctx, "path", "/api/users")
		InfoAddMany(ctx, map[string]any{
			"user_id":       "123",
			"status":        200,
			"response_size": 1024,
		})

		Flush(ctx)
	}
}
