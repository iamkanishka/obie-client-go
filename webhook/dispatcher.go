// Package webhook extends the base events package with a typed dispatcher
// that routes OBIE event notifications to strongly-typed handlers, supports
// multiple subscribers per event type, and provides dead-letter queue (DLQ)
// semantics for failed handlers.
package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ────────────────────────────────────────────────────────────────────────────
// Event types and envelope
// ────────────────────────────────────────────────────────────────────────────

// EventType is the URN identifying an OBIE event type.
type EventType string

const (
	EventTypeResourceUpdate                          EventType = "urn:uk:org:openbanking:events:resource-update"
	EventTypeConsentAuthRevoked                      EventType = "urn:uk:org:openbanking:events:consent-authorization-revoked"
	EventTypeAccountAccessConsentLinkedAccountUpdate EventType = "urn:uk:org:openbanking:events:account-access-consent-linked-account-update"
)

// Envelope is the top-level OBIE event notification structure.
type Envelope struct {
	Iss    string                     `json:"iss"`
	Iat    int64                      `json:"iat"`
	Jti    string                     `json:"jti"`
	Aud    string                     `json:"aud"`
	Sub    string                     `json:"sub"`
	Txn    string                     `json:"txn"`
	Toe    int64                      `json:"toe"`
	Events map[EventType]json.RawMessage `json:"events"`
}

// ReceivedAt returns the time-of-event as a time.Time.
func (e *Envelope) ReceivedAt() time.Time {
	return time.Unix(e.Toe, 0)
}

// ────────────────────────────────────────────────────────────────────────────
// Typed event payloads
// ────────────────────────────────────────────────────────────────────────────

// ResourceUpdateEvent is the payload for EventTypeResourceUpdate.
type ResourceUpdateEvent struct {
	Subject ResourceUpdateSubject `json:"subject"`
}

// ResourceUpdateSubject describes the updated resource.
type ResourceUpdateSubject struct {
	SubjectType    string            `json:"subject_type"`
	HTTPStatusCode int               `json:"http_status_code"`
	Links          map[string]string `json:"links"`
	Version        string            `json:"version"`
}

// ConsentAuthRevokedEvent is the payload for EventTypeConsentAuthRevoked.
type ConsentAuthRevokedEvent struct {
	Reason string `json:"reason,omitempty"`
}

// ────────────────────────────────────────────────────────────────────────────
// Handler types
// ────────────────────────────────────────────────────────────────────────────

// Handler is a function that processes a decoded event payload.
// The raw bytes are the JSON value of the event within the "events" map.
type Handler func(ctx context.Context, envelope *Envelope, raw json.RawMessage) error

// ────────────────────────────────────────────────────────────────────────────
// Dead-letter queue
// ────────────────────────────────────────────────────────────────────────────

// DeadLetterItem stores a failed event delivery attempt.
type DeadLetterItem struct {
	EventType EventType
	Envelope  *Envelope
	Raw       json.RawMessage
	Err       error
	FailedAt  time.Time
	Attempts  int
}

// DLQ is a simple in-memory dead-letter queue.
type DLQ struct {
	mu    sync.Mutex
	items []DeadLetterItem
	cap   int
}

// NewDLQ creates a DLQ with the given capacity. When full, oldest items are
// dropped to make room.
func NewDLQ(capacity int) *DLQ {
	if capacity <= 0 {
		capacity = 1000
	}
	return &DLQ{cap: capacity}
}

// Push adds an item to the DLQ.
func (q *DLQ) Push(item DeadLetterItem) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.items) >= q.cap {
		q.items = q.items[1:] // drop oldest
	}
	q.items = append(q.items, item)
}

// Drain returns and removes all items from the DLQ.
func (q *DLQ) Drain() []DeadLetterItem {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]DeadLetterItem, len(q.items))
	copy(out, q.items)
	q.items = nil
	return out
}

// Len returns the number of items in the DLQ.
func (q *DLQ) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}

// ────────────────────────────────────────────────────────────────────────────
// Dispatcher
// ────────────────────────────────────────────────────────────────────────────

// Dispatcher routes incoming OBIE event notifications to registered handlers.
// It supports multiple handlers per event type and delivers to all of them.
type Dispatcher struct {
	mu       sync.RWMutex
	handlers map[EventType][]Handler
	catch    []Handler // wildcard handlers receive every event
	dlq      *DLQ
	log      Logger
}

// Logger is the logging interface used by Dispatcher.
type Logger interface {
	Errorf(format string, args ...any)
	Infof(format string, args ...any)
}

type nopLog struct{}
func (nopLog) Errorf(_ string, _ ...any) {}
func (nopLog) Infof(_ string, _ ...any)  {}

// NewDispatcher creates a Dispatcher. Pass a non-nil DLQ to enable
// dead-lettering of failed handler invocations.
func NewDispatcher(dlq *DLQ, log Logger) *Dispatcher {
	if log == nil {
		log = nopLog{}
	}
	return &Dispatcher{
		handlers: make(map[EventType][]Handler),
		dlq:      dlq,
		log:      log,
	}
}

// On registers handler for the given eventType.
// Multiple handlers per event type are supported and all will be invoked.
func (d *Dispatcher) On(eventType EventType, handler Handler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers[eventType] = append(d.handlers[eventType], handler)
}

// OnAny registers a handler that is invoked for every event type.
func (d *Dispatcher) OnAny(handler Handler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.catch = append(d.catch, handler)
}

// Dispatch decodes env and delivers each event to its registered handlers.
// If a handler returns an error it is recorded in the DLQ (if set) and
// dispatch continues to remaining handlers — no error is returned to the caller.
func (d *Dispatcher) Dispatch(ctx context.Context, env *Envelope) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	for evType, rawPayload := range env.Events {
		handlers := append(d.handlers[evType], d.catch...)
		for _, h := range handlers {
			if err := h(ctx, env, rawPayload); err != nil {
				d.log.Errorf("webhook: handler error for %s (jti=%s): %v", evType, env.Jti, err)
				if d.dlq != nil {
					d.dlq.Push(DeadLetterItem{
						EventType: evType,
						Envelope:  env,
						Raw:       rawPayload,
						Err:       err,
						FailedAt:  time.Now(),
						Attempts:  1,
					})
				}
			} else {
				d.log.Infof("webhook: delivered %s (jti=%s)", evType, env.Jti)
			}
		}
	}
}

// DispatchJSON is a convenience wrapper that parses raw JSON into an Envelope
// and then calls Dispatch.
func (d *Dispatcher) DispatchJSON(ctx context.Context, body []byte) error {
	var env Envelope
	if err := json.Unmarshal(body, &env); err != nil {
		return fmt.Errorf("webhook: decode envelope: %w", err)
	}
	d.Dispatch(ctx, &env)
	return nil
}

// ────────────────────────────────────────────────────────────────────────────
// Typed handler helpers
// ────────────────────────────────────────────────────────────────────────────

// OnResourceUpdate registers a typed handler for resource-update events.
func (d *Dispatcher) OnResourceUpdate(fn func(ctx context.Context, env *Envelope, ev ResourceUpdateEvent) error) {
	d.On(EventTypeResourceUpdate, func(ctx context.Context, env *Envelope, raw json.RawMessage) error {
		var ev ResourceUpdateEvent
		if err := json.Unmarshal(raw, &ev); err != nil {
			return fmt.Errorf("webhook: decode ResourceUpdateEvent: %w", err)
		}
		return fn(ctx, env, ev)
	})
}

// OnConsentRevoked registers a typed handler for consent-authorization-revoked events.
func (d *Dispatcher) OnConsentRevoked(fn func(ctx context.Context, env *Envelope, ev ConsentAuthRevokedEvent) error) {
	d.On(EventTypeConsentAuthRevoked, func(ctx context.Context, env *Envelope, raw json.RawMessage) error {
		var ev ConsentAuthRevokedEvent
		if err := json.Unmarshal(raw, &ev); err != nil {
			return fmt.Errorf("webhook: decode ConsentAuthRevokedEvent: %w", err)
		}
		return fn(ctx, env, ev)
	})
}

// ReplayDLQ attempts to re-deliver every item currently in the DLQ.
// Successfully re-delivered items are removed; persistent failures remain.
func (d *Dispatcher) ReplayDLQ(ctx context.Context) {
	if d.dlq == nil {
		return
	}
	items := d.dlq.Drain()
	for _, item := range items {
		d.mu.RLock()
		handlers := append(d.handlers[item.EventType], d.catch...)
		d.mu.RUnlock()

		for _, h := range handlers {
			if err := h(ctx, item.Envelope, item.Raw); err != nil {
				item.Attempts++
				item.Err = err
				d.dlq.Push(item) // back to DLQ
			}
		}
	}
}
