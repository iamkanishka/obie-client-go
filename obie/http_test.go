package obie

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// staticToken is a tokenProvider that always returns the same token.
type staticToken struct{ tok string }

func (s staticToken) AccessToken(_ context.Context) (string, error) { return s.tok, nil }

func newTestHTTPClient(t *testing.T, srv *httptest.Server) *httpClient {
	t.Helper()
	cfg := &Config{
		FinancialID: "test-fid",
		Timeout:     5 * time.Second,
		MaxRetries:  2,
		Logger:      nopLogger{},
		HTTPClient:  srv.Client(),
	}
	cfg.defaults()
	return newHTTPClient(cfg, staticToken{tok: "test-token"})
}

func TestHTTPClient_Get_Success(t *testing.T) {
	type response struct {
		Value string `json:"value"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("missing or wrong Authorization header: %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("x-fapi-financial-id") != "test-fid" {
			t.Errorf("missing x-fapi-financial-id header")
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response{Value: "hello"}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	hc := newTestHTTPClient(t, srv)

	var out response
	if err := hc.get(context.Background(), srv.URL+"/test", &out); err != nil {
		t.Fatalf("get: %v", err)
	}
	if out.Value != "hello" {
		t.Errorf("got %q, want %q", out.Value, "hello")
	}
}

func TestHTTPClient_Post_Success(t *testing.T) {
	type request struct{ Foo string }
	type response struct{ Bar string }

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.Foo != "baz" {
			t.Errorf("request body Foo: got %q, want %q", req.Foo, "baz")
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response{Bar: "qux"}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	hc := newTestHTTPClient(t, srv)

	var out response
	err := hc.post(context.Background(), srv.URL+"/test",
		request{Foo: "baz"}, &out, doOptions{idempotencyKey: "idem-123"})
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	if out.Bar != "qux" {
		t.Errorf("got %q, want %q", out.Bar, "qux")
	}
}

func TestHTTPClient_Get_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		if err := json.NewEncoder(w).Encode(map[string]any{
			"Code":    "UK.OBIE.Resource.NotFound",
			"Message": "Unauthorized",
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	hc := newTestHTTPClient(t, srv)
	hc.cfg.MaxRetries = 0 // disable retries for this test

	var out any
	err := hc.get(context.Background(), srv.URL+"/fail", &out)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("StatusCode: got %d, want %d", apiErr.StatusCode, http.StatusUnauthorized)
	}
}

func TestHTTPClient_IdempotencyKey_Injected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("x-idempotency-key")
		if key != "my-idem-key" {
			t.Errorf("x-idempotency-key: got %q, want %q", key, "my-idem-key")
		}

		w.WriteHeader(http.StatusOK)

		if _, err := w.Write([]byte("{}")); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer srv.Close()

	hc := newTestHTTPClient(t, srv)
	var out any
	hc.post(context.Background(), srv.URL+"/pay", nil, &out, //nolint:errcheck
		doOptions{idempotencyKey: "my-idem-key"})
}

func TestHTTPClient_JWSSignature_Injected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sig := r.Header.Get("x-jws-signature")
		if sig != "header..signature" {
			t.Errorf("x-jws-signature: got %q, want %q", sig, "header..signature")
		}

		w.WriteHeader(http.StatusOK)

		if _, err := w.Write([]byte("{}")); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer srv.Close()

	hc := newTestHTTPClient(t, srv)
	var out any
	hc.post(context.Background(), srv.URL+"/pay", nil, &out, //nolint:errcheck
		doOptions{jwsSignature: "header..signature"})
}

func TestHTTPClient_RetryOnError(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		w.WriteHeader(http.StatusOK)

		if _, err := w.Write([]byte(`{"ok":true}`)); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer srv.Close()

	hc := newTestHTTPClient(t, srv)
	hc.cfg.MaxRetries = 3

	var out map[string]bool
	if err := hc.get(context.Background(), srv.URL+"/rate", &out); err != nil {
		t.Fatalf("get with retry: %v", err)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}
