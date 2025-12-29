# canonlog

[![Go Reference](https://pkg.go.dev/badge/github.com/nhalm/canonlog.svg)](https://pkg.go.dev/github.com/nhalm/canonlog)
[![codecov](https://codecov.io/gh/nhalm/canonlog/branch/main/graph/badge.svg)](https://codecov.io/gh/nhalm/canonlog)
[![Go Report Card](https://goreportcard.com/badge/github.com/nhalm/canonlog)](https://goreportcard.com/report/github.com/nhalm/canonlog)

Canonical Logger - A structured logging library for Go that accumulates context and emits single-line log entries.

## Philosophy

Inspired by logging patterns from companies like Stripe, Uber, and Google, canonlog collects context throughout a unit of work's lifecycle and outputs everything in a single, parseable log line. This approach:

- Uses less storage and bandwidth
- Improves log parsing performance
- Makes debugging easier - all related data in one place
- Reduces log noise and fragmentation

## Features

- **Context-scoped logging** - Accumulate context throughout any unit of work
- **Single-line output** - All data in one structured log entry
- **Level-gated accumulation** - Fields only accumulate if log level is enabled
- **Automatic level escalation** - Final log emits at highest accumulated level
- **UUIDv7 request IDs** - Time-ordered, sortable unique identifiers (RFC 9562)
- **Standard library integration** - Built on Go's `log/slog`
- **Flexible middleware** - Standard library HTTP and chi router support

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

		// Add fields at info level
		canonlog.InfoAdd(ctx, "user_id", "123")
		canonlog.InfoAdd(ctx, "action", "fetch_profile")

		// Add multiple fields at once
		canonlog.InfoAddMany(ctx, map[string]any{
			"cache_hit":  true,
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

		canonlog.InfoAdd(ctx, "user_id", userID)

		w.Write([]byte("OK"))
	})

	http.ListenAndServe(":8080", r)
}
```

## Log Levels

Canonlog supports level-gated accumulation. Fields are only accumulated if the configured log level allows:

```go
canonlog.SetupGlobalLogger("info", "json")  // Only info and above

canonlog.DebugAdd(ctx, "debug_field", "value")  // Ignored - level too low
canonlog.InfoAdd(ctx, "info_field", "value")    // Accumulated
canonlog.WarnAdd(ctx, "warn_field", "value")    // Accumulated, escalates level to Warn
canonlog.ErrorAdd(ctx, "error_field", "value")  // Accumulated, escalates level to Error
```

The final log is emitted at the highest accumulated level. If you call `ErrorAdd`, the log will be emitted at ERROR level regardless of other fields.

## Example Output

### JSON Format

```json
{
  "time": "2025-01-15T10:30:45Z",
  "level": "INFO",
  "msg": "Completed",
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
time=2025-01-15T10:30:45Z level=INFO msg=Completed duration=45.2ms duration_ms=45 requestID=018e8e9e-45a1-7000-8000-123456789abc method=GET path=/api/users/123 user_agent=Mozilla/5.0... remote_ip=192.168.1.1:54321 host=api.example.com status=200 response_size=1024 user_id=123 action=fetch_profile cache_hit=true db_queries=2
```

## API Reference

### Core

**`SetupGlobalLogger(logLevel, logFormat string)`** - Configure global slog logger. Levels: "debug", "info", "warn", "error". Formats: "json", "text".

**`GenerateRequestID() string`** - Generate UUIDv7 request ID.

**`RequestIDGenerator`** - Global variable for custom ID generation. Override to customize.

### Logger

**`New() *Logger`** - Create new logger instance.

**`(*Logger).DebugAdd(key, value) *Logger`** - Add field at debug level (chainable).

**`(*Logger).DebugAddMany(map[string]any) *Logger`** - Add multiple fields at debug level (chainable).

**`(*Logger).InfoAdd(key, value) *Logger`** - Add field at info level (chainable).

**`(*Logger).InfoAddMany(map[string]any) *Logger`** - Add multiple fields at info level (chainable).

**`(*Logger).WarnAdd(key, value) *Logger`** - Add field at warn level, escalates log level (chainable).

**`(*Logger).WarnAddMany(map[string]any) *Logger`** - Add multiple fields at warn level, escalates log level (chainable).

**`(*Logger).ErrorAdd(key, value) *Logger`** - Add field at error level, escalates log level (chainable).

**`(*Logger).ErrorAddMany(map[string]any) *Logger`** - Add multiple fields at error level, escalates log level (chainable).

**`(*Logger).WithError(error) *Logger`** - Add error, sets level to error (chainable).

**`(*Logger).SetMessage(string) *Logger`** - Set custom log message (chainable).

**`(*Logger).Flush(ctx)`** - Emit accumulated log entry.

### Context Helpers

**`NewContext(ctx) context.Context`** - Create context with new logger.

**`GetLogger(ctx) *Logger`** - Retrieve logger from context (creates new if none exists).

**`DebugAdd(ctx, key, value)`** - Add field at debug level.

**`DebugAddMany(ctx, map[string]any)`** - Add multiple fields at debug level.

**`InfoAdd(ctx, key, value)`** - Add field at info level.

**`InfoAddMany(ctx, map[string]any)`** - Add multiple fields at info level.

**`WarnAdd(ctx, key, value)`** - Add field at warn level.

**`WarnAddMany(ctx, map[string]any)`** - Add multiple fields at warn level.

**`ErrorAdd(ctx, key, value)`** - Add field at error level.

**`ErrorAddMany(ctx, map[string]any)`** - Add multiple fields at error level.

**`WithError(ctx, error)`** - Add error to logger in context.

**`Flush(ctx)`** - Emit accumulated log entry.

### Middleware

**`Middleware(generator) func(http.Handler) http.Handler`** - Standard HTTP middleware. Pass `nil` for default UUIDv7 generation.

**`ChiMiddleware(generator) func(http.Handler) http.Handler`** - Chi router middleware with RequestID integration. Pass `nil` for default.

## Multi-Layer Architecture

Canonlog works naturally with layered applications. The context flows through all layers:

```go
// API Handler
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()  // Logger already in context from middleware
	userID := chi.URLParam(r, "id")

	canonlog.InfoAdd(ctx, "handler", "GetUser")

	user, err := h.userService.GetByID(ctx, userID)
	if err != nil {
		canonlog.WithError(ctx, err)
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}

// Service Layer
func (s *UserService) GetByID(ctx context.Context, id string) (*User, error) {
	canonlog.InfoAdd(ctx, "service", "UserService.GetByID")
	canonlog.InfoAdd(ctx, "user_id", id)

	return s.repo.FindByID(ctx, id)
}

// Repository Layer
func (r *UserRepo) FindByID(ctx context.Context, id string) (*User, error) {
	canonlog.InfoAdd(ctx, "repo", "UserRepo.FindByID")

	// Query database...
	canonlog.InfoAdd(ctx, "db_query_ms", queryTime)

	return user, nil
}
```

All fields accumulate and are emitted in a single log line when the middleware flushes.

## Non-HTTP Usage

Use the logger independently for background jobs, workers, or CLI tools:

```go
func processJob(jobID string) error {
	ctx := canonlog.NewContext(context.Background())

	canonlog.InfoAdd(ctx, "job_id", jobID)
	canonlog.InfoAdd(ctx, "worker", "background-processor")

	defer canonlog.Flush(ctx)

	recordsProcessed := 0
	for i := 0; i < 1500; i++ {
		// Process each record
		recordsProcessed++
	}

	canonlog.InfoAdd(ctx, "records_processed", recordsProcessed)
	return nil
}
```

Or use the Logger directly:

```go
func processJob(jobID string) error {
	ctx := context.Background()
	l := canonlog.New()

	l.InfoAdd("job_id", jobID).
		InfoAdd("worker", "background-processor")

	defer l.Flush(ctx)

	// ... processing ...

	l.InfoAdd("records_processed", recordsProcessed)
	return nil
}
```

## License

MIT
