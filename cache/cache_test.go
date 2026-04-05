package cache_test

import (
	"testing"
	"time"

	"github.com/iamkanishka/obie-client-go/cache"
)

func TestCache_SetGet(t *testing.T) {
	c := cache.New[string, int](time.Minute)
	c.Set("foo", 42)

	v, ok := c.Get("foo")
	if !ok {
		t.Fatal("expected key to be found")
	}
	if v != 42 {
		t.Errorf("got %d, want 42", v)
	}
}

func TestCache_Expiry(t *testing.T) {
	c := cache.New[string, string](10 * time.Millisecond)
	c.Set("key", "value")

	time.Sleep(20 * time.Millisecond)

	_, ok := c.Get("key")
	if ok {
		t.Error("expected key to be expired")
	}
}

func TestCache_Delete(t *testing.T) {
	c := cache.New[string, string](time.Minute)
	c.Set("key", "val")
	c.Delete("key")
	_, ok := c.Get("key")
	if ok {
		t.Error("expected key to be absent after Delete")
	}
}

func TestCache_Flush(t *testing.T) {
	c := cache.New[string, int](time.Minute)
	c.Set("a", 1)
	c.Set("b", 2)
	c.Flush()
	if c.Len() != 0 {
		t.Errorf("expected 0 items after Flush, got %d", c.Len())
	}
}

func TestCache_SetWithTTL(t *testing.T) {
	c := cache.New[string, string](time.Hour)
	c.SetWithTTL("short", "val", 10*time.Millisecond)
	c.Set("long", "val")

	time.Sleep(20 * time.Millisecond)

	if _, ok := c.Get("short"); ok {
		t.Error("short-lived key should have expired")
	}
	if _, ok := c.Get("long"); !ok {
		t.Error("long-lived key should still be present")
	}
}

func TestConsentCache_StoreLoad(t *testing.T) {
	cc := cache.NewConsentCache(time.Minute)

	entry := cache.ConsentEntry{
		ConsentID: "consent-123",
		Status:    "AwaitingAuthorisation",
		Payload:   []byte(`{"Data":{}}`),
		CreatedAt: time.Now(),
	}
	cc.Store(entry)

	got, ok := cc.Load("consent-123")
	if !ok {
		t.Fatal("expected consent to be found")
	}
	if got.Status != "AwaitingAuthorisation" {
		t.Errorf("Status: got %q, want %q", got.Status, "AwaitingAuthorisation")
	}
}

func TestConsentCache_ExpiresWithPayload(t *testing.T) {
	cc := cache.NewConsentCache(time.Hour)

	entry := cache.ConsentEntry{
		ConsentID: "exp-consent",
		Status:    "Authorised",
		ExpiresAt: time.Now().Add(10 * time.Millisecond),
	}
	cc.Store(entry)
	time.Sleep(20 * time.Millisecond)

	_, ok := cc.Load("exp-consent")
	if ok {
		t.Error("expected expired consent to be absent")
	}
}

func TestConsentCache_Revoke(t *testing.T) {
	cc := cache.NewConsentCache(time.Minute)
	cc.Store(cache.ConsentEntry{ConsentID: "to-revoke", Status: "Authorised"})
	cc.Revoke("to-revoke")
	_, ok := cc.Load("to-revoke")
	if ok {
		t.Error("expected revoked consent to be absent")
	}
}

func TestResponseCache_SetGet(t *testing.T) {
	rc := cache.NewResponseCache(time.Minute)
	rc.Set("https://api.example.com/accounts", cache.ResponseEntry{
		Body:       []byte(`{"Data":{"Account":[]}}`),
		ETag:       `"abc123"`,
		StatusCode: 200,
	})
	e, ok := rc.Get("https://api.example.com/accounts")
	if !ok {
		t.Fatal("expected cached response")
	}
	if e.ETag != `"abc123"` {
		t.Errorf("ETag: got %q, want %q", e.ETag, `"abc123"`)
	}
}
