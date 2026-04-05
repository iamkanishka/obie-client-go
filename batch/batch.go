// Package batch provides primitives for executing concurrent OBIE API calls
// with bounded parallelism, result aggregation, and partial-failure handling.
//
// The primary use-case is fetching data across multiple accounts in parallel —
// e.g. fetching balances or transactions for all accounts in one round-trip.
package batch

import (
	"context"
	"fmt"
	"sync"
)

// ────────────────────────────────────────────────────────────────────────────
// Result
// ────────────────────────────────────────────────────────────────────────────

// Result wraps the outcome of a single item in a batch operation.
// T is the success type; if Err is non-nil, Value is the zero value.
type Result[T any] struct {
	Key   string // identifies which input produced this result
	Value T
	Err   error
}

// ────────────────────────────────────────────────────────────────────────────
// Executor
// ────────────────────────────────────────────────────────────────────────────

// Executor runs a function over a slice of keys with bounded concurrency.
// It collects all results (including errors) so callers can distinguish
// partial failures from total failures.
type Executor[K any, V any] struct {
	concurrency int
}

// NewExecutor creates an Executor with the given maximum concurrency.
// If concurrency <= 0 it defaults to 5.
func NewExecutor[K any, V any](concurrency int) *Executor[K, V] {
	if concurrency <= 0 {
		concurrency = 5
	}
	return &Executor[K, V]{concurrency: concurrency}
}

// Run executes fn(ctx, key) for each key in keys, respecting the concurrency
// limit. The results slice is returned in the same order as keys.
// The context is checked before each goroutine launch; cancellation stops
// new launches but allows in-flight calls to complete.
func (e *Executor[K, V]) Run(
	ctx context.Context,
	keys []K,
	keyString func(K) string,
	fn func(ctx context.Context, key K) (V, error),
) []Result[V] {
	results := make([]Result[V], len(keys))
	sem := make(chan struct{}, e.concurrency)
	var wg sync.WaitGroup

	for i, key := range keys {
		// Check context before launching.
		select {
		case <-ctx.Done():
			results[i] = Result[V]{Key: keyString(key), Err: fmt.Errorf("batch: context cancelled: %w", ctx.Err())}
			continue
		default:
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, k K) {
			defer wg.Done()
			defer func() { <-sem }()
			v, err := fn(ctx, k)
			results[idx] = Result[V]{Key: keyString(k), Value: v, Err: err}
		}(i, key)
	}

	wg.Wait()
	return results
}

// ────────────────────────────────────────────────────────────────────────────
// Partition helpers
// ────────────────────────────────────────────────────────────────────────────

// Partition separates a results slice into successful values and errors.
func Partition[V any](results []Result[V]) (successes []V, failures []error) {
	for _, r := range results {
		if r.Err != nil {
			failures = append(failures, fmt.Errorf("%s: %w", r.Key, r.Err))
		} else {
			successes = append(successes, r.Value)
		}
	}
	return
}

// MustAll returns the values from results, or returns an error if any result
// contained a failure. Useful when all-or-nothing semantics are required.
func MustAll[V any](results []Result[V]) ([]V, error) {
	vals, errs := Partition(results)
	if len(errs) > 0 {
		msgs := make([]string, len(errs))
		for i, e := range errs {
			msgs[i] = e.Error()
		}
		return nil, fmt.Errorf("batch: %d failure(s): %v", len(errs), msgs)
	}
	return vals, nil
}

// ────────────────────────────────────────────────────────────────────────────
// AccountFetcher — typed helper for the most common OBIE fan-out pattern
// ────────────────────────────────────────────────────────────────────────────

// AccountFetcher runs a fetch function in parallel for every account ID.
// It is a convenience wrapper around Executor[string, V].
type AccountFetcher[V any] struct {
	exec *Executor[string, V]
}

// NewAccountFetcher creates an AccountFetcher with bounded concurrency.
func NewAccountFetcher[V any](concurrency int) *AccountFetcher[V] {
	return &AccountFetcher[V]{exec: NewExecutor[string, V](concurrency)}
}

// FetchAll invokes fn(ctx, accountID) for every ID in accountIDs concurrently
// and returns all results.
func (f *AccountFetcher[V]) FetchAll(
	ctx context.Context,
	accountIDs []string,
	fn func(ctx context.Context, accountID string) (V, error),
) []Result[V] {
	return f.exec.Run(ctx, accountIDs, func(s string) string { return s }, fn)
}

// ────────────────────────────────────────────────────────────────────────────
// Pipeline — sequential stage processing with early-exit on error
// ────────────────────────────────────────────────────────────────────────────

// Stage is a function that transforms a value of type T into a new value,
// returning an error to abort the pipeline.
type Stage[T any] func(ctx context.Context, input T) (T, error)

// Pipeline executes a sequence of stages over an initial value, returning
// the final value or the first error encountered.
func Pipeline[T any](ctx context.Context, initial T, stages ...Stage[T]) (T, error) {
	current := initial
	for i, stage := range stages {
		var err error
		current, err = stage(ctx, current)
		if err != nil {
			return current, fmt.Errorf("pipeline: stage %d: %w", i, err)
		}
		if ctx.Err() != nil {
			return current, fmt.Errorf("pipeline: context cancelled at stage %d: %w", i, ctx.Err())
		}
	}
	return current, nil
}
