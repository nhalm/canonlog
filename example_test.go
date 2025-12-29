package canonlog_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/nhalm/canonlog"
	canonhttp "github.com/nhalm/canonlog/http"
)

func ExampleSetupGlobalLogger() {
	canonlog.SetupGlobalLogger("info", "json")
}

func ExampleGenerateRequestID() {
	id := canonlog.GenerateRequestID()
	fmt.Printf("Generated request ID: %s\n", id)
}

func ExampleRequestLogger() {
	ctx := context.Background()
	rl := canonlog.NewRequestLogger()

	rl.WithField("user_id", "123")
	rl.WithField("action", "create_order")
	rl.WithFields(map[string]any{
		"amount":   99.99,
		"currency": "USD",
	})

	defer rl.Log(ctx)
}

func ExampleSet() {
	ctx := canonlog.NewRequestContext(context.Background())

	canonlog.Set(ctx, "user_id", "123")
	canonlog.Set(ctx, "action", "fetch_profile")
}

func ExampleSetAll() {
	ctx := canonlog.NewRequestContext(context.Background())

	canonlog.SetAll(ctx, map[string]any{
		"user_id": "123",
		"org_id":  "456",
		"role":    "admin",
	})
}

func ExampleSetError() {
	ctx := canonlog.NewRequestContext(context.Background())

	err := fmt.Errorf("database connection failed")
	canonlog.SetError(ctx, err)
}

func ExampleMiddleware() {
	mux := http.NewServeMux()

	handler := canonhttp.Middleware(nil)(mux)

	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		canonlog.Set(ctx, "user_id", "123")
		canonlog.SetAll(ctx, map[string]any{
			"action": "list_users",
			"page":   1,
		})

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	req := httptest.NewRequest("GET", "/api/users", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
}

func ExampleMiddleware_customGenerator() {
	customGenerator := func() string {
		return "custom-id-format"
	}

	mux := http.NewServeMux()
	handler := canonhttp.Middleware(customGenerator)(mux)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
}

func ExampleRequestLogger_chainable() {
	ctx := context.Background()

	rl := canonlog.NewRequestLogger()
	rl.WithField("user_id", "123").
		WithField("action", "login").
		WithFields(map[string]any{
			"ip":         "192.168.1.1",
			"user_agent": "Mozilla/5.0",
		})

	defer rl.Log(ctx)
}
