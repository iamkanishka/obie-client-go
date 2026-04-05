// Package idempotency provides an idempotency-key store that enforces
// exactly-once semantics for payment and consent submissions.
//
// OBIE mandates that POST requests to payment endpoints include an
// x-idempotency-key. The store records keys with their response payloads so
// that duplicate requests can be detected and the original response returned
// without re-executing the operation.
package idempotency

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Status represents the current state of an idempotent operation.
type Status string

const (
	StatusPending   Status = "pending"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

// Record stores the state and response for an idempotent operation.
type Record struct {
	Key        string
	Status     Status
	StatusCode int
	Response   json.RawMessage
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ExpiresAt  time.Time
}

// ErrDuplicateKey is returned when an operation with the same key is already in the Pending state.
type ErrDuplicateKey struct {
	Key    string
	Status Status
}

func (e *ErrDuplicateKey) Error() string {
	return fmt.Sprintf("idempotency: duplicate key %q (current status: %s)", e.Key, e.Status)
}

// Store is a thread-safe in-memory idempotency key store.
type Store struct {
	mu      sync.RWMutex
	records map[string]*Record
	ttl     time.Duration
}

// NewStore creates a Store with the given TTL for record retention.
func NewStore(ttl time.Duration) *Store {
	s := &Store{
		records: make(map[string]*Record),
		ttl:     ttl,
	}
	go s.runEviction()
	return s
}

// Begin registers a new idempotency key in the Pending state.
func (s *Store) Begin(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if rec, ok := s.records[key]; ok {
		if time.Now().Before(rec.ExpiresAt) {
			return &ErrDuplicateKey{Key: key, Status: rec.Status}
		}
		delete(s.records, key)
	}

	s.records[key] = &Record{
		Key:       key,
		Status:    StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ExpiresAt: time.Now().Add(s.ttl),
	}
	return nil
}

// Complete marks the operation as successfully completed and stores the response.
func (s *Store) Complete(key string, statusCode int, response json.RawMessage) error {
	return s.update(key, StatusCompleted, statusCode, response)
}

// Fail marks the operation as permanently failed.
func (s *Store) Fail(key string, statusCode int, response json.RawMessage) error {
	return s.update(key, StatusFailed, statusCode, response)
}

// Get retrieves the record for key, if present and unexpired.
func (s *Store) Get(key string) (*Record, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rec, ok := s.records[key]
	if !ok || time.Now().After(rec.ExpiresAt) {
		return nil, false
	}
	cp := *rec
	return &cp, true
}

// Delete removes the record for key.
func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.records, key)
}

func (s *Store) update(key string, status Status, statusCode int, response json.RawMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.records[key]
	if !ok {
		return fmt.Errorf("idempotency: key %q not found", key)
	}
	rec.Status = status
	rec.StatusCode = statusCode
	rec.Response = response
	rec.UpdatedAt = time.Now()
	rec.ExpiresAt = time.Now().Add(s.ttl)
	return nil
}

func (s *Store) runEviction() {
	ticker := time.NewTicker(s.ttl)
	defer ticker.Stop()
	for range ticker.C {
		s.evict()
	}
}

func (s *Store) evict() {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, rec := range s.records {
		if now.After(rec.ExpiresAt) {
			delete(s.records, k)
		}
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Server-side HTTP middleware
// ────────────────────────────────────────────────────────────────────────────

// Middleware returns an http.Handler middleware that enforces idempotency
// using the x-idempotency-key header. On the first request with a given key,
// the operation proceeds normally and the response is stored. Subsequent
// requests with the same key receive the stored response immediately without
// re-executing the handler.
func Middleware(store *Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("x-idempotency-key")
			if key == "" || r.Method == http.MethodGet {
				next.ServeHTTP(w, r)
				return
			}

			if rec, ok := store.Get(key); ok && rec.Status == StatusCompleted {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("x-idempotency-replayed", "true")
				w.WriteHeader(rec.StatusCode)
				w.Write(rec.Response) //nolint:errcheck
				return
			}

			if err := store.Begin(key); err != nil {
				http.Error(w, err.Error(), http.StatusConflict)
				return
			}

			rw := &capturingWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rw, r)

			if rw.statusCode >= 200 && rw.statusCode < 300 {
				store.Complete(key, rw.statusCode, rw.body) //nolint:errcheck
			} else {
				store.Fail(key, rw.statusCode, rw.body) //nolint:errcheck
			}
		})
	}
}

type capturingWriter struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (cw *capturingWriter) WriteHeader(code int) {
	cw.statusCode = code
	cw.ResponseWriter.WriteHeader(code)
}

func (cw *capturingWriter) Write(b []byte) (int, error) {
	cw.body = append(cw.body, b...)
	return cw.ResponseWriter.Write(b)
}
