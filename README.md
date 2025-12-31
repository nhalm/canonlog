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
- **Standard library integration** - Built on Go's `log/slog`
- **Zero dependencies** - Only uses Go standard library

## Installation

```bash
go get github.com/nhalm/canonlog
```

## Quick Start

```go
package main

import (
	"context"

	"github.com/nhalm/canonlog"
)

func main() {
	canonlog.SetupGlobalLogger("info", "json")

	ctx := canonlog.NewContext(context.Background())
	defer canonlog.Flush(ctx)

	canonlog.InfoAdd(ctx, "user_id", "123")
	canonlog.InfoAdd(ctx, "action", "fetch_profile")
	canonlog.InfoAddMany(ctx, map[string]any{
		"cache_hit":  true,
		"db_queries": 2,
	})
}
```

## Log Levels

Canonlog supports level-gated accumulation. Fields are only accumulated if the configured log level allows:

```go
canonlog.SetupGlobalLogger("info", "text")  // Only info and above

canonlog.DebugAdd(ctx, "debug_field", "value")  // Ignored - level too low
canonlog.InfoAdd(ctx, "info_field", "value")    // Accumulated
canonlog.WarnAdd(ctx, "warn_field", "value")    // Accumulated, escalates level to Warn
canonlog.ErrorAdd(ctx, err)                     // Appended to errors array, escalates level to Error
```

The final log is emitted at the highest accumulated level. If you call `ErrorAdd`, the log will be emitted at ERROR level regardless of other fields. All errors are collected in an `errors` array for consistent querying.

**Important:** If you set the level to "info", `DebugAdd` calls are silently ignored. This is by design for performance - no work is done when the level is gated.

## Thread Safety

Logger instances are safe for concurrent use. Multiple goroutines spawned from a single request can safely add fields to the same logger:

```go
func processWork(ctx context.Context) {
    var wg sync.WaitGroup
    wg.Add(2)

    go func() {
        defer wg.Done()
        canonlog.InfoAdd(ctx, "task1", "done")  // Safe
    }()

    go func() {
        defer wg.Done()
        canonlog.InfoAdd(ctx, "task2", "done")  // Safe
    }()

    wg.Wait()
}
```

## Example Output

### Text Format (default)

```
time=2025-01-15T10:30:45Z level=INFO msg="" duration=45.2ms duration_ms=45 user_id=123 action=fetch_profile cache_hit=true db_queries=2
```

### JSON Format

```json
{
  "time": "2025-01-15T10:30:45Z",
  "level": "INFO",
  "msg": "",
  "duration": "45.2ms",
  "duration_ms": 45,
  "user_id": "123",
  "action": "fetch_profile",
  "cache_hit": true,
  "db_queries": 2
}
```

## API Reference

### Core

**`SetupGlobalLogger(logLevel, logFormat string)`** - Configure global slog logger. Levels: "debug", "info", "warn", "error". Formats: "text" (default), "json".

### Logger

**`New() *Logger`** - Create new logger instance.

**`(*Logger).DebugAdd(key, value) *Logger`** - Add field at debug level (chainable).

**`(*Logger).DebugAddMany(map[string]any) *Logger`** - Add multiple fields at debug level (chainable).

**`(*Logger).InfoAdd(key, value) *Logger`** - Add field at info level (chainable).

**`(*Logger).InfoAddMany(map[string]any) *Logger`** - Add multiple fields at info level (chainable).

**`(*Logger).WarnAdd(key, value) *Logger`** - Add field at warn level, escalates log level (chainable).

**`(*Logger).WarnAddMany(map[string]any) *Logger`** - Add multiple fields at warn level, escalates log level (chainable).

**`(*Logger).ErrorAdd(err error) *Logger`** - Append error to errors array, escalates log level (chainable).

**`(*Logger).Flush(ctx)`** - Emit accumulated log entry and reset logger for reuse.

### Context Helpers

**`NewContext(ctx) context.Context`** - Create context with new logger.

**`GetLogger(ctx) *Logger`** - Retrieve logger from context for chaining. Panics if no logger exists.

**`TryGetLogger(ctx) (*Logger, bool)`** - Retrieve logger from context without panicking. Returns (nil, false) if no logger.

**`DebugAdd(ctx, key, value)`** - Add field at debug level.

**`DebugAddMany(ctx, map[string]any)`** - Add multiple fields at debug level.

**`InfoAdd(ctx, key, value)`** - Add field at info level.

**`InfoAddMany(ctx, map[string]any)`** - Add multiple fields at info level.

**`WarnAdd(ctx, key, value)`** - Add field at warn level.

**`WarnAddMany(ctx, map[string]any)`** - Add multiple fields at warn level.

**`ErrorAdd(ctx, err error)`** - Append error to errors array, escalates log level.

**`Flush(ctx)`** - Emit accumulated log entry and reset logger for reuse.

## Multi-Layer Architecture

Canonlog works naturally with layered applications. The context flows through all layers:

```go
// Handler
func (h *Handler) GetUser(ctx context.Context, userID string) (*User, error) {
	canonlog.InfoAdd(ctx, "handler", "GetUser")

	user, err := h.userService.GetByID(ctx, userID)
	if err != nil {
		canonlog.ErrorAdd(ctx, err)
		return nil, err
	}

	return user, nil
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

All fields accumulate and are emitted in a single log line when Flush is called.

## Background Jobs

Use canonlog for background jobs, workers, or CLI tools:

```go
func processJob(jobID string) error {
	ctx := canonlog.NewContext(context.Background())
	defer canonlog.Flush(ctx)

	canonlog.InfoAdd(ctx, "job_id", jobID)
	canonlog.InfoAdd(ctx, "worker", "background-processor")

	recordsProcessed := 0
	for i := 0; i < 1500; i++ {
		// Process each record
		recordsProcessed++
	}

	canonlog.InfoAdd(ctx, "records_processed", recordsProcessed)
	return nil
}
```

### Batch Processing

Flush resets the logger, so you can reuse it for multiple log entries:

```go
func processBatches(ctx context.Context, batches []Batch) error {
	ctx = canonlog.NewContext(ctx)

	for _, batch := range batches {
		canonlog.InfoAdd(ctx, "batch_id", batch.ID)
		canonlog.InfoAdd(ctx, "size", len(batch.Items))

		if err := processBatch(ctx, batch); err != nil {
			canonlog.ErrorAdd(ctx, err)
		}

		canonlog.Flush(ctx)  // Emit log line and reset for next batch
	}
	return nil
}
```

Each Flush emits a log entry and resets the logger (clears fields, errors, level, and restarts the duration timer).

### Using GetLogger for Chaining

Use `GetLogger` when you want to chain multiple field additions:

```go
func handleRequest(ctx context.Context, req *Request) {
	// Context helpers - one call at a time
	canonlog.InfoAdd(ctx, "user_id", req.UserID)
	canonlog.InfoAdd(ctx, "action", req.Action)

	// GetLogger - chain multiple additions
	canonlog.GetLogger(ctx).
		InfoAdd("ip", req.RemoteAddr).
		InfoAdd("user_agent", req.UserAgent).
		InfoAddMany(map[string]any{
			"method": req.Method,
			"path":   req.Path,
		})
}
```

Both approaches modify the same logger in the context.

## License

MIT
