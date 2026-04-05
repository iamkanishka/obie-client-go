// Package cache provides in-memory TTL caching for OBIE consent objects and
// arbitrary key-value data. All operations are safe for concurrent use.
package cache

import (
	"sync"
	"time"
)

// ────────────────────────────────────────────────────────────────────────────
// Generic TTL cache
// ────────────────────────────────────────────────────────────────────────────

// entry holds a value and its expiry timestamp.
type entry[V any] struct {
	value     V
	expiresAt time.Time
}

// Cache is a generic, thread-safe TTL cache.
type Cache[K comparable, V any] struct {
	mu      sync.RWMutex
	items   map[K]entry[V]
	defaultTTL time.Duration
}

// New creates a Cache with the given default TTL for all Set calls that don't
// specify their own TTL.
func New[K comparable, V any](defaultTTL time.Duration) *Cache[K, V] {
	c := &Cache[K, V]{
		items:      make(map[K]entry[V]),
		defaultTTL: defaultTTL,
	}
	go c.runEviction()
	return c
}

// Set stores value under key with the default TTL.
func (c *Cache[K, V]) Set(key K, value V) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL stores value under key with an explicit TTL.
func (c *Cache[K, V]) SetWithTTL(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = entry[V]{value: value, expiresAt: time.Now().Add(ttl)}
}

// Get returns the value stored under key and whether it was found and unexpired.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.items[key]
	if !ok || time.Now().After(e.expiresAt) {
		var zero V
		return zero, false
	}
	return e.value, true
}

// Delete removes the entry for key.
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Flush removes all entries.
func (c *Cache[K, V]) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[K]entry[V])
}

// Len returns the number of (possibly stale) entries.
func (c *Cache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// runEviction periodically removes expired entries so the map doesn't grow
// indefinitely. Runs every defaultTTL or 60 s, whichever is shorter.
func (c *Cache[K, V]) runEviction() {
	interval := c.defaultTTL
	if interval > 60*time.Second || interval <= 0 {
		interval = 60 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		c.evict()
	}
}

func (c *Cache[K, V]) evict() {
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, e := range c.items {
		if now.After(e.expiresAt) {
			delete(c.items, k)
		}
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Typed consent cache
// ────────────────────────────────────────────────────────────────────────────

// ConsentEntry holds a serialised consent for any service (AIS/PIS/VRP/CBPII).
type ConsentEntry struct {
	ConsentID string
	Status    string
	Payload   []byte // raw JSON of the consent response
	CreatedAt time.Time
	ExpiresAt time.Time
}

// ConsentCache is a named, TTL-backed store for OBIE consent objects.
type ConsentCache struct {
	inner *Cache[string, ConsentEntry]
}

// NewConsentCache creates a ConsentCache with the given default TTL.
func NewConsentCache(defaultTTL time.Duration) *ConsentCache {
	return &ConsentCache{inner: New[string, ConsentEntry](defaultTTL)}
}

// Store saves a consent entry keyed by its ConsentID.
// If entry.ExpiresAt is set, the cache entry will expire at that time;
// otherwise the default TTL applies.
func (cc *ConsentCache) Store(entry ConsentEntry) {
	ttl := cc.inner.defaultTTL
	if !entry.ExpiresAt.IsZero() {
		if d := time.Until(entry.ExpiresAt); d > 0 {
			ttl = d
		}
	}
	cc.inner.SetWithTTL(entry.ConsentID, entry, ttl)
}

// Load retrieves a consent entry by ConsentID.
func (cc *ConsentCache) Load(consentID string) (ConsentEntry, bool) {
	return cc.inner.Get(consentID)
}

// Revoke removes a consent from the cache (e.g. after authorisation or rejection).
func (cc *ConsentCache) Revoke(consentID string) {
	cc.inner.Delete(consentID)
}

// ────────────────────────────────────────────────────────────────────────────
// Response cache (for GET idempotency / ETag support)
// ────────────────────────────────────────────────────────────────────────────

// ResponseEntry caches a raw HTTP response body and its ETag.
type ResponseEntry struct {
	Body       []byte
	ETag       string
	StatusCode int
	CachedAt   time.Time
}

// ResponseCache caches GET responses keyed by URL.
type ResponseCache struct {
	inner *Cache[string, ResponseEntry]
}

// NewResponseCache creates a ResponseCache with the given TTL.
func NewResponseCache(ttl time.Duration) *ResponseCache {
	return &ResponseCache{inner: New[string, ResponseEntry](ttl)}
}

// Set stores a response entry for url.
func (rc *ResponseCache) Set(url string, entry ResponseEntry) {
	entry.CachedAt = time.Now()
	rc.inner.Set(url, entry)
}

// Get returns the cached response for url, if present and unexpired.
func (rc *ResponseCache) Get(url string) (ResponseEntry, bool) {
	return rc.inner.Get(url)
}

// Invalidate removes the cached response for url.
func (rc *ResponseCache) Invalidate(url string) {
	rc.inner.Delete(url)
}
