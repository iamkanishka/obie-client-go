package webhook_test

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/iamkanishka/obie-client-go/webhook"
)

func buildEnvelope(events map[webhook.EventType]any) *webhook.Envelope {
	raw := make(map[webhook.EventType]json.RawMessage, len(events))
	for k, v := range events {
		b, _ := json.Marshal(v)
		raw[k] = b
	}
	return &webhook.Envelope{
		Jti:    "test-jti",
		Events: raw,
	}
}

func TestDispatcher_RoutesToCorrectHandler(t *testing.T) {
	d := webhook.NewDispatcher(nil, nil)

	var called int32
	d.On(webhook.EventTypeResourceUpdate, func(_ context.Context, _ *webhook.Envelope, _ json.RawMessage) error {
		atomic.AddInt32(&called, 1)
		return nil
	})

	env := buildEnvelope(map[webhook.EventType]any{
		webhook.EventTypeResourceUpdate: map[string]string{"foo": "bar"},
	})
	d.Dispatch(context.Background(), env)

	if atomic.LoadInt32(&called) != 1 {
		t.Errorf("handler call count: got %d, want 1", called)
	}
}

func TestDispatcher_MultipleHandlersSameType(t *testing.T) {
	d := webhook.NewDispatcher(nil, nil)

	var count int32
	inc := func(_ context.Context, _ *webhook.Envelope, _ json.RawMessage) error {
		atomic.AddInt32(&count, 1)
		return nil
	}

	d.On(webhook.EventTypeResourceUpdate, inc)
	d.On(webhook.EventTypeResourceUpdate, inc)
	d.On(webhook.EventTypeResourceUpdate, inc)

	env := buildEnvelope(map[webhook.EventType]any{
		webhook.EventTypeResourceUpdate: map[string]string{},
	})
	d.Dispatch(context.Background(), env)

	if atomic.LoadInt32(&count) != 3 {
		t.Errorf("expected 3 handler invocations, got %d", count)
	}
}

func TestDispatcher_WildcardHandlerReceivesAllEvents(t *testing.T) {
	d := webhook.NewDispatcher(nil, nil)

	var caught []webhook.EventType
	var mu sync.Mutex // protect slice
	d.OnAny(func(_ context.Context, env *webhook.Envelope, _ json.RawMessage) error {
		// Collect event types from envelope.
		for et := range env.Events {
			mu.Lock()
			caught = append(caught, et)
			mu.Unlock()
		}
		return nil
	})

	env := buildEnvelope(map[webhook.EventType]any{
		webhook.EventTypeResourceUpdate: map[string]string{},
		webhook.EventTypeConsentAuthRevoked: map[string]string{},
	})
	d.Dispatch(context.Background(), env)

	// Wildcard fires once per event type × handler invocation.
	if len(caught) < 2 {
		t.Errorf("wildcard handler caught %d events, want >= 2", len(caught))
	}
}

func TestDispatcher_FailedHandlerGoesToDLQ(t *testing.T) {
	dlq := webhook.NewDLQ(100)
	d := webhook.NewDispatcher(dlq, nil)

	d.On(webhook.EventTypeResourceUpdate, func(_ context.Context, _ *webhook.Envelope, _ json.RawMessage) error {
		return errors.New("handler failure")
	})

	env := buildEnvelope(map[webhook.EventType]any{
		webhook.EventTypeResourceUpdate: map[string]string{},
	})
	d.Dispatch(context.Background(), env)

	if dlq.Len() != 1 {
		t.Errorf("DLQ length: got %d, want 1", dlq.Len())
	}
}

func TestDispatcher_DispatchJSON(t *testing.T) {
	d := webhook.NewDispatcher(nil, nil)

	var called int32
	d.On(webhook.EventTypeConsentAuthRevoked, func(_ context.Context, _ *webhook.Envelope, _ json.RawMessage) error {
		atomic.AddInt32(&called, 1)
		return nil
	})

	body, _ := json.Marshal(map[string]any{
		"jti": "jti-1",
		"events": map[string]any{
			string(webhook.EventTypeConsentAuthRevoked): map[string]string{},
		},
	})
	if err := d.DispatchJSON(context.Background(), body); err != nil {
		t.Fatalf("DispatchJSON: %v", err)
	}
	if atomic.LoadInt32(&called) != 1 {
		t.Errorf("handler calls: got %d, want 1", called)
	}
}

func TestDispatcher_TypedResourceUpdateHandler(t *testing.T) {
	d := webhook.NewDispatcher(nil, nil)

	var got webhook.ResourceUpdateEvent
	d.OnResourceUpdate(func(_ context.Context, _ *webhook.Envelope, ev webhook.ResourceUpdateEvent) error {
		got = ev
		return nil
	})

	env := buildEnvelope(map[webhook.EventType]any{
		webhook.EventTypeResourceUpdate: webhook.ResourceUpdateEvent{
			Subject: webhook.ResourceUpdateSubject{
				SubjectType:    "http://openbanking.org.uk/rid_http://openbanking.org.uk/rty",
				HTTPStatusCode: 200,
			},
		},
	})
	d.Dispatch(context.Background(), env)

	if got.Subject.HTTPStatusCode != 200 {
		t.Errorf("decoded event HTTPStatusCode: got %d, want 200", got.Subject.HTTPStatusCode)
	}
}

func TestDLQ_Capacity(t *testing.T) {
	dlq := webhook.NewDLQ(3)
	for i := 0; i < 5; i++ {
		dlq.Push(webhook.DeadLetterItem{EventType: webhook.EventTypeResourceUpdate})
	}
	if dlq.Len() != 3 {
		t.Errorf("DLQ length: got %d, want 3 (capacity)", dlq.Len())
	}
}

func TestDLQ_Drain(t *testing.T) {
	dlq := webhook.NewDLQ(10)
	dlq.Push(webhook.DeadLetterItem{})
	dlq.Push(webhook.DeadLetterItem{})

	items := dlq.Drain()
	if len(items) != 2 {
		t.Errorf("drained %d items, want 2", len(items))
	}
	if dlq.Len() != 0 {
		t.Error("DLQ should be empty after Drain")
	}
}

func TestDispatcher_ReplayDLQ(t *testing.T) {
	dlq := webhook.NewDLQ(100)
	d := webhook.NewDispatcher(dlq, nil)

	attempts := int32(0)
	d.On(webhook.EventTypeResourceUpdate, func(_ context.Context, _ *webhook.Envelope, _ json.RawMessage) error {
		n := atomic.AddInt32(&attempts, 1)
		if n < 3 {
			return errors.New("transient failure")
		}
		return nil
	})

	env := buildEnvelope(map[webhook.EventType]any{
		webhook.EventTypeResourceUpdate: map[string]string{},
	})
	d.Dispatch(context.Background(), env) // attempt 1 → fails → DLQ
	d.ReplayDLQ(context.Background())     // attempt 2 → fails → back to DLQ
	d.ReplayDLQ(context.Background())     // attempt 3 → succeeds → removed from DLQ

	if dlq.Len() != 0 {
		t.Errorf("expected empty DLQ after successful replay, got %d items", dlq.Len())
	}
}
