package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"
)

var _requestCounter uint64

type contextKey string

const RequestIDKey contextKey = "request_id"

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Add a request ID to the request context
		requestID := atomic.AddUint64(&_requestCounter, 1)
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		r = r.WithContext(ctx)

		// Create a custom ResponseWriter to capture status code
		wrappedWriter := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		slog.Info("New request",
			"request_id", requestID,
			"method", r.Method,
			"path", r.URL.Path,
			"query", r.URL.RawQuery,
		)

		next.ServeHTTP(wrappedWriter, r)

		duration := time.Since(start)
		slog.Info("Request completed",
			"request_id", requestID,
			"status", wrappedWriter.statusCode,
			"size", wrappedWriter.size,
			"duration", duration.String(),
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size

	return size, err
}

func GetRequestID(ctx context.Context) uint64 {
	if id, ok := ctx.Value(RequestIDKey).(uint64); ok {
		return id
	}
	return 0 // 0 is the fallback value if the request ID is not found
}
