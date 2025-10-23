package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/nhalm/canonlog"
)

func TestMiddleware(t *testing.T) {
	handler := Middleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		canonlog.AddRequestField(ctx, "test_field", "test_value")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	requestID := rec.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Error("X-Request-ID header not set")
	}
}

func TestMiddlewareWithExistingRequestID(t *testing.T) {
	existingID := "existing-request-id"

	handler := Middleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", existingID)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	requestID := rec.Header().Get("X-Request-ID")
	if requestID != existingID {
		t.Errorf("Expected request ID %s, got %s", existingID, requestID)
	}
}

func TestMiddlewareCustomGenerator(t *testing.T) {
	customID := "custom-generated-id"
	customGenerator := func() string {
		return customID
	}

	handler := Middleware(customGenerator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	requestID := rec.Header().Get("X-Request-ID")
	if requestID != customID {
		t.Errorf("Expected custom ID %s, got %s", customID, requestID)
	}
}

func TestMiddlewareResponseWriter(t *testing.T) {
	handler := Middleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("test response"))
	}))

	req := httptest.NewRequest("POST", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", rec.Code)
	}

	if rec.Body.String() != "test response" {
		t.Errorf("Expected body 'test response', got '%s'", rec.Body.String())
	}
}

func TestResponseWriterWrite(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rec, status: http.StatusOK}

	data := []byte("test data")
	n, err := rw.Write(data)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if n != len(data) {
		t.Errorf("Expected %d bytes written, got %d", len(data), n)
	}

	if rw.bytesWritten != int64(len(data)) {
		t.Errorf("Expected bytesWritten %d, got %d", len(data), rw.bytesWritten)
	}
}

func TestResponseWriterWriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rec, status: http.StatusOK}

	rw.WriteHeader(http.StatusNotFound)

	if rw.status != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rw.status)
	}
}

func TestMiddlewareContextPropagation(t *testing.T) {
	var capturedCtx *canonlog.RequestLogger

	handler := Middleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		capturedCtx = canonlog.GetRequestLogger(ctx)

		canonlog.AddRequestField(ctx, "test", "value")

		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if capturedCtx == nil {
		t.Fatal("Request logger not found in context")
	}
}

func TestChiMiddleware(t *testing.T) {
	handler := ChiMiddleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		canonlog.AddRequestField(ctx, "test_field", "test_value")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	requestID := rec.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Error("X-Request-ID header not set")
	}
}

func TestChiMiddlewareWithChiRequestID(t *testing.T) {
	chiID := "chi-generated-id"

	handler := ChiMiddleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), chimiddleware.RequestIDKey, chiID)
		r = r.WithContext(ctx)

		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), chimiddleware.RequestIDKey, chiID)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	requestID := rec.Header().Get("X-Request-ID")
	if requestID != chiID {
		t.Errorf("Expected chi request ID %s, got %s", chiID, requestID)
	}
}

func TestChiMiddlewareWithHeaderFallback(t *testing.T) {
	headerID := "header-request-id"

	handler := ChiMiddleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", headerID)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	requestID := rec.Header().Get("X-Request-ID")
	if requestID != headerID {
		t.Errorf("Expected header request ID %s, got %s", headerID, requestID)
	}
}

func TestChiMiddlewareCustomGenerator(t *testing.T) {
	customID := "custom-chi-id"
	customGenerator := func() string {
		return customID
	}

	handler := ChiMiddleware(customGenerator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	requestID := rec.Header().Get("X-Request-ID")
	if requestID != customID {
		t.Errorf("Expected custom ID %s, got %s", customID, requestID)
	}
}
