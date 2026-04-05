package ratelimit_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/iamkanishka/obie-client-go/ratelimit"
)


// roundTripFunc adapts a function to implement http.RoundTripper.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestLimiter_TokensAvailableAtStart(t *testing.T) {
	l := ratelimit.NewLimiter(10, 10)
	if avail := l.Available(); avail < 9.9 {
		t.Errorf("expected ~10 tokens at start, got %.2f", avail)
	}
}

func TestLimiter_TryAcquire_Drains(t *testing.T) {
	l := ratelimit.NewLimiter(1, 3)
	for i := 0; i < 3; i++ {
		if !l.TryAcquire() {
			t.Fatalf("TryAcquire failed on attempt %d (expected success)", i+1)
		}
	}
	if l.TryAcquire() {
		t.Error("expected TryAcquire to fail when bucket is empty")
	}
}

func TestLimiter_Wait_RespectsContext(t *testing.T) {
	// Bucket immediately empty.
	l := ratelimit.NewLimiter(0.001, 0) // near-zero refill rate
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := l.Wait(ctx)
	if err == nil {
		t.Error("expected error when context is cancelled, got nil")
	}
}

func TestMiddleware_Passes200Through(t *testing.T) {
	l := ratelimit.NewLimiter(100, 100)
	var calls int32
	base := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		atomic.AddInt32(&calls, 1)
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader("{}")),
			Header:     make(http.Header),
		}, nil
	})

	transport := ratelimit.Middleware(l, 2)(base)
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("status: got %d, want 200", resp.StatusCode)
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Errorf("calls: got %d, want 1", calls)
	}
}

func TestMiddleware_Retries429(t *testing.T) {
	l := ratelimit.NewLimiter(100, 100)
	var calls int32
	base := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		n := atomic.AddInt32(&calls, 1)
		status := 429
		if n >= 3 {
			status = 200
		}
		hdr := make(http.Header)
		if status == 429 {
			hdr.Set("Retry-After", "0") // instant retry in tests
		}
		return &http.Response{
			StatusCode: status,
			Body:       io.NopCloser(strings.NewReader("{}")),
			Header:     hdr,
		}, nil
	})

	transport := ratelimit.Middleware(l, 3)(base)
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("final status: got %d, want 200", resp.StatusCode)
	}
	if atomic.LoadInt32(&calls) != 3 {
		t.Errorf("calls: got %d, want 3", calls)
	}
}
