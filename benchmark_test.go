package canonlog

import (
	"context"
	"log/slog"
	"testing"
)

// setTestLogLevel sets the log level for benchmarking and returns a cleanup function.
func setBenchLogLevel(level slog.Level) func() {
	old := logLevel.Load()
	logLevel.Store(int32(level))
	return func() { logLevel.Store(old) }
}

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
	defer setBenchLogLevel(slog.LevelInfo)()

	l := New()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		l.InfoAdd("key", "value")
	}
}

func BenchmarkLoggerInfoAddMany(b *testing.B) {
	defer setBenchLogLevel(slog.LevelInfo)()

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
	defer setBenchLogLevel(slog.LevelInfo)()

	ctx := NewContext(context.Background())
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		InfoAdd(ctx, "key", "value")
	}
}

func BenchmarkInfoAddMany(b *testing.B) {
	defer setBenchLogLevel(slog.LevelInfo)()

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
	defer setBenchLogLevel(slog.LevelInfo)()

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
		_ = GetLogger(ctx)
	}
}

func BenchmarkTryGetLogger(b *testing.B) {
	ctx := NewContext(context.Background())
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = TryGetLogger(ctx)
	}
}

func BenchmarkTryGetLoggerMiss(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = TryGetLogger(ctx)
	}
}

func BenchmarkFullRequestCycle(b *testing.B) {
	defer setBenchLogLevel(slog.LevelInfo)()

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
