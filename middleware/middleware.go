// Package middleware provides composable HTTP middleware for the OBIE SDK.
// Middleware wraps an http.RoundTripper, allowing cross-cutting concerns
// (logging, metrics, tracing, auth injection) to be layered cleanly.
package middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// RoundTripFunc is a function that implements http.RoundTripper.
type RoundTripFunc func(*http.Request) (*http.Response, error)

// RoundTrip implements http.RoundTripper.
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

// Middleware wraps a RoundTripper to produce a new RoundTripper.
type Middleware func(http.RoundTripper) http.RoundTripper

// Chain composes multiple middleware into a single RoundTripper, applied
// in the order given (first middleware is the outermost wrapper).
//
//	Chain(A, B, C)(base) == A(B(C(base)))
func Chain(base http.RoundTripper, mw ...Middleware) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	// Apply in reverse so first middleware is outermost.
	for i := len(mw) - 1; i >= 0; i-- {
		base = mw[i](base)
	}
	return base
}

// ────────────────────────────────────────────────────────────────────────────
// Logging middleware
// ────────────────────────────────────────────────────────────────────────────

// Logger is the logging interface consumed by LoggingMiddleware.
type Logger interface {
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
}

// LoggingMiddleware logs every outbound request and its response.
func LoggingMiddleware(log Logger) Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripFunc(func(req *http.Request) (*http.Response, error) {
			start := time.Now()
			log.Infof("→ %s %s", req.Method, req.URL)

			resp, err := next.RoundTrip(req)
			elapsed := time.Since(start)

			if err != nil {
				log.Errorf("← %s %s error after %s: %v", req.Method, req.URL, elapsed, err)
				return nil, err
			}
			log.Infof("← %s %s %d (%s)", req.Method, req.URL, resp.StatusCode, elapsed)
			return resp, nil
		})
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Request / response body capture middleware (for debugging / audit)
// ────────────────────────────────────────────────────────────────────────────

// BodyCapture stores captured request/response bodies for a single call.
type BodyCapture struct {
	RequestBody  []byte
	ResponseBody []byte
	StatusCode   int
}

// CapturingMiddleware captures request and response bodies into the provided
// *BodyCapture. Useful for debugging and audit logging.
// Only captures up to maxBytes bytes to avoid OOM on large responses.
func CapturingMiddleware(capture *BodyCapture, maxBytes int64) Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripFunc(func(req *http.Request) (*http.Response, error) {
			// Capture request body.
			if req.Body != nil {
				var buf bytes.Buffer
				tee := io.TeeReader(req.Body, &buf)
				capture.RequestBody, _ = io.ReadAll(io.LimitReader(tee, maxBytes))
				req.Body = io.NopCloser(&buf)
			}

			resp, err := next.RoundTrip(req)
			if err != nil || resp == nil {
				return resp, err
			}

			// Capture response body without consuming it.
			var buf bytes.Buffer
			tee := io.TeeReader(resp.Body, &buf)
			capture.ResponseBody, _ = io.ReadAll(io.LimitReader(tee, maxBytes))
			capture.StatusCode = resp.StatusCode
			resp.Body = io.NopCloser(&buf)
			return resp, nil
		})
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Correlation ID middleware
// ────────────────────────────────────────────────────────────────────────────

type correlationKeyType struct{}

// CorrelationIDKey is the context key for injecting a correlation ID.
var CorrelationIDKey = correlationKeyType{}

// WithCorrelationID stores a correlation ID in ctx for use by the middleware.
func WithCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, CorrelationIDKey, id)
}

// CorrelationIDMiddleware injects an x-correlation-id header. The value is
// read from ctx if present, otherwise a new UUID is generated.
func CorrelationIDMiddleware(generate func() string) Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripFunc(func(req *http.Request) (*http.Response, error) {
			id, _ := req.Context().Value(CorrelationIDKey).(string)
			if id == "" {
				id = generate()
			}
			req = req.Clone(req.Context())
			req.Header.Set("x-correlation-id", id)
			return next.RoundTrip(req)
		})
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Header injection middleware
// ────────────────────────────────────────────────────────────────────────────

// HeadersMiddleware injects a fixed set of headers into every request.
func HeadersMiddleware(headers map[string]string) Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripFunc(func(req *http.Request) (*http.Response, error) {
			req = req.Clone(req.Context())
			for k, v := range headers {
				req.Header.Set(k, v)
			}
			return next.RoundTrip(req)
		})
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Timeout middleware
// ────────────────────────────────────────────────────────────────────────────

// TimeoutMiddleware applies a per-request deadline, overriding any existing
// deadline if the provided timeout is shorter.
func TimeoutMiddleware(timeout time.Duration) Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripFunc(func(req *http.Request) (*http.Response, error) {
			ctx, cancel := context.WithTimeout(req.Context(), timeout)
			defer cancel()
			return next.RoundTrip(req.WithContext(ctx))
		})
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Dry-run middleware
// ────────────────────────────────────────────────────────────────────────────

// DryRunMiddleware intercepts all state-mutating methods (POST/PUT/PATCH/DELETE)
// and returns a synthetic 200 OK without forwarding the request. Useful for
// testing consent/payment flows without hitting real endpoints.
func DryRunMiddleware() Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.Method {
			case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
				body := fmt.Sprintf(`{"_dryRun":true,"method":%q,"url":%q}`,
					req.Method, req.URL.String())
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(bytes.NewBufferString(body)),
					Request:    req,
				}, nil
			}
			return next.RoundTrip(req)
		})
	}
}
