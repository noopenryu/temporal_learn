package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type lokiLogEntry struct {
	Timestamp int64
	Line      string
	Service   string
	Level     string
	Env       string
}

// LokiHandler dispatches logs to stdout and asynchronously pushes them in batches to Loki
type LokiHandler struct {
	serviceName   string
	env           string
	lokiPushURL   string
	stdoutHandler slog.Handler
	httpClient    *http.Client
	queue         chan lokiLogEntry
	stopChan      chan struct{}
	wg            sync.WaitGroup
	attrs         []slog.Attr
	groupPrefix   string
}

// LoggerOptions configures the observability logger
type LoggerOptions struct {
	ServiceName string
	Env         string
	LokiURL     string
	QueueSize   int
	BatchSize   int
	FlushPeriod time.Duration
	Level       slog.Level
}

// InitLogger sets up standard JSON logging to stdout and async Loki streaming
func InitLogger(opts LoggerOptions) (*slog.Logger, func()) {
	if opts.Env == "" {
		opts.Env = "local"
	}
	if opts.LokiURL == "" {
		opts.LokiURL = "http://localhost:3100"
	}
	if opts.QueueSize <= 0 {
		opts.QueueSize = 2048
	}
	if opts.BatchSize <= 0 {
		opts.BatchSize = 50
	}
	if opts.FlushPeriod <= 0 {
		opts.FlushPeriod = 500 * time.Millisecond
	}

	pushURL := fmt.Sprintf("%s/loki/api/v1/push", opts.LokiURL)

	stdoutHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: opts.Level,
	})

	lh := &LokiHandler{
		serviceName:   opts.ServiceName,
		env:           opts.Env,
		lokiPushURL:   pushURL,
		stdoutHandler: stdoutHandler,
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
		queue:    make(chan lokiLogEntry, opts.QueueSize),
		stopChan: make(chan struct{}),
	}

	lh.wg.Add(1)
	go lh.batchWorker(opts.BatchSize, opts.FlushPeriod)

	logger := slog.New(lh)
	slog.SetDefault(logger)

	cleanup := func() {
		close(lh.stopChan)
		lh.wg.Wait()
	}

	return logger, cleanup
}

func (h *LokiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.stdoutHandler.Enabled(ctx, level)
}

func (h *LokiHandler) Handle(ctx context.Context, r slog.Record) error {
	logMap := make(map[string]any, 12)
	logMap["timestamp"] = r.Time.UTC().Format(time.RFC3339Nano)
	logMap["level"] = r.Level.String()
	logMap["service"] = h.serviceName
	logMap["env"] = h.env
	logMap["message"] = r.Message

	if tc, ok := FromContext(ctx); ok && tc.TraceID != "" {
		logMap["trace_id"] = tc.TraceID
		logMap["span_id"] = tc.SpanID
		if tc.ParentSpanID != "" {
			logMap["parent_span_id"] = tc.ParentSpanID
		}
	}

	for _, attr := range h.attrs {
		logMap[attr.Key] = attr.Value.Any()
	}

	r.Attrs(func(a slog.Attr) bool {
		logMap[a.Key] = a.Value.Any()
		return true
	})

	lineBytes, err := json.Marshal(logMap)
	if err != nil {
		lineBytes = []byte(fmt.Sprintf(`{"level":"%s","message":"%s"}`, r.Level, r.Message))
	}

	if tc, ok := FromContext(ctx); ok && tc.TraceID != "" {
		r.AddAttrs(slog.String("trace_id", tc.TraceID), slog.String("span_id", tc.SpanID))
	}
	_ = h.stdoutHandler.Handle(ctx, r)

	entry := lokiLogEntry{
		Timestamp: r.Time.UnixNano(),
		Line:      string(lineBytes),
		Service:   h.serviceName,
		Level:     r.Level.String(),
		Env:       h.env,
	}

	select {
	case h.queue <- entry:
	default:
	}

	return nil
}

func (h *LokiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &LokiHandler{
		serviceName:   h.serviceName,
		env:           h.env,
		lokiPushURL:   h.lokiPushURL,
		stdoutHandler: h.stdoutHandler.WithAttrs(attrs),
		httpClient:    h.httpClient,
		queue:         h.queue,
		stopChan:      h.stopChan,
		attrs:         newAttrs,
	}
}

func (h *LokiHandler) WithGroup(name string) slog.Handler {
	return &LokiHandler{
		serviceName:   h.serviceName,
		env:           h.env,
		lokiPushURL:   h.lokiPushURL,
		stdoutHandler: h.stdoutHandler.WithGroup(name),
		httpClient:    h.httpClient,
		queue:         h.queue,
		stopChan:      h.stopChan,
		attrs:         h.attrs,
		groupPrefix:   name,
	}
}

type lokiPushStream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

type lokiPushPayload struct {
	Streams []lokiPushStream `json:"streams"`
}

func (h *LokiHandler) batchWorker(batchSize int, flushPeriod time.Duration) {
	defer h.wg.Done()

	ticker := time.NewTicker(flushPeriod)
	defer ticker.Stop()

	entries := make([]lokiLogEntry, 0, batchSize)

	flush := func() {
		if len(entries) == 0 {
			return
		}

		streamMap := make(map[string]*lokiPushStream)
		for _, e := range entries {
			key := fmt.Sprintf("%s|%s|%s", e.Service, e.Env, e.Level)
			s, exists := streamMap[key]
			if !exists {
				s = &lokiPushStream{
					Stream: map[string]string{
						"service": e.Service,
						"env":     e.Env,
						"level":   e.Level,
					},
					Values: make([][]string, 0, len(entries)),
				}
				streamMap[key] = s
			}
			s.Values = append(s.Values, []string{
				strconv.FormatInt(e.Timestamp, 10),
				e.Line,
			})
		}

		streams := make([]lokiPushStream, 0, len(streamMap))
		for _, s := range streamMap {
			streams = append(streams, *s)
		}

		payload := lokiPushPayload{Streams: streams}
		data, err := json.Marshal(payload)
		if err == nil {
			req, reqErr := http.NewRequest(http.MethodPost, h.lokiPushURL, bytes.NewReader(data))
			if reqErr == nil {
				req.Header.Set("Content-Type", "application/json")
				resp, doErr := h.httpClient.Do(req)
				if doErr == nil {
					_ = resp.Body.Close()
				}
			}
		}

		entries = entries[:0]
	}

	for {
		select {
		case entry := <-h.queue:
			entries = append(entries, entry)
			if len(entries) >= batchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		case <-h.stopChan:
			for {
				select {
				case entry := <-h.queue:
					entries = append(entries, entry)
				default:
					flush()
					return
				}
			}
		}
	}
}
