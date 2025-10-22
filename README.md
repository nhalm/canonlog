# canonlog

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
		canonlog.AddRequestField(ctx, "user_id", "123")
		canonlog.AddRequestField(ctx, "action", "fetch_profile")

		// Add multiple fields at once
		canonlog.AddRequestFields(ctx, map[string]any{
			"cache_hit": true,
			"db_queries": 2,
		})

		// Log errors (automatically sets log level to error)
		if err := someOperation(); err != nil {
			canonlog.AddRequestError(ctx, err)
			http.Error(w, "Internal error", 500)
			return
		}

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

		canonlog.AddRequestField(ctx, "user_id", userID)

		w.Write([]byte("OK"))
	})

	http.ListenAndServe(":8080", r)
}
```

## Example Output

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

## API Reference

### Core Functions

#### `SetupGlobalLogger(logLevel, logFormat string)`

Configure the global slog logger.

- `logLevel`: "debug", "info", "warn", "error" (default: "info")
- `logFormat`: "json", "text" (default: "text")

```go
canonlog.SetupGlobalLogger("debug", "json")
```

#### `GenerateRequestID() string`

Generate a UUIDv7-based request ID (time-ordered, sortable).

```go
id := canonlog.GenerateRequestID()
// Returns: "018e8e9e-45a1-7000-8000-123456789abc"
```

#### `RequestIDGenerator`

Global variable to customize request ID generation.

```go
// Use custom ID format
canonlog.RequestIDGenerator = func() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}
```

### Request Logger

#### `NewRequestLogger() *RequestLogger`

Create a new request logger (typically done by middleware).

#### `(*RequestLogger).WithField(key string, value any) *RequestLogger`

Add a single field to the logger (chainable).

```go
rl := canonlog.NewRequestLogger()
rl.WithField("user_id", "123").WithField("action", "login")
```

#### `(*RequestLogger).WithFields(fields map[string]any) *RequestLogger`

Add multiple fields at once (chainable).

#### `(*RequestLogger).WithError(err error) *RequestLogger`

Add an error and set log level to error (chainable).

#### `(*RequestLogger).Log(ctx context.Context)`

Emit the accumulated log entry.

### Context Helpers

#### `AddRequestField(ctx context.Context, key string, value any)`

Add a field to the request logger stored in context.

```go
canonlog.AddRequestField(ctx, "user_id", userID)
```

#### `AddRequestFields(ctx context.Context, fields map[string]any)`

Add multiple fields to the request logger in context.

```go
canonlog.AddRequestFields(ctx, map[string]any{
	"user_id": "123",
	"role": "admin",
})
```

#### `AddRequestError(ctx context.Context, err error)`

Add an error to the request logger in context.

```go
if err != nil {
	canonlog.AddRequestError(ctx, err)
}
```

#### `LogRequest(ctx context.Context)`

Manually log the accumulated request data (normally called by middleware defer).

### Middleware

#### `Middleware(generator func() string) func(http.Handler) http.Handler`

Standard library HTTP middleware.

- Checks `X-Request-ID` header
- Generates UUIDv7 if not present
- Sets `X-Request-ID` response header
- Logs request on completion

```go
// Use default UUIDv7 generation
handler := canonhttp.Middleware(nil)(yourHandler)

// Custom ID generation for this middleware instance
handler := canonhttp.Middleware(func() string {
	return uuid.New().String()
})(yourHandler)
```

#### `ChiMiddleware(generator func() string) func(http.Handler) http.Handler`

Chi router middleware with chi.RequestID integration.

- Checks chi's `middleware.GetReqID()` first
- Falls back to `X-Request-ID` header
- Generates UUIDv7 if not present
- Sets `X-Request-ID` response header
- Logs request on completion

```go
r.Use(middleware.RequestID) // Optional but recommended
r.Use(canonhttp.ChiMiddleware(nil))
```

## Request ID Flow in Microservices

Canonlog supports request ID propagation across service boundaries:

1. **First-line service**: Generates request ID if not present in incoming request
2. **Downstream services**: Accept request ID from `X-Request-ID` header
3. **All services**: Pass request ID to subsequent downstream calls

```go
// In your HTTP client
func callDownstream(ctx context.Context, url string) error {
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

	// Propagate request ID to downstream service
	if rl := canonlog.GetRequestLogger(ctx); rl != nil {
		// Extract requestID from logger fields if you stored it
		// Or store it separately in context for easy access
		req.Header.Set("X-Request-ID", requestID)
	}

	resp, err := http.DefaultClient.Do(req)
	// ...
}
```

## Non-HTTP Usage

Use the request logger independently for background jobs, workers, or CLI tools:

```go
func processJob(jobID string) {
	ctx := canonlog.NewRequestContext(context.Background())
	rl := canonlog.GetRequestLogger(ctx)

	rl.WithField("job_id", jobID)
	rl.WithField("worker", "background-processor")

	defer rl.Log(ctx)

	// Process work
	if err := doWork(); err != nil {
		rl.WithError(err)
		return
	}

	rl.WithField("records_processed", 1500)
}
```

## License

MIT
