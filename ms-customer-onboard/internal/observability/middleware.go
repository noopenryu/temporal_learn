package observability

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

func (rw *responseWriter) Status() int {
	if rw.statusCode == 0 {
		return http.StatusOK
	}
	return rw.statusCode
}

func (rw *responseWriter) BytesWritten() int {
	return rw.bytesWritten
}

// TracingMiddleware extracts or generates trace context, enriches response headers, and records structured access logs
func TracingMiddleware(serviceName string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tc := ExtractTraceContext(r)
			ctx := WithTraceContext(r.Context(), tc)

			w.Header().Set("X-Trace-Id", tc.TraceID)
			w.Header().Set("X-Span-Id", tc.SpanID)
			w.Header().Set("traceparent", fmt.Sprintf("00-%s-%s-01", tc.TraceID, tc.SpanID))

			rw := &responseWriter{ResponseWriter: w}
			start := time.Now()

			next.ServeHTTP(rw, r.WithContext(ctx))

			duration := time.Since(start)

			slog.InfoContext(ctx, "HTTP request handled",
				slog.Group("http",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.Int("status_code", rw.Status()),
					slog.Float64("duration_ms", float64(duration.Microseconds())/1000.0),
					slog.Int("bytes", rw.BytesWritten()),
					slog.String("client_ip", r.RemoteAddr),
				),
			)
		})
	}
}
