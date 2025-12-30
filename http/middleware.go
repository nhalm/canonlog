package http

import (
	"bufio"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/nhalm/canonlog"
)

// Middleware creates standard library HTTP middleware that sets up canonical logging.
// It accumulates request data throughout the request lifecycle and outputs a single log line at the end.
//
// The middleware:
//   - Checks for X-Request-ID header in incoming request
//   - Generates a new UUIDv7 if no request ID is present
//   - Sets X-Request-ID header on response
//   - Captures request/response metadata
//   - Logs everything in a single line when request completes
//
// Optional generator parameter allows per-middleware override of ID generation.
// Pass nil to use the package default (canonlog.RequestIDGenerator).
func Middleware(generator func() string) func(http.Handler) http.Handler {
	if generator == nil {
		generator = canonlog.RequestIDGenerator
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := canonlog.NewContext(r.Context())

			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generator()
			}

			canonlog.InfoAddMany(ctx, map[string]any{
				"requestID":  requestID,
				"method":     r.Method,
				"path":       r.URL.Path,
				"user_agent": r.UserAgent(),
				"remote_ip":  r.RemoteAddr,
				"host":       r.Host,
			})

			w.Header().Set("X-Request-ID", requestID)

			ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			defer func() {
				canonlog.InfoAddMany(ctx, map[string]any{
					"status":        ww.status,
					"response_size": ww.bytesWritten,
				})
				canonlog.Flush(ctx)
			}()

			next.ServeHTTP(ww, r.WithContext(ctx))
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	status       int
	bytesWritten int64
	wroteHeader  bool
}

func (w *responseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriter) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(data)
	w.bytesWritten += int64(n)
	return n, err
}

// Flush implements http.Flusher for streaming responses.
func (w *responseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Hijack implements http.Hijacker for WebSocket upgrades.
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := w.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

// Unwrap returns the underlying ResponseWriter for interface checks.
func (w *responseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// Push implements http.Pusher for HTTP/2 server push.
func (w *responseWriter) Push(target string, opts *http.PushOptions) error {
	if p, ok := w.ResponseWriter.(http.Pusher); ok {
		return p.Push(target, opts)
	}
	return http.ErrNotSupported
}

// ChiMiddleware creates chi-compatible HTTP middleware that integrates with chi's RequestID middleware.
// It accumulates request data throughout the request lifecycle and outputs a single log line at the end.
//
// The middleware:
//   - First checks chi's middleware.GetReqID() for existing request ID
//   - Falls back to X-Request-ID header if chi RequestID is not set
//   - Generates a new UUIDv7 if no request ID is present
//   - Sets X-Request-ID header on response
//   - Captures request/response metadata
//   - Logs everything in a single line when request completes
//
// Optional generator parameter allows per-middleware override of ID generation.
// Pass nil to use the package default (canonlog.RequestIDGenerator).
func ChiMiddleware(generator func() string) func(http.Handler) http.Handler {
	if generator == nil {
		generator = canonlog.RequestIDGenerator
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := canonlog.NewContext(r.Context())

			requestID := middleware.GetReqID(ctx)
			if requestID == "" {
				requestID = r.Header.Get("X-Request-ID")
			}
			if requestID == "" {
				requestID = generator()
			}

			canonlog.InfoAddMany(ctx, map[string]any{
				"requestID":  requestID,
				"method":     r.Method,
				"path":       r.URL.Path,
				"user_agent": r.UserAgent(),
				"remote_ip":  r.RemoteAddr,
				"host":       r.Host,
			})

			w.Header().Set("X-Request-ID", requestID)

			ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			defer func() {
				canonlog.InfoAddMany(ctx, map[string]any{
					"status":        ww.status,
					"response_size": ww.bytesWritten,
				})
				canonlog.Flush(ctx)
			}()

			next.ServeHTTP(ww, r.WithContext(ctx))
		})
	}
}
