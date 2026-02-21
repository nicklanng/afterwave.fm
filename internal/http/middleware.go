package http

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// RealIP sets r.RemoteAddr to the client IP from X-Real-IP or X-Forwarded-For
// (first entry), so downstream handlers and logs see the real client IP.
func RealIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ip := r.Header.Get("X-Real-IP"); ip != "" {
			r.RemoteAddr = net.JoinHostPort(strings.TrimSpace(ip), "0")
			next.ServeHTTP(w, r)
			return
		}
		if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
			first := strings.TrimSpace(strings.Split(fwd, ",")[0])
			if first != "" {
				r.RemoteAddr = net.JoinHostPort(first, "0")
			}
		}
		next.ServeHTTP(w, r)
	})
}

// responseRecorder wraps ResponseWriter to record status code and bytes written.
type responseRecorder struct {
	http.ResponseWriter
	status int
	written int64
}

func (r *responseRecorder) WriteHeader(code int) {
	if r.status == 0 {
		r.status = code
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(p []byte) (n int, err error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err = r.ResponseWriter.Write(p)
	r.written += int64(n)
	return n, err
}

func (r *responseRecorder) Status() int {
	if r.status == 0 {
		return http.StatusOK
	}
	return r.status
}

// RequestLogger logs each request with method, path, status, duration, and
// otel trace_id and span_id from the request context (when set by otelhttp).
func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := &responseRecorder{ResponseWriter: w, status: 0}
			start := time.Now()
			next.ServeHTTP(rec, r)
			attrs := []any{
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rec.Status()),
				slog.Duration("duration", time.Since(start)),
			}
			if sc := trace.SpanFromContext(r.Context()).SpanContext(); sc.IsValid() {
				attrs = append(attrs, slog.String("trace_id", sc.TraceID().String()), slog.String("span_id", sc.SpanID().String()))
			}
			logger.Info("request", attrs...)
		})
	}
}

// Recoverer recovers from panics, logs the panic, and returns HTTP 500.
func Recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					attrs := []any{
						slog.Any("panic", err),
						slog.String("method", r.Method),
						slog.String("path", r.URL.Path),
					}
					if sc := trace.SpanFromContext(r.Context()).SpanContext(); sc.IsValid() {
						attrs = append(attrs, slog.String("trace_id", sc.TraceID().String()))
					}
					logger.Error("panic recovered", attrs...)
					w.WriteHeader(http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
