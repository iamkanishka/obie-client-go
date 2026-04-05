package pagination_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/iamkanishka/obie-client-go/pagination"
)

type txn struct {
	ID     string `json:"TransactionId"`
	Amount string `json:"Amount"`
}

// buildPage returns JSON for a single page with optional next URL.
func buildPage(items []txn, nextURL string) []byte {
	type links struct {
		Next string `json:"Next,omitempty"`
	}
	type data struct {
		Transaction []txn `json:"Transaction"`
	}
	type page struct {
		Data  data  `json:"Data"`
		Links links `json:"Links"`
	}
	p := page{
		Data:  data{Transaction: items},
		Links: links{Next: nextURL},
	}
	b, _ := json.Marshal(p)
	return b
}

func TestIterator_SinglePage(t *testing.T) {
	items := []txn{
		{ID: "tx1", Amount: "10.00"},
		{ID: "tx2", Amount: "20.00"},
	}
	fetcher := func(_ context.Context, url string) ([]byte, error) {
		return buildPage(items, ""), nil
	}

	iter := pagination.New[txn](context.Background(), "http://api/transactions", fetcher, "Data.Transaction")

	var got []txn
	for iter.Next() {
		got = append(got, iter.Item())
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("iterator error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("got %d items, want 2", len(got))
	}
	if got[0].ID != "tx1" {
		t.Errorf("first item ID: got %q, want %q", got[0].ID, "tx1")
	}
}

func TestIterator_MultiplePages(t *testing.T) {
	pages := map[string][]txn{
		"http://api/transactions":        {{ID: "tx1"}, {ID: "tx2"}},
		"http://api/transactions?page=2": {{ID: "tx3"}, {ID: "tx4"}},
		"http://api/transactions?page=3": {{ID: "tx5"}},
	}
	nextURLs := map[string]string{
		"http://api/transactions":        "http://api/transactions?page=2",
		"http://api/transactions?page=2": "http://api/transactions?page=3",
		"http://api/transactions?page=3": "",
	}

	fetcher := func(_ context.Context, url string) ([]byte, error) {
		items, ok := pages[url]
		if !ok {
			return nil, fmt.Errorf("unknown URL: %s", url)
		}
		return buildPage(items, nextURLs[url]), nil
	}

	all, err := pagination.New[txn](
		context.Background(), "http://api/transactions", fetcher, "Data.Transaction",
	).All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(all) != 5 {
		t.Errorf("got %d items across pages, want 5", len(all))
	}
	if all[4].ID != "tx5" {
		t.Errorf("last item: got %q, want %q", all[4].ID, "tx5")
	}
}

func TestIterator_FetcherError(t *testing.T) {
	fetcher := func(_ context.Context, url string) ([]byte, error) {
		return nil, fmt.Errorf("network failure")
	}

	iter := pagination.New[txn](context.Background(), "http://api/transactions", fetcher, "Data.Transaction")

	for iter.Next() {
		iter.Item()
	}
	if iter.Err() == nil {
		t.Error("expected error from failed fetcher, got nil")
	}
}

func TestIterator_EmptyResponse(t *testing.T) {
	fetcher := func(_ context.Context, url string) ([]byte, error) {
		return buildPage(nil, ""), nil
	}

	all, err := pagination.New[txn](
		context.Background(), "http://api/transactions", fetcher, "Data.Transaction",
	).All()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 0 {
		t.Errorf("expected 0 items, got %d", len(all))
	}
}

func TestIterator_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	call := 0
	fetcher := func(c context.Context, url string) ([]byte, error) {
		call++
		if call == 2 {
			cancel() // cancel after first page
		}
		if c.Err() != nil {
			return nil, c.Err()
		}
		return buildPage([]txn{{ID: "tx"}}, "http://api/p2"), nil
	}

	iter := pagination.New[txn](ctx, "http://api/p1", fetcher, "Data.Transaction")
	count := 0
	for iter.Next() {
		iter.Item()
		count++
	}

	if iter.Err() == nil {
		t.Error("expected context cancellation error")
	}
	if count == 0 {
		t.Error("expected at least one item from first page")
	}
}
