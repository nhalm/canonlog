package canonlog_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/nhalm/canonlog"
	canonhttp "github.com/nhalm/canonlog/http"
)

func ExampleSetupGlobalLogger() {
	canonlog.SetupGlobalLogger("info", "text")
}

func ExampleGenerateRequestID() {
	id := canonlog.GenerateRequestID()
	fmt.Printf("Generated request ID: %s\n", id)
}

func ExampleLogger() {
	ctx := context.Background()
	l := canonlog.New()

	l.InfoAdd("user_id", "123")
	l.InfoAdd("action", "create_order")
	l.InfoAddMany(map[string]any{
		"amount":   99.99,
		"currency": "USD",
	})

	defer l.Flush(ctx)
}

func ExampleInfoAdd() {
	ctx := canonlog.NewContext(context.Background())

	canonlog.InfoAdd(ctx, "user_id", "123")
	canonlog.InfoAdd(ctx, "action", "fetch_profile")
}

func ExampleInfoAddMany() {
	ctx := canonlog.NewContext(context.Background())

	canonlog.InfoAddMany(ctx, map[string]any{
		"user_id": "123",
		"org_id":  "456",
		"role":    "admin",
	})
}

func ExampleErrorAdd() {
	ctx := canonlog.NewContext(context.Background())

	canonlog.ErrorAdd(ctx, errors.New("payment failed"))
}

func ExampleMiddleware() {
	mux := http.NewServeMux()

	handler := canonhttp.Middleware(nil)(mux)

	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		canonlog.InfoAdd(ctx, "user_id", "123")
		canonlog.InfoAddMany(ctx, map[string]any{
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

func ExampleLogger_chainable() {
	ctx := context.Background()

	l := canonlog.New()
	l.DebugAdd("cache", "hit").
		InfoAdd("user_id", "123").
		InfoAdd("action", "login").
		InfoAddMany(map[string]any{
			"ip":         "192.168.1.1",
			"user_agent": "Mozilla/5.0",
		})

	defer l.Flush(ctx)
}
