package observability_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/iamkanishka/obie-client-go/observability"
)


// roundTripFunc adapts a function to implement http.RoundTripper.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// recordingSpan captures attribute and error calls for assertions.
type recordingSpan struct {
	attrs  map[string]any
	errors []error
	ended  bool
}

func newRecordingSpan() *recordingSpan {
	return &recordingSpan{attrs: make(map[string]any)}
}

func (s *recordingSpan) SetAttribute(k string, v any) { s.attrs[k] = v }
func (s *recordingSpan) RecordError(err error)               { s.errors = append(s.errors, err) }
func (s *recordingSpan) End()                                { s.ended = true }

// recordingTracer creates recordingSpans and stores them for inspection.
type recordingTracer struct {
	spans []*recordingSpan
}

func (t *recordingTracer) Start(ctx context.Context, name string) (context.Context, observability.Span) {
	span := newRecordingSpan()
	t.spans = append(t.spans, span)
	return ctx, span
}

func TestInstrumentedTransport_RecordsSpan(t *testing.T) {
	tracer := &recordingTracer{}
	recorder := observability.NewInMemoryRecorder()

	base := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader("{}")),
			Header:     make(http.Header),
		}, nil
	})

	transport := observability.NewInstrumentedTransport(base, observability.TransportConfig{
		ComponentName: "test",
		Tracer:        tracer,
		Metrics:       recorder,
	})

	req, _ := http.NewRequest(http.MethodGet, "http://example.com/accounts", nil)
	req.Header.Set("x-fapi-interaction-id", "iid-123")
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	defer resp.Body.Close()

	if len(tracer.spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(tracer.spans))
	}
	span := tracer.spans[0]
	if !span.ended {
		t.Error("expected span to be ended")
	}
	if span.attrs["http.method"] != "GET" {
		t.Errorf("http.method: got %v, want GET", span.attrs["http.method"])
	}
	if span.attrs["obie.interaction_id"] != "iid-123" {
		t.Errorf("obie.interaction_id: got %v, want iid-123", span.attrs["obie.interaction_id"])
	}
}

func TestInstrumentedTransport_RecordsError(t *testing.T) {
	tracer := &recordingTracer{}
	recorder := observability.NewInMemoryRecorder()
	expectedErr := errors.New("network failure")

	base := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, expectedErr
	})

	transport := observability.NewInstrumentedTransport(base, observability.TransportConfig{
		Tracer:  tracer,
		Metrics: recorder,
	})

	req, _ := http.NewRequest(http.MethodGet, "http://example.com/payments", nil)
	transport.RoundTrip(req) //nolint:errcheck

	if len(tracer.spans[0].errors) == 0 {
		t.Error("expected error to be recorded on span")
	}
}

func TestInMemoryRecorder_ErrorRate(t *testing.T) {
	rec := observability.NewInMemoryRecorder()
	rec.RecordRequest("GET", "/accounts", 200, 50*time.Millisecond, nil)
	rec.RecordRequest("GET", "/accounts", 200, 60*time.Millisecond, nil)
	rec.RecordRequest("GET", "/accounts", 500, 70*time.Millisecond, nil)
	rec.RecordRequest("GET", "/accounts", 200, 40*time.Millisecond, nil)

	rate := rec.ErrorRate()
	if rate < 0.24 || rate > 0.26 {
		t.Errorf("error rate: got %.2f, want ~0.25", rate)
	}
}

func TestInMemoryRecorder_AverageDuration(t *testing.T) {
	rec := observability.NewInMemoryRecorder()
	rec.RecordRequest("GET", "/accounts", 200, 100*time.Millisecond, nil)
	rec.RecordRequest("GET", "/accounts", 200, 200*time.Millisecond, nil)

	avg := rec.AverageDuration()
	if avg != 150*time.Millisecond {
		t.Errorf("avg duration: got %v, want 150ms", avg)
	}
}

func TestHealthChecker_Healthy(t *testing.T) {
	rec := observability.NewInMemoryRecorder()
	rec.RecordRequest("GET", "/accounts", 200, 20*time.Millisecond, nil)
	rec.RecordRequest("GET", "/accounts", 200, 30*time.Millisecond, nil)

	hc := observability.NewHealthChecker(rec, 0.1, time.Second)
	status := hc.Check()

	if !status.Healthy {
		t.Errorf("expected Healthy=true, got false; error rate=%.2f, avg=%v",
			status.ErrorRate, status.AvgDuration)
	}
}

func TestHealthChecker_UnhealthyHighErrorRate(t *testing.T) {
	rec := observability.NewInMemoryRecorder()
	rec.RecordRequest("GET", "/accounts", 500, 20*time.Millisecond, nil)
	rec.RecordRequest("GET", "/accounts", 500, 20*time.Millisecond, nil)

	hc := observability.NewHealthChecker(rec, 0.1, time.Second)
	status := hc.Check()

	if status.Healthy {
		t.Error("expected Healthy=false for high error rate")
	}
}

func TestInMemoryRecorder_Flush(t *testing.T) {
	rec := observability.NewInMemoryRecorder()
	rec.RecordRequest("GET", "/x", 200, time.Millisecond, nil)
	rec.Flush()
	if len(rec.Records()) != 0 {
		t.Error("expected 0 records after Flush")
	}
}
