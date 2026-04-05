package batch_test

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync/atomic"
	"testing"
	"time"

	"github.com/iamkanishka/obie-client-go/batch"
)

func TestExecutor_RunsAllKeys(t *testing.T) {
	exec := batch.NewExecutor[string, string](3)
	keys := []string{"a", "b", "c", "d", "e"}

	results := exec.Run(context.Background(), keys,
		func(k string) string { return k },
		func(_ context.Context, k string) (string, error) {
			return "val-" + k, nil
		},
	)

	if len(results) != len(keys) {
		t.Fatalf("expected %d results, got %d", len(keys), len(results))
	}
	for i, r := range results {
		if r.Err != nil {
			t.Errorf("result[%d]: unexpected error: %v", i, r.Err)
		}
		want := "val-" + keys[i]
		if r.Value != want {
			t.Errorf("result[%d]: got %q, want %q", i, r.Value, want)
		}
	}
}

func TestExecutor_ConcurrencyLimit(t *testing.T) {
	const limit = 3
	var maxConcurrent, current int64

	exec := batch.NewExecutor[string, struct{}](limit)
	keys := make([]string, 20)
	for i := range keys {
		keys[i] = fmt.Sprintf("key-%d", i)
	}

	exec.Run(context.Background(), keys,
		func(k string) string { return k },
		func(_ context.Context, _ string) (struct{}, error) {
			c := atomic.AddInt64(&current, 1)
			for {
				m := atomic.LoadInt64(&maxConcurrent)
				if c <= m || atomic.CompareAndSwapInt64(&maxConcurrent, m, c) {
					break
				}
			}
			time.Sleep(2 * time.Millisecond)
			atomic.AddInt64(&current, -1)
			return struct{}{}, nil
		},
	)

	if maxConcurrent > int64(limit) {
		t.Errorf("max concurrent goroutines %d exceeded limit %d", maxConcurrent, limit)
	}
}

func TestExecutor_CollectsErrors(t *testing.T) {
	exec := batch.NewExecutor[string, int](5)
	keys := []string{"ok", "fail", "ok2"}

	results := exec.Run(context.Background(), keys,
		func(k string) string { return k },
		func(_ context.Context, k string) (int, error) {
			if k == "fail" {
				return 0, errors.New("simulated failure")
			}
			return 42, nil
		},
	)

	vals, errs := batch.Partition(results)
	if len(vals) != 2 {
		t.Errorf("successes: got %d, want 2", len(vals))
	}
	if len(errs) != 1 {
		t.Errorf("failures: got %d, want 1", len(errs))
	}
}

func TestMustAll_ReturnsErrorOnAnyFailure(t *testing.T) {
	results := []batch.Result[string]{
		{Key: "a", Value: "v1"},
		{Key: "b", Err: errors.New("boom")},
	}
	_, err := batch.MustAll(results)
	if err == nil {
		t.Error("expected error from MustAll with one failure")
	}
}

func TestMustAll_ReturnsValuesWhenAllSucceed(t *testing.T) {
	results := []batch.Result[string]{
		{Key: "a", Value: "v1"},
		{Key: "b", Value: "v2"},
	}
	vals, err := batch.MustAll(results)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vals) != 2 {
		t.Errorf("got %d values, want 2", len(vals))
	}
}

func TestAccountFetcher_FetchAll(t *testing.T) {
	fetcher := batch.NewAccountFetcher[string](4)
	accountIDs := []string{"acc-1", "acc-2", "acc-3"}

	results := fetcher.FetchAll(context.Background(), accountIDs,
		func(_ context.Context, id string) (string, error) {
			return "balance-for-" + id, nil
		},
	)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// Sort for deterministic assertion.
	sort.Slice(results, func(i, j int) bool { return results[i].Key < results[j].Key })
	if results[0].Value != "balance-for-acc-1" {
		t.Errorf("result[0]: got %q", results[0].Value)
	}
}

func TestExecutor_ContextCancellation(t *testing.T) {
	exec := batch.NewExecutor[string, string](1)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	keys := []string{"a", "b", "c"}
	results := exec.Run(ctx, keys,
		func(k string) string { return k },
		func(_ context.Context, k string) (string, error) {
			return k, nil
		},
	)

	// At least some results should contain context errors.
	var ctxErrors int
	for _, r := range results {
		if r.Err != nil {
			ctxErrors++
		}
	}
	if ctxErrors == 0 {
		t.Error("expected at least one context-cancelled error")
	}
}

func TestPipeline_ExecutesStagesInOrder(t *testing.T) {
	var order []int
	stage := func(n int) batch.Stage[int] {
		return func(_ context.Context, v int) (int, error) {
			order = append(order, n)
			return v + 1, nil
		}
	}

	result, err := batch.Pipeline(context.Background(), 0,
		stage(1), stage(2), stage(3),
	)
	if err != nil {
		t.Fatalf("Pipeline: %v", err)
	}
	if result != 3 {
		t.Errorf("final value: got %d, want 3", result)
	}
	if len(order) != 3 || order[0] != 1 || order[1] != 2 || order[2] != 3 {
		t.Errorf("stage order: %v", order)
	}
}

func TestPipeline_AbortOnError(t *testing.T) {
	called := 0
	_, err := batch.Pipeline(context.Background(), "start",
		func(_ context.Context, v string) (string, error) { called++; return v, nil },
		func(_ context.Context, v string) (string, error) { called++; return "", errors.New("stage 2 failed") },
		func(_ context.Context, v string) (string, error) { called++; return v, nil },
	)
	if err == nil {
		t.Error("expected error from failed stage")
	}
	if called != 2 {
		t.Errorf("expected 2 stages called (abort on error), got %d", called)
	}
}
