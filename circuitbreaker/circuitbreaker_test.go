package circuitbreaker_test

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/iamkanishka/obie-client-go/circuitbreaker"
)


// roundTripFunc adapts a function to implement http.RoundTripper.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func serverErrorResp() *http.Response {
	return &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(strings.NewReader("")),
		Header:     make(http.Header),
	}
}

func okResp() *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("{}")),
		Header:     make(http.Header),
	}
}

func TestCircuitBreaker_ClosedToOpen(t *testing.T) {
	cb := circuitbreaker.New(circuitbreaker.Config{
		MaxFailures: 3,
		OpenTimeout: time.Hour, // don't auto-recover during test
	})

	calls := 0
	base := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		return serverErrorResp(), nil
	})
	transport := cb.Middleware()(base)
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)

	for i := 0; i < 3; i++ {
		transport.RoundTrip(req) //nolint:errcheck
	}

	if cb.State() != circuitbreaker.StateOpen {
		t.Errorf("expected Open after %d failures, got %s", 3, cb.State())
	}

	// Next call should be rejected without hitting base.
	callsBefore := calls
	_, err := transport.RoundTrip(req)
	if err == nil {
		t.Error("expected error from open circuit, got nil")
	}
	var cbErr *circuitbreaker.ErrCircuitOpen
	if !errors.As(err, &cbErr) {
		t.Errorf("expected ErrCircuitOpen, got %T", err)
	}
	if calls != callsBefore {
		t.Error("base should not be called when circuit is open")
	}
}

func TestCircuitBreaker_HalfOpen_SuccessCloses(t *testing.T) {
	cb := circuitbreaker.New(circuitbreaker.Config{
		MaxFailures:      2,
		OpenTimeout:      1 * time.Millisecond,
		SuccessThreshold: 2,
	})

	calls := 0
	base := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		if calls <= 2 {
			return serverErrorResp(), nil // initial failures
		}
		return okResp(), nil
	})
	transport := cb.Middleware()(base)
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)

	// Trigger Open.
	transport.RoundTrip(req) //nolint:errcheck
	transport.RoundTrip(req) //nolint:errcheck

	if cb.State() != circuitbreaker.StateOpen {
		t.Fatalf("expected Open, got %s", cb.State())
	}

	// Wait for open timeout to expire.
	time.Sleep(5 * time.Millisecond)

	// Two successful probes should close the circuit.
	for i := 0; i < 2; i++ {
		resp, err := transport.RoundTrip(req)
		if err != nil {
			t.Fatalf("probe %d: unexpected error: %v", i, err)
		}
		resp.Body.Close() //nolint:errcheck // inside loop, cannot defer
	}

	if cb.State() != circuitbreaker.StateClosed {
		t.Errorf("expected Closed after successful probes, got %s", cb.State())
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cb := circuitbreaker.New(circuitbreaker.Config{
		MaxFailures: 1,
		OpenTimeout: time.Hour,
	})

	base := roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return nil, errors.New("transport error")
	})
	transport := cb.Middleware()(base)
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	transport.RoundTrip(req) //nolint:errcheck

	if cb.State() != circuitbreaker.StateOpen {
		t.Fatalf("expected Open, got %s", cb.State())
	}

	cb.Reset()

	if cb.State() != circuitbreaker.StateClosed {
		t.Errorf("after Reset expected Closed, got %s", cb.State())
	}
}

func TestCircuitBreaker_OnStateChange(t *testing.T) {
	var transitions []string
	cb := circuitbreaker.New(circuitbreaker.Config{
		MaxFailures: 2,
		OpenTimeout: 1 * time.Millisecond,
		SuccessThreshold: 1,
		OnStateChange: func(from, to circuitbreaker.State) {
			transitions = append(transitions, from.String()+"→"+to.String())
		},
	})

	base := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return serverErrorResp(), nil
	})
	transport := cb.Middleware()(base)
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)

	transport.RoundTrip(req) //nolint:errcheck
	transport.RoundTrip(req) //nolint:errcheck

	if len(transitions) == 0 {
		t.Error("expected at least one state transition event")
	}
	if transitions[0] != "Closed→Open" {
		t.Errorf("first transition: got %q, want %q", transitions[0], "Closed→Open")
	}
}
