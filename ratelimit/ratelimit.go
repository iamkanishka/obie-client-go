// Package ratelimit provides a token-bucket rate limiter and an HTTP middleware
// that respects Retry-After headers returned by OBIE ASPSPs.
package ratelimit

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// ────────────────────────────────────────────────────────────────────────────
// Token-bucket limiter
// ────────────────────────────────────────────────────────────────────────────

// Limiter is a thread-safe token-bucket rate limiter.
type Limiter struct {
	mu       sync.Mutex
	rate     float64   // tokens added per second
	burst    float64   // maximum token capacity
	tokens   float64   // current token count
	lastTime time.Time // last refill timestamp
}

// NewLimiter creates a Limiter allowing rate requests per second with a burst
// capacity of burst. The bucket starts full.
func NewLimiter(rate, burst float64) *Limiter {
	return &Limiter{
		rate:     rate,
		burst:    burst,
		tokens:   burst,
		lastTime: time.Now(),
	}
}

// Wait blocks until a token is available or ctx is cancelled.
func (l *Limiter) Wait(ctx context.Context) error {
	for {
		l.mu.Lock()
		l.refill()
		if l.tokens >= 1 {
			l.tokens--
			l.mu.Unlock()
			return nil
		}
		// Calculate wait for next token.
		wait := time.Duration((1-l.tokens)/l.rate*1e9) * time.Nanosecond
		l.mu.Unlock()

		select {
		case <-ctx.Done():
			return fmt.Errorf("ratelimit: context cancelled while waiting: %w", ctx.Err())
		case <-time.After(wait):
		}
	}
}

// TryAcquire attempts to acquire a token without blocking.
// Returns true if a token was acquired, false if the bucket is empty.
func (l *Limiter) TryAcquire() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.refill()
	if l.tokens >= 1 {
		l.tokens--
		return true
	}
	return false
}

// Available returns the approximate number of currently available tokens.
func (l *Limiter) Available() float64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.refill()
	return l.tokens
}

func (l *Limiter) refill() {
	now := time.Now()
	elapsed := now.Sub(l.lastTime).Seconds()
	l.tokens = min64(l.burst, l.tokens+elapsed*l.rate)
	l.lastTime = now
}

func min64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// ────────────────────────────────────────────────────────────────────────────
// HTTP middleware
// ────────────────────────────────────────────────────────────────────────────

// Middleware returns an http.RoundTripper that:
//  1. Enforces the token-bucket limit before forwarding the request.
//  2. On 429 responses, reads the Retry-After header (seconds or HTTP-date)
//     and waits the prescribed duration before the next attempt.
//  3. Retries up to maxRetries times.
func Middleware(limiter *Limiter, maxRetries int) func(http.RoundTripper) http.RoundTripper {
	return func(next http.RoundTripper) http.RoundTripper {
		return &rateLimitedTransport{
			next:       next,
			limiter:    limiter,
			maxRetries: maxRetries,
		}
	}
}

type rateLimitedTransport struct {
	next       http.RoundTripper
	limiter    *Limiter
	maxRetries int
}

func (t *rateLimitedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Wait for a token before the first attempt.
	if err := t.limiter.Wait(req.Context()); err != nil {
		return nil, err
	}

	var (
		resp    *http.Response
		err     error
		attempt int
	)

	for attempt = 0; attempt <= t.maxRetries; attempt++ {
		resp, err = t.next.RoundTrip(req)
		if err != nil || resp == nil {
			return resp, err
		}

		if resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		}

		// 429 — parse Retry-After, drain and close body to reuse the connection.
		wait := parseRetryAfter(resp.Header.Get("Retry-After"))
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close() //nolint:errcheck

		if attempt < t.maxRetries {
			select {
			case <-req.Context().Done():
				return nil, fmt.Errorf("ratelimit: context cancelled during 429 backoff: %w", req.Context().Err())
			case <-time.After(wait):
			}
			// Re-acquire token for the retry.
			if err := t.limiter.Wait(req.Context()); err != nil {
				return nil, err
			}
		}
	}

	return nil, fmt.Errorf("ratelimit: all %d retries exhausted after 429 responses", t.maxRetries)
}

// parseRetryAfter parses the value of a Retry-After header.
// Supports both delta-seconds and HTTP-date formats.
func parseRetryAfter(header string) time.Duration {
	if header == "" {
		return time.Second // default 1 s
	}
	// Delta-seconds format.
	if secs, err := strconv.Atoi(header); err == nil {
		return time.Duration(secs) * time.Second
	}
	// HTTP-date format.
	for _, layout := range []string{time.RFC1123, time.RFC850, time.ANSIC} {
		if t, err := time.Parse(layout, header); err == nil {
			d := time.Until(t)
			if d < 0 {
				return 0
			}
			return d
		}
	}
	return time.Second
}
