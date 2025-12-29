package canonlog

import (
	"context"
	"testing"
)

func BenchmarkGenerateRequestID(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = GenerateRequestID()
	}
}

func BenchmarkNewRequestLogger(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = NewRequestLogger()
	}
}

func BenchmarkRequestLoggerWithField(b *testing.B) {
	rl := NewRequestLogger()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rl.WithField("key", "value")
	}
}

func BenchmarkRequestLoggerWithFields(b *testing.B) {
	rl := NewRequestLogger()
	fields := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rl.WithFields(fields)
	}
}

func BenchmarkSet(b *testing.B) {
	ctx := NewRequestContext(context.Background())
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		Set(ctx, "key", "value")
	}
}

func BenchmarkSetAll(b *testing.B) {
	ctx := NewRequestContext(context.Background())
	fields := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		SetAll(ctx, fields)
	}
}

func BenchmarkRequestLoggerLog(b *testing.B) {
	ctx := context.Background()
	SetupGlobalLogger("info", "json")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rl := NewRequestLogger()
		rl.WithField("user_id", "123")
		rl.WithField("action", "test")
		rl.WithFields(map[string]interface{}{
			"key1": "value1",
			"key2": 123,
			"key3": true,
		})
		rl.Log(ctx)
	}
}

func BenchmarkRequestLoggerLogWithError(b *testing.B) {
	ctx := context.Background()
	SetupGlobalLogger("error", "json")
	testErr := context.DeadlineExceeded

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rl := NewRequestLogger()
		rl.WithField("user_id", "123")
		rl.WithError(testErr)
		rl.Log(ctx)
	}
}

func BenchmarkGetLogger(b *testing.B) {
	ctx := NewRequestContext(context.Background())
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = GetLogger(ctx)
	}
}

func BenchmarkGetLoggerFallback(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = GetLogger(ctx)
	}
}

func BenchmarkFullRequestCycle(b *testing.B) {
	SetupGlobalLogger("info", "json")
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx := NewRequestContext(context.Background())

		Set(ctx, "request_id", GenerateRequestID())
		Set(ctx, "method", "GET")
		Set(ctx, "path", "/api/users")
		SetAll(ctx, map[string]interface{}{
			"user_id":       "123",
			"status":        200,
			"response_size": 1024,
		})

		LogRequest(ctx)
	}
}
