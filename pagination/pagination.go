// Package pagination provides a generic, lazy-evaluation iterator for OBIE
// APIs that return paginated responses using HATEOAS "next" links.
package pagination

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Page represents a single page returned by an OBIE list endpoint.
// T is the type of each item in the page.
type Page[T any] struct {
	Items   []T
	HasNext bool
	NextURL string
	// Raw holds the full decoded page for callers that need Links/Meta.
	Raw json.RawMessage
}

// Fetcher is the function signature used by Iterator to retrieve pages.
// It must return the raw JSON body for the given URL.
type Fetcher func(ctx context.Context, url string) ([]byte, error)

// Iterator iterates over all pages of an OBIE list endpoint.
// It is NOT safe for concurrent use; create one iterator per goroutine.
//
// Usage:
//
//	iter := pagination.New[models.OBTransaction6](
//	    ctx, firstURL, fetcher, "Data.Transaction",
//	)
//	for iter.Next() {
//	    txn := iter.Item()
//	    // process txn
//	}
//	if err := iter.Err(); err != nil { … }
type Iterator[T any] struct {
	ctx      context.Context
	fetcher  Fetcher
	dataPath string // dot-separated JSON path to the items array, e.g. "Data.Transaction"

	currentPage []T
	currentIdx  int
	nextURL     string
	started     bool
	done        bool
	err         error
}

// New creates an Iterator that starts at firstURL.
// dataPath is the dot-separated path within the response JSON to the array of items,
// e.g. "Data.Transaction" or "Data.Account".
func New[T any](ctx context.Context, firstURL string, fetcher Fetcher, dataPath string) *Iterator[T] {
	return &Iterator[T]{
		ctx:      ctx,
		fetcher:  fetcher,
		dataPath: dataPath,
		nextURL:  firstURL,
	}
}

// Next advances the iterator to the next item. Returns true if an item is
// available, false when all items have been consumed or an error occurred.
func (it *Iterator[T]) Next() bool {
	if it.done {
		return false
	}

	// Serve from current page buffer.
	if it.currentIdx < len(it.currentPage) {
		return true
	}

	// No more items in current page — fetch next page if there is one.
	if it.started && it.nextURL == "" {
		it.done = true
		return false
	}

	if err := it.fetchPage(); err != nil {
		it.err = err
		it.done = true
		return false
	}

	it.started = true
	return it.currentIdx < len(it.currentPage)
}

// Item returns the current item. Must only be called after Next() returns true.
func (it *Iterator[T]) Item() T {
	item := it.currentPage[it.currentIdx]
	it.currentIdx++
	return item
}

// Err returns any error encountered during iteration.
func (it *Iterator[T]) Err() error { return it.err }

// All eagerly collects all remaining items into a slice.
// After calling All(), the iterator is exhausted.
func (it *Iterator[T]) All() ([]T, error) {
	var result []T
	for it.Next() {
		result = append(result, it.Item())
	}
	return result, it.Err()
}

func (it *Iterator[T]) fetchPage() error {
	body, err := it.fetcher(it.ctx, it.nextURL)
	if err != nil {
		return fmt.Errorf("pagination: fetch %s: %w", it.nextURL, err)
	}

	// Decode into a raw map so we can traverse the dataPath.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return fmt.Errorf("pagination: decode response: %w", err)
	}

	// Extract items at dataPath.
	items, err := extractAtPath[T](raw, it.dataPath)
	if err != nil {
		return err
	}

	// Extract next URL from Links.Next.
	it.nextURL = extractNextURL(raw)
	it.currentPage = items
	it.currentIdx = 0
	return nil
}

// extractAtPath walks a dot-separated path through a JSON object and
// unmarshals the value at that path into []T.
func extractAtPath[T any](raw map[string]json.RawMessage, path string) ([]T, error) {
	current := raw
	segments := splitPath(path)

	for i, seg := range segments {
		v, ok := current[seg]
		if !ok {
			return nil, fmt.Errorf("pagination: key %q not found at path segment %d of %q", seg, i, path)
		}

		if i == len(segments)-1 {
			// Final segment — decode as []T.
			var items []T
			if err := json.Unmarshal(v, &items); err != nil {
				return nil, fmt.Errorf("pagination: decode items at %q: %w", path, err)
			}
			return items, nil
		}

		// Intermediate segment — must be an object.
		var nested map[string]json.RawMessage
		if err := json.Unmarshal(v, &nested); err != nil {
			return nil, fmt.Errorf("pagination: traverse path at %q: %w", seg, err)
		}
		current = nested
	}
	return nil, fmt.Errorf("pagination: empty path")
}

// extractNextURL pulls the Links.Next URL from a decoded response.
func extractNextURL(raw map[string]json.RawMessage) string {
	linksRaw, ok := raw["Links"]
	if !ok {
		return ""
	}
	var links struct {
		Next string `json:"Next"`
	}
	if err := json.Unmarshal(linksRaw, &links); err != nil {
		return ""
	}
	return links.Next
}

func splitPath(path string) []string {
	var segments []string
	start := 0
	for i := 0; i < len(path); i++ {
		if path[i] == '.' {
			segments = append(segments, path[start:i])
			start = i + 1
		}
	}
	segments = append(segments, path[start:])
	return segments
}

// ────────────────────────────────────────────────────────────────────────────
// HTTP-based Fetcher helper
// ────────────────────────────────────────────────────────────────────────────

// HTTPFetcher creates a Fetcher that makes authenticated GET requests using
// the provided http.Client and bearer token.
func HTTPFetcher(client *http.Client, bearerToken string) Fetcher {
	return func(ctx context.Context, url string) ([]byte, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+bearerToken)
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("pagination: HTTP %d: %s", resp.StatusCode, string(body))
		}
		return body, nil
	}
}
