package obie

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// httpClient is the internal HTTP layer used by all services.
type httpClient struct {
	cfg    *Config
	client *http.Client
	auth   tokenProvider
	log    Logger
}

// tokenProvider abstracts token acquisition so that auth can be swapped in tests.
type tokenProvider interface {
	AccessToken(ctx context.Context) (string, error)
}

// newHTTPClient creates an httpClient wired to the given Config.
func newHTTPClient(cfg *Config, tp tokenProvider) *httpClient {
	hc := cfg.HTTPClient
	if hc == nil {
		hc = &http.Client{Timeout: cfg.Timeout}
	}
	return &httpClient{
		cfg:    cfg,
		client: hc,
		auth:   tp,
		log:    cfg.Logger,
	}
}

// doOptions carries optional per-call settings.
type doOptions struct {
	idempotencyKey string
	jwsSignature   string
	extraHeaders   map[string]string
}

func (h *httpClient) get(ctx context.Context, url string, out any) error {
	return h.do(ctx, http.MethodGet, url, nil, out, doOptions{})
}

func (h *httpClient) post(ctx context.Context, url string, body any, out any, opts doOptions) error {
	return h.do(ctx, http.MethodPost, url, body, out, opts)
}

func (h *httpClient) put(ctx context.Context, url string, body any, out any, opts doOptions) error {
	return h.do(ctx, http.MethodPut, url, body, out, opts)
}

func (h *httpClient) delete(ctx context.Context, url string) error {
	return h.do(ctx, http.MethodDelete, url, nil, nil, doOptions{})
}

// func (h *httpClient) patch(ctx context.Context, url string, body any, out any, opts doOptions) error {
// 	return h.do(ctx, http.MethodPatch, url, body, out, opts)
// }

func (h *httpClient) do(ctx context.Context, method, url string, body any, out any, opts doOptions) error {
	var lastErr error

	for attempt := 0; attempt <= h.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			backoff := h.backoffDuration(attempt)
			h.log.Warnf("obie: retrying %s %s (attempt %d/%d) after %s: %v",
				method, url, attempt, h.cfg.MaxRetries, backoff, lastErr)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		req, err := h.buildRequest(ctx, method, url, body, opts)
		if err != nil {
			return fmt.Errorf("obie: build request: %w", err)
		}

		h.log.Debugf("obie: -> %s %s", method, url)
		resp, err := h.client.Do(req)
		if err != nil {
			lastErr = err
			h.log.Errorf("obie: transport error %s %s: %v", method, url, err)
			continue
		}

		// resp.Body is read fully then closed immediately (cannot defer inside retry loop).
		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close() //nolint:errcheck
		if err != nil {
			return fmt.Errorf("obie: read response: %w", err)
		}

		h.log.Debugf("obie: <- %s %s %d (%d bytes)", method, url, resp.StatusCode, len(respBody))

		for _, hook := range h.cfg.ResponseHooks {
			hook(req, resp)
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			wait := parseRetryAfterHeader(resp.Header.Get("Retry-After"))
			lastErr = fmt.Errorf("obie: rate limited (429), retry after %s", wait)
			h.log.Warnf("obie: 429 on %s %s, waiting %s", method, url, wait)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
			continue
		}

		if resp.StatusCode >= 500 && isIdempotentMethod(method) {
			lastErr = fmt.Errorf("obie: server error %d", resp.StatusCode)
			continue
		}

		if resp.StatusCode >= 400 {
			return parseAPIError(resp, respBody)
		}

		if out != nil && len(respBody) > 0 {
			if err := json.Unmarshal(respBody, out); err != nil {
				return fmt.Errorf("obie: decode response: %w", err)
			}
		}
		return nil
	}
	return fmt.Errorf("obie: all %d retries exhausted: %w", h.cfg.MaxRetries, lastErr)
}

// backoffDuration computes exponential backoff with +-25% jitter.
// Uses crypto/rand for unpredictable retry timing (prevents thundering herd).
func (h *httpClient) backoffDuration(attempt int) time.Duration {
	base := time.Duration(math.Pow(2, float64(attempt-1))) * 500 * time.Millisecond
	if base > 30*time.Second {
		base = 30 * time.Second
	}
	half := int64(base) / 2
	if half < 1 {
		half = 1
	}
	jitter := time.Duration(cryptoRandInt63n(half + 1))
	if cryptoRandBit() {
		return base - jitter/2
	}
	return base + jitter/2
}

// cryptoRandInt63n returns a non-negative pseudo-random int64 in [0,n) using crypto/rand.
func cryptoRandInt63n(n int64) int64 {
	if n <= 0 {
		return 0
	}
	var b [8]byte
	_, _ = rand.Read(b[:])
	var v uint64
	_ = binary.Read(bytes.NewReader(b[:]), binary.BigEndian, &v)
	return int64(v>>1) % n
}

// cryptoRandBit returns a random bool using crypto/rand.
func cryptoRandBit() bool {
	var b [1]byte
	_, _ = rand.Read(b[:])
	return b[0]&1 == 1
}

func (h *httpClient) buildRequest(ctx context.Context, method, url string, body any, opts doOptions) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encode body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	token, err := h.auth.AccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("get access token: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	// ── FAPI Headers (mandatory per Open Banking Security Profile) ────────
	if h.cfg.FinancialID != "" {
		req.Header.Set("x-fapi-financial-id", h.cfg.FinancialID)
	}
	// x-fapi-interaction-id: unique per-request UUID. ASPSPs must echo this in responses.
	req.Header.Set("x-fapi-interaction-id", uuid.New().String())
	// x-fapi-auth-date: date the PSU last logged in to the TPP. RFC1123 format.
	req.Header.Set("x-fapi-auth-date", time.Now().UTC().Format(time.RFC1123))
	// x-fapi-customer-ip-address: PSU's IP (omit for M2M / scheduled payments).
	if h.cfg.CustomerIPAddress != "" {
		req.Header.Set("x-fapi-customer-ip-address", h.cfg.CustomerIPAddress)
	}

	if opts.idempotencyKey != "" {
		req.Header.Set("x-idempotency-key", opts.idempotencyKey)
	}
	if opts.jwsSignature != "" {
		req.Header.Set("x-jws-signature", opts.jwsSignature)
	}
	for k, v := range opts.extraHeaders {
		req.Header.Set(k, v)
	}

	for _, hook := range h.cfg.RequestHooks {
		hook(req)
	}

	return req, nil
}

// parseRetryAfterHeader parses a Retry-After header (delta-seconds or HTTP-date).
func parseRetryAfterHeader(header string) time.Duration {
	if header == "" {
		return time.Second
	}
	if secs, err := strconv.Atoi(header); err == nil {
		return time.Duration(secs) * time.Second
	}
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

func isIdempotentMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodDelete, http.MethodPut:
		return true
	}
	return false
}
