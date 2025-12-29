# canonlog

[![Go Reference](https://pkg.go.dev/badge/github.com/nhalm/canonlog.svg)](https://pkg.go.dev/github.com/nhalm/canonlog)
[![codecov](https://codecov.io/gh/nhalm/canonlog/branch/main/graph/badge.svg)](https://codecov.io/gh/nhalm/canonlog)
[![Go Report Card](https://goreportcard.com/badge/github.com/nhalm/canonlog)](https://goreportcard.com/report/github.com/nhalm/canonlog)

Canonical Logger - A structured logging library for Go that accumulates request context and emits single-line log entries.

## Philosophy

Inspired by logging patterns from companies like Stripe, Uber, and Google, canonlog collects context throughout a request's lifecycle and outputs everything in a single, parseable log line. This approach:

- Uses less storage and bandwidth
- Improves log parsing performance
- Makes debugging easier - all request data in one place
- Reduces log noise and fragmentation

## Features

- **Request-scoped logging** - Accumulate context throughout request lifecycle
- **Single-line output** - All request data in one structured log entry
- **UUIDv7 request IDs** - Time-ordered, sortable unique identifiers (RFC 9562)
- **Standard library integration** - Built on Go's `log/slog`
- **Flexible middleware** - Standard library HTTP and chi router support
- **Customizable ID generation** - Override default UUIDv7 generation globally or per-middleware

## Installation

```bash
go get github.com/nhalm/canonlog
```

## Quick Start

### Standard Library HTTP

```go
package main

import (
	"net/http"

	"github.com/nhalm/canonlog"
	canonhttp "github.com/nhalm/canonlog/http"
)

func main() {
	// Setup global logger (optional - uses slog defaults if not called)
	canonlog.SetupGlobalLogger("info", "json")

	mux := http.NewServeMux()

	// Add canonical logging middleware
	handler := canonhttp.Middleware(nil)(mux)

	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Add fields throughout request processing
		canonlog.Set(ctx, "user_id", "123")
		canonlog.Set(ctx, "action", "fetch_profile")

		// Add multiple fields at once
		canonlog.SetAll(ctx, map[string]any{
			"cache_hit": true,
			"db_queries": 2,
		})

		w.Write([]byte("OK"))
	})

	http.ListenAndServe(":8080", handler)
}
```

### Chi Router

```go
package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nhalm/canonlog"
	canonhttp "github.com/nhalm/canonlog/http"
)

func main() {
	canonlog.SetupGlobalLogger("info", "json")

	r := chi.NewRouter()

	// Use chi's RequestID middleware (optional but recommended)
	r.Use(middleware.RequestID)

	// Add canonical logging middleware (integrates with chi's RequestID)
	r.Use(canonhttp.ChiMiddleware(nil))

	r.Get("/api/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := chi.URLParam(r, "id")

		canonlog.Set(ctx, "user_id", userID)

		w.Write([]byte("OK"))
	})

	http.ListenAndServe(":8080", r)
}
```

## Example Output

### JSON Format

```json
{
  "time": "2025-01-15T10:30:45Z",
  "level": "INFO",
  "msg": "Request completed",
  "duration": "45.2ms",
  "duration_ms": 45,
  "requestID": "018e8e9e-45a1-7000-8000-123456789abc",
  "method": "GET",
  "path": "/api/users/123",
  "user_agent": "Mozilla/5.0...",
  "remote_ip": "192.168.1.1:54321",
  "host": "api.example.com",
  "status": 200,
  "response_size": 1024,
  "user_id": "123",
  "action": "fetch_profile",
  "cache_hit": true,
  "db_queries": 2
}
```

### Text (logfmt) Format

```
time=2025-01-15T10:30:45Z level=INFO msg="Request completed" duration=45.2ms duration_ms=45 requestID=018e8e9e-45a1-7000-8000-123456789abc method=GET path=/api/users/123 user_agent=Mozilla/5.0... remote_ip=192.168.1.1:54321 host=api.example.com status=200 response_size=1024 user_id=123 action=fetch_profile cache_hit=true db_queries=2
```

## API Reference

### Core

**`SetupGlobalLogger(logLevel, logFormat string)`** - Configure global slog logger. Levels: "debug", "info", "warn", "error". Formats: "json", "text".

**`GenerateRequestID() string`** - Generate UUIDv7 request ID.

**`RequestIDGenerator`** - Global variable for custom ID generation. Override to customize.

### Request Logger

**`NewRequestLogger() *RequestLogger`** - Create new request logger.

**`(*RequestLogger).WithField(key, value) *RequestLogger`** - Add field (chainable).

**`(*RequestLogger).WithFields(map[string]any) *RequestLogger`** - Add multiple fields (chainable).

**`(*RequestLogger).WithError(error) *RequestLogger`** - Add error, sets level to error (chainable).

**`(*RequestLogger).Log(ctx)`** - Emit accumulated log entry.

### Context Helpers

**`Set(ctx, key, value)`** - Add field to logger in context.

**`SetAll(ctx, map[string]any)`** - Add multiple fields to logger in context.

**`SetError(ctx, error)`** - Add error to logger in context.

**`GetLogger(ctx) *RequestLogger`** - Retrieve logger from context.

**`LogRequest(ctx)`** - Manually log accumulated data.

### Middleware

**`Middleware(generator) func(http.Handler) http.Handler`** - Standard HTTP middleware. Pass `nil` for default UUIDv7 generation.

**`ChiMiddleware(generator) func(http.Handler) http.Handler`** - Chi router middleware with RequestID integration. Pass `nil` for default.

## Request ID Flow in Microservices

Canonlog supports request ID propagation across service boundaries:

1. **First-line service**: Generates request ID if not present in incoming request
2. **Downstream services**: Accept request ID from `X-Request-ID` header
3. **All services**: Pass request ID to subsequent downstream calls

```go
// Store request ID in a custom context key for easy access
type contextKey string
const requestIDKey contextKey = "requestID"

// Middleware wrapper to store request ID
func storeRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID != "" {
			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

// Propagate request ID to downstream service
func callDownstream(ctx context.Context, url string) error {
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		req.Header.Set("X-Request-ID", requestID)
	}

	resp, err := http.DefaultClient.Do(req)
	// ...
	return err
}
```

## Non-HTTP Usage

Use the request logger independently for background jobs, workers, or CLI tools:

```go
func processJob(jobID string) error {
	ctx := canonlog.NewRequestContext(context.Background())
	rl := canonlog.GetLogger(ctx)

	rl.WithField("job_id", jobID)
	rl.WithField("worker", "background-processor")

	defer rl.Log(ctx)

	recordsProcessed := 0
	for i := 0; i < 1500; i++ {
		// Process each record
		recordsProcessed++
	}

	rl.WithField("records_processed", recordsProcessed)
	return nil
}
```

## License

MIT
