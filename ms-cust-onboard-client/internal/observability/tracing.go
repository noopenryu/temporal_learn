package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
)

type contextKey string

const traceContextKey contextKey = "trace_context"

// TraceContext represents distributed tracing metadata (W3C TraceContext compliant)
type TraceContext struct {
	TraceID      string `json:"trace_id"`
	SpanID       string `json:"span_id"`
	ParentSpanID string `json:"parent_span_id,omitempty"`
}

// GenerateTraceID produces a 32-character hexadecimal string (16 bytes random)
func GenerateTraceID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%032x", 1)
	}
	return hex.EncodeToString(b)
}

// GenerateSpanID produces a 16-character hexadecimal string (8 bytes random)
func GenerateSpanID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%016x", 1)
	}
	return hex.EncodeToString(b)
}

// FromContext extracts TraceContext from context.Context
func FromContext(ctx context.Context) (TraceContext, bool) {
	if ctx == nil {
		return TraceContext{}, false
	}
	tc, ok := ctx.Value(traceContextKey).(TraceContext)
	return tc, ok
}

// WithTraceContext injects TraceContext into context.Context
func WithTraceContext(ctx context.Context, tc TraceContext) context.Context {
	return context.WithValue(ctx, traceContextKey, tc)
}

// ExtractTraceContext reads W3C traceparent or correlation headers from incoming HTTP requests
func ExtractTraceContext(r *http.Request) TraceContext {
	var traceID, parentSpanID string

	// 1. Check W3C traceparent header: 00-<32hex>-<16hex>-<flags>
	traceparent := strings.TrimSpace(r.Header.Get("traceparent"))
	if traceparent != "" {
		parts := strings.Split(traceparent, "-")
		if len(parts) >= 4 && parts[0] == "00" && len(parts[1]) == 32 && len(parts[2]) == 16 {
			traceID = parts[1]
			parentSpanID = parts[2]
		}
	}

	// 2. Fallback to X-Trace-Id / X-Request-Id if not parsed
	if traceID == "" {
		traceID = strings.TrimSpace(r.Header.Get("X-Trace-Id"))
		if traceID == "" {
			traceID = strings.TrimSpace(r.Header.Get("X-Request-Id"))
		}
		// If custom header is not 32 chars, keep or generate fresh
		if len(traceID) != 32 {
			if traceID == "" {
				traceID = GenerateTraceID()
			} else {
				// Pad or format to consistent 32 chars
				clean := strings.ReplaceAll(traceID, "-", "")
				if len(clean) >= 32 {
					traceID = clean[:32]
				} else {
					traceID = fmt.Sprintf("%-32s", clean)
					traceID = strings.ReplaceAll(traceID, " ", "0")
				}
			}
		}
	}

	// 3. Fallback to X-Span-Id for parent
	if parentSpanID == "" {
		parentSpanID = strings.TrimSpace(r.Header.Get("X-Span-Id"))
	}

	spanID := GenerateSpanID()

	return TraceContext{
		TraceID:      traceID,
		SpanID:       spanID,
		ParentSpanID: parentSpanID,
	}
}

// InjectTraceContext injects W3C traceparent and correlation headers into an outbound HTTP request
func InjectTraceContext(ctx context.Context, req *http.Request) {
	tc, ok := FromContext(ctx)
	if !ok || tc.TraceID == "" {
		tc = TraceContext{
			TraceID: GenerateTraceID(),
			SpanID:  GenerateSpanID(),
		}
	}

	childSpanID := GenerateSpanID()

	req.Header.Set("traceparent", fmt.Sprintf("00-%s-%s-01", tc.TraceID, childSpanID))
	req.Header.Set("X-Trace-Id", tc.TraceID)
	req.Header.Set("X-Span-Id", childSpanID)
	if tc.SpanID != "" {
		req.Header.Set("X-Parent-Span-Id", tc.SpanID)
	}
}
