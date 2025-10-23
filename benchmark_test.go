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

func BenchmarkAddRequestField(b *testing.B) {
	ctx := NewRequestContext(context.Background())
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		AddRequestField(ctx, "key", "value")
	}
}

func BenchmarkAddRequestFields(b *testing.B) {
	ctx := NewRequestContext(context.Background())
	fields := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		AddRequestFields(ctx, fields)
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

func BenchmarkGetRequestLogger(b *testing.B) {
	ctx := NewRequestContext(context.Background())
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = GetRequestLogger(ctx)
	}
}

func BenchmarkGetRequestLoggerFallback(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = GetRequestLogger(ctx)
	}
}

func BenchmarkFullRequestCycle(b *testing.B) {
	SetupGlobalLogger("info", "json")
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx := NewRequestContext(context.Background())

		AddRequestField(ctx, "request_id", GenerateRequestID())
		AddRequestField(ctx, "method", "GET")
		AddRequestField(ctx, "path", "/api/users")
		AddRequestFields(ctx, map[string]interface{}{
			"user_id": "123",
			"status": 200,
			"response_size": 1024,
		})

		LogRequest(ctx)
	}
}
