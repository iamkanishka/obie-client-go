package middleware_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/iamkanishka/obie-client-go/middleware"
)

// roundTripFunc satisfies http.RoundTripper for test use.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func okResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestChain_OrderPreserved(t *testing.T) {
	var order []string
	makeMiddleware := func(name string) middleware.Middleware {
		return func(next http.RoundTripper) http.RoundTripper {
			return roundTripFunc(func(req *http.Request) (*http.Response, error) {
				order = append(order, name+"-before")
				resp, err := next.RoundTrip(req)
				order = append(order, name+"-after")
				return resp, err
			})
		}
	}

	base := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		order = append(order, "base")
		return okResponse("{}"), nil
	})

	transport := middleware.Chain(base, makeMiddleware("A"), makeMiddleware("B"))
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	transport.RoundTrip(req) //nolint:errcheck

	want := []string{"A-before", "B-before", "base", "B-after", "A-after"}
	for i, got := range order {
		if i >= len(want) || got != want[i] {
			t.Errorf("chain order[%d]: got %q, want %q", i, got, want[i])
		}
	}
}

type stubLog struct{ lines []string }

func (s *stubLog) Infof(f string, a ...any)  { s.lines = append(s.lines, "INFO") }
func (s *stubLog) Warnf(f string, a ...any)  { s.lines = append(s.lines, "WARN") }
func (s *stubLog) Errorf(f string, a ...any) { s.lines = append(s.lines, "ERROR") }

func TestLoggingMiddleware_LogsRequest(t *testing.T) {
	log := &stubLog{}
	base := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return okResponse("{}"), nil
	})
	transport := middleware.Chain(base, middleware.LoggingMiddleware(log))
	req, _ := http.NewRequest(http.MethodGet, "http://example.com/test", nil)
	transport.RoundTrip(req) //nolint:errcheck

	if len(log.lines) < 2 {
		t.Errorf("expected at least 2 log lines, got %d", len(log.lines))
	}
}

func TestCapturingMiddleware_CapturesBody(t *testing.T) {
	capture := &middleware.BodyCapture{}
	base := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return okResponse(`{"ok":true}`), nil
	})
	transport := middleware.Chain(base, middleware.CapturingMiddleware(capture, 1024))

	body := `{"payment":"123"}`
	req, _ := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewBufferString(body))
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	defer resp.Body.Close()

	if string(capture.RequestBody) != body {
		t.Errorf("RequestBody: got %q, want %q", capture.RequestBody, body)
	}
	if !strings.Contains(string(capture.ResponseBody), "ok") {
		t.Errorf("ResponseBody missing expected content: %q", capture.ResponseBody)
	}
	if capture.StatusCode != http.StatusOK {
		t.Errorf("StatusCode: got %d, want 200", capture.StatusCode)
	}
}

func TestCorrelationIDMiddleware_InjectsHeader(t *testing.T) {
	var gotID string
	base := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		gotID = req.Header.Get("x-correlation-id")
		return okResponse("{}"), nil
	})
	transport := middleware.Chain(base,
		middleware.CorrelationIDMiddleware(func() string { return "generated-id" }))

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	ctx := middleware.WithCorrelationID(req.Context(), "my-trace-id")
	transport.RoundTrip(req.WithContext(ctx)) //nolint:errcheck

	if gotID != "my-trace-id" {
		t.Errorf("x-correlation-id: got %q, want %q", gotID, "my-trace-id")
	}
}

func TestCorrelationIDMiddleware_GeneratesWhenMissing(t *testing.T) {
	var gotID string
	base := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		gotID = req.Header.Get("x-correlation-id")
		return okResponse("{}"), nil
	})
	transport := middleware.Chain(base,
		middleware.CorrelationIDMiddleware(func() string { return "auto-id" }))

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	transport.RoundTrip(req) //nolint:errcheck

	if gotID != "auto-id" {
		t.Errorf("generated correlation ID: got %q, want %q", gotID, "auto-id")
	}
}

func TestDryRunMiddleware_BlocksMutatingMethods(t *testing.T) {
	called := false
	base := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		called = true
		return okResponse("{}"), nil
	})
	transport := middleware.Chain(base, middleware.DryRunMiddleware())

	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
		called = false
		req, _ := http.NewRequest(method, "http://example.com", nil)
		resp, err := transport.RoundTrip(req)
		if err != nil {
			t.Fatalf("%s dry-run: %v", method, err)
		}
		if called {
			t.Errorf("%s: expected base NOT to be called in dry-run mode", method)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("%s: expected 200, got %d", method, resp.StatusCode)
		}
	}
}

func TestDryRunMiddleware_PassesGetThrough(t *testing.T) {
	called := false
	base := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		called = true
		return okResponse("{}"), nil
	})
	transport := middleware.Chain(base, middleware.DryRunMiddleware())

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	transport.RoundTrip(req) //nolint:errcheck

	if !called {
		t.Error("GET should pass through dry-run middleware")
	}
}

func TestTimeoutMiddleware_CancelsContext(t *testing.T) {
	var ctxDeadline bool
	base := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		_, ctxDeadline = req.Context().Deadline()
		return okResponse("{}"), nil
	})
	transport := middleware.Chain(base, middleware.TimeoutMiddleware(5000000000)) // 5s

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	transport.RoundTrip(req.WithContext(context.Background())) //nolint:errcheck

	if !ctxDeadline {
		t.Error("expected request context to have a deadline set by TimeoutMiddleware")
	}
}
