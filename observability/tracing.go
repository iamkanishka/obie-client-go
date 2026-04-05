// Package observability integrates the OBIE SDK with OpenTelemetry for
// distributed tracing and metrics. It provides an http.RoundTripper middleware
// that creates spans for every outbound API call and records RED metrics
// (Rate, Errors, Duration).
//
// Usage:
//
//	provider := observability.NewTracerProvider("obie-service")
//	transport := observability.InstrumentedTransport(http.DefaultTransport, "obie")
//	client := &http.Client{Transport: transport}
package observability

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// ────────────────────────────────────────────────────────────────────────────
// Span / Tracer interfaces (thin abstractions to avoid hard OTel dependency)
// ────────────────────────────────────────────────────────────────────────────

// Span represents a single unit of work in a trace.
type Span interface {
	// SetAttribute sets a key-value attribute on the span.
	SetAttribute(key string, value any)
	// RecordError marks the span as errored.
	RecordError(err error)
	// End finalises the span.
	End()
}

// Tracer creates spans.
type Tracer interface {
	// Start creates a new span with the given name, derived from ctx.
	Start(ctx context.Context, name string) (context.Context, Span)
}

// MetricsRecorder records RED metrics for API calls.
type MetricsRecorder interface {
	// RecordRequest records a completed API call.
	RecordRequest(method, url string, statusCode int, duration time.Duration, err error)
}

// ────────────────────────────────────────────────────────────────────────────
// No-op implementations (used when no tracer/recorder is configured)
// ────────────────────────────────────────────────────────────────────────────

type nopSpan struct{}

func (nopSpan) SetAttribute(_ string, _ any) {}
func (nopSpan) RecordError(_ error)                  {}
func (nopSpan) End()                                 {}

type nopTracer struct{}

func (nopTracer) Start(ctx context.Context, _ string) (context.Context, Span) {
	return ctx, nopSpan{}
}

type nopMetrics struct{}

func (nopMetrics) RecordRequest(_ string, _ string, _ int, _ time.Duration, _ error) {}

// ────────────────────────────────────────────────────────────────────────────
// In-process metrics recorder (for testing and simple deployments)
// ────────────────────────────────────────────────────────────────────────────

// RequestRecord stores a single recorded API call.
type RequestRecord struct {
	Method     string
	URL        string
	StatusCode int
	Duration   time.Duration
	Err        error
	RecordedAt time.Time
}

// InMemoryRecorder records API calls in memory and exposes them for inspection.
// It is safe for concurrent use.
type InMemoryRecorder struct {
	mu      sync.RWMutex
	records []RequestRecord
}

// NewInMemoryRecorder creates a recorder that accumulates all API calls.
func NewInMemoryRecorder() *InMemoryRecorder {
	return &InMemoryRecorder{}
}

// RecordRequest implements MetricsRecorder.
func (r *InMemoryRecorder) RecordRequest(method, url string, statusCode int, duration time.Duration, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records = append(r.records, RequestRecord{
		Method:     method,
		URL:        url,
		StatusCode: statusCode,
		Duration:   duration,
		Err:        err,
		RecordedAt: time.Now(),
	})
}

// Records returns a snapshot of all recorded calls.
func (r *InMemoryRecorder) Records() []RequestRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]RequestRecord, len(r.records))
	copy(out, r.records)
	return out
}

// ErrorRate returns the proportion of recorded calls that resulted in errors.
func (r *InMemoryRecorder) ErrorRate() float64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.records) == 0 {
		return 0
	}
	var errs int
	for _, rec := range r.records {
		if rec.Err != nil || rec.StatusCode >= 500 {
			errs++
		}
	}
	return float64(errs) / float64(len(r.records))
}

// AverageDuration returns the mean request duration.
func (r *InMemoryRecorder) AverageDuration() time.Duration {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.records) == 0 {
		return 0
	}
	var total time.Duration
	for _, rec := range r.records {
		total += rec.Duration
	}
	return total / time.Duration(len(r.records))
}

// Flush clears all recorded calls.
func (r *InMemoryRecorder) Flush() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records = nil
}

// ────────────────────────────────────────────────────────────────────────────
// Instrumented transport
// ────────────────────────────────────────────────────────────────────────────

// TransportConfig configures the instrumented transport.
type TransportConfig struct {
	// ComponentName is the service name embedded in spans (e.g. "obie-accounts").
	ComponentName string
	// Tracer creates spans. When nil, a no-op tracer is used.
	Tracer Tracer
	// Metrics records RED metrics. When nil, a no-op recorder is used.
	Metrics MetricsRecorder
	// SanitiseURL removes sensitive query parameters from recorded URLs.
	SanitiseURL func(string) string
}

type instrumentedTransport struct {
	next   http.RoundTripper
	cfg    TransportConfig
}

// NewInstrumentedTransport wraps next with tracing and metrics instrumentation.
func NewInstrumentedTransport(next http.RoundTripper, cfg TransportConfig) http.RoundTripper {
	if next == nil {
		next = http.DefaultTransport
	}
	if cfg.Tracer == nil {
		cfg.Tracer = nopTracer{}
	}
	if cfg.Metrics == nil {
		cfg.Metrics = nopMetrics{}
	}
	if cfg.SanitiseURL == nil {
		cfg.SanitiseURL = func(u string) string { return u }
	}
	return &instrumentedTransport{next: next, cfg: cfg}
}

func (t *instrumentedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	spanName := fmt.Sprintf("%s %s", req.Method, t.cfg.SanitiseURL(req.URL.Path))
	ctx, span := t.cfg.Tracer.Start(req.Context(), spanName)
	defer span.End()

	span.SetAttribute("http.method", req.Method)
	span.SetAttribute("http.url", t.cfg.SanitiseURL(req.URL.String()))
	span.SetAttribute("http.host", req.Host)
	if fid := req.Header.Get("x-fapi-financial-id"); fid != "" {
		span.SetAttribute("obie.financial_id", fid)
	}
	if iid := req.Header.Get("x-fapi-interaction-id"); iid != "" {
		span.SetAttribute("obie.interaction_id", iid)
	}

	start := time.Now()
	resp, err := t.next.RoundTrip(req.WithContext(ctx))
	duration := time.Since(start)

	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode
		span.SetAttribute("http.status_code", strconv.Itoa(statusCode))
	}
	if err != nil {
		span.RecordError(err)
	}

	t.cfg.Metrics.RecordRequest(req.Method, t.cfg.SanitiseURL(req.URL.String()), statusCode, duration, err)
	return resp, err
}

// ────────────────────────────────────────────────────────────────────────────
// Health check
// ────────────────────────────────────────────────────────────────────────────

// HealthStatus represents the health of the OBIE connectivity.
type HealthStatus struct {
	Healthy       bool
	ErrorRate     float64
	AvgDuration   time.Duration
	TotalRequests int
	LastError     error
	CheckedAt     time.Time
}

// HealthChecker evaluates health based on observed metrics.
type HealthChecker struct {
	recorder        *InMemoryRecorder
	maxErrorRate    float64
	maxAvgDuration  time.Duration
}

// NewHealthChecker creates a HealthChecker using the provided recorder.
// maxErrorRate: 0.0–1.0; maxAvgDuration: maximum acceptable average latency.
func NewHealthChecker(recorder *InMemoryRecorder, maxErrorRate float64, maxAvgDuration time.Duration) *HealthChecker {
	return &HealthChecker{
		recorder:       recorder,
		maxErrorRate:   maxErrorRate,
		maxAvgDuration: maxAvgDuration,
	}
}

// Check evaluates and returns the current health status.
func (hc *HealthChecker) Check() HealthStatus {
	records := hc.recorder.Records()
	errorRate := hc.recorder.ErrorRate()
	avgDuration := hc.recorder.AverageDuration()

	var lastErr error
	for i := len(records) - 1; i >= 0; i-- {
		if records[i].Err != nil {
			lastErr = records[i].Err
			break
		}
	}

	healthy := errorRate <= hc.maxErrorRate &&
		(hc.maxAvgDuration == 0 || avgDuration <= hc.maxAvgDuration)

	return HealthStatus{
		Healthy:       healthy,
		ErrorRate:     errorRate,
		AvgDuration:   avgDuration,
		TotalRequests: len(records),
		LastError:     lastErr,
		CheckedAt:     time.Now(),
	}
}
