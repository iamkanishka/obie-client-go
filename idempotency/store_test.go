package idempotency_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/iamkanishka/obie-client-go/idempotency"
)

func TestStore_BeginAndGet(t *testing.T) {
	s := idempotency.NewStore(time.Minute)

	if err := s.Begin("key-1"); err != nil {
		t.Fatalf("Begin: %v", err)
	}

	rec, ok := s.Get("key-1")
	if !ok {
		t.Fatal("expected record after Begin")
	}
	if rec.Status != idempotency.StatusPending {
		t.Errorf("Status: got %s, want Pending", rec.Status)
	}
}

func TestStore_DuplicateKeyReturnsError(t *testing.T) {
	s := idempotency.NewStore(time.Minute)
	s.Begin("dup-key") //nolint:errcheck

	err := s.Begin("dup-key")
	if err == nil {
		t.Fatal("expected ErrDuplicateKey, got nil")
	}
	var dupErr *idempotency.ErrDuplicateKey
	if !errors.As(err, &dupErr) {
		t.Errorf("expected *ErrDuplicateKey, got %T", err)
	}
	if dupErr.Key != "dup-key" {
		t.Errorf("Key: got %q, want %q", dupErr.Key, "dup-key")
	}
}

func TestStore_CompleteAndGet(t *testing.T) {
	s := idempotency.NewStore(time.Minute)
	s.Begin("pay-1") //nolint:errcheck

	payload := json.RawMessage(`{"DomesticPaymentId":"pay-1","Status":"Pending"}`)
	if err := s.Complete("pay-1", 201, payload); err != nil {
		t.Fatalf("Complete: %v", err)
	}

	rec, ok := s.Get("pay-1")
	if !ok {
		t.Fatal("expected record after Complete")
	}
	if rec.Status != idempotency.StatusCompleted {
		t.Errorf("Status: got %s, want Completed", rec.Status)
	}
	if rec.StatusCode != 201 {
		t.Errorf("StatusCode: got %d, want 201", rec.StatusCode)
	}
}

func TestStore_ExpiredRecordCanBeReused(t *testing.T) {
	s := idempotency.NewStore(10 * time.Millisecond)
	s.Begin("short-key") //nolint:errcheck

	time.Sleep(20 * time.Millisecond)

	// Should succeed since the old record is expired.
	if err := s.Begin("short-key"); err != nil {
		t.Errorf("expected Begin to succeed after expiry, got: %v", err)
	}
}

func TestStore_Delete(t *testing.T) {
	s := idempotency.NewStore(time.Minute)
	s.Begin("del-key") //nolint:errcheck
	s.Delete("del-key")

	_, ok := s.Get("del-key")
	if ok {
		t.Error("expected record to be absent after Delete")
	}
}

func TestMiddleware_ReplaysCachedResponse(t *testing.T) {
	store := idempotency.NewStore(time.Minute)
	callCount := 0

	handler := idempotency.Middleware(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		if _, err := w.Write([]byte(`{"id":"pay-123"}`)); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))

	makeRequest := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/payments", nil)
		req.Header.Set("x-idempotency-key", "idem-abc")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		return rr
	}

	first := makeRequest()
	if first.Code != http.StatusCreated {
		t.Errorf("first response: got %d, want 201", first.Code)
	}

	second := makeRequest()
	if second.Code != http.StatusCreated {
		t.Errorf("second (replayed) response: got %d, want 201", second.Code)
	}
	if second.Header().Get("x-idempotency-replayed") != "true" {
		t.Error("expected x-idempotency-replayed header on second request")
	}

	// Handler must only have been called once.
	if callCount != 1 {
		t.Errorf("handler call count: got %d, want 1", callCount)
	}
}

func TestMiddleware_GETPassesThrough(t *testing.T) {
	store := idempotency.NewStore(time.Minute)
	callCount := 0

	handler := idempotency.Middleware(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/accounts", nil)
		req.Header.Set("x-idempotency-key", "same-key")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	if callCount != 3 {
		t.Errorf("GET requests should not be deduplicated; call count: %d, want 3", callCount)
	}
}
