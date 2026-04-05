// Package circuitbreaker implements the circuit-breaker pattern for the OBIE
// SDK's HTTP layer. It prevents cascading failures by stopping requests to an
// ASPSP that is consistently returning errors, giving it time to recover.
//
// States:
//
//	Closed   – normal operation; failures are counted.
//	Open     – requests are rejected immediately; after openTimeout the circuit
//	           moves to HalfOpen.
//	HalfOpen – a single probe request is allowed; success closes the circuit,
//	           failure re-opens it.
package circuitbreaker

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// roundTripFunc adapts a function to implement http.RoundTripper.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// State represents the current circuit-breaker state.
type State int

const (
	StateClosed   State = iota // normal operation
	StateOpen                  // circuit open — requests rejected
	StateHalfOpen              // single probe request allowed
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "Closed"
	case StateOpen:
		return "Open"
	case StateHalfOpen:
		return "HalfOpen"
	default:
		return "Unknown"
	}
}

// StateChangeFunc is called whenever the circuit changes state.
type StateChangeFunc func(from, to State)

// Config holds circuit-breaker configuration.
type Config struct {
	// MaxFailures is the number of consecutive failures before opening the circuit.
	MaxFailures int
	// OpenTimeout is how long the circuit stays open before moving to HalfOpen.
	OpenTimeout time.Duration
	// SuccessThreshold is the number of consecutive successes in HalfOpen before closing.
	SuccessThreshold int
	// IsFailure classifies a response as a failure. Defaults to 5xx status codes and errors.
	IsFailure func(*http.Response, error) bool
	// OnStateChange is called when the circuit transitions between states.
	OnStateChange StateChangeFunc
}

func (c *Config) defaults() {
	if c.MaxFailures == 0 {
		c.MaxFailures = 5
	}
	if c.OpenTimeout == 0 {
		c.OpenTimeout = 30 * time.Second
	}
	if c.SuccessThreshold == 0 {
		c.SuccessThreshold = 2
	}
	if c.IsFailure == nil {
		c.IsFailure = defaultIsFailure
	}
}

func defaultIsFailure(resp *http.Response, err error) bool {
	if err != nil {
		return true
	}
	return resp != nil && resp.StatusCode >= 500
}

// ErrCircuitOpen is returned when the circuit is open and a request is blocked.
type ErrCircuitOpen struct {
	OpenedAt time.Time
}

func (e ErrCircuitOpen) Error() string {
	return fmt.Sprintf("circuitbreaker: circuit open since %s", e.OpenedAt.Format(time.RFC3339))
}

// CircuitBreaker wraps an http.RoundTripper with circuit-breaker semantics.
type CircuitBreaker struct {
	cfg Config

	mu               sync.Mutex
	state            State
	failures         int
	successes        int
	openedAt         time.Time
	halfOpenInFlight bool
}

// New creates a CircuitBreaker with the given Config.
func New(cfg Config) *CircuitBreaker {
	cfg.defaults()
	return &CircuitBreaker{cfg: cfg}
}

// allow returns nil if the request is permitted, or ErrCircuitOpen if blocked.
func (cb *CircuitBreaker) allow() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return nil
	case StateOpen:
		if time.Since(cb.openedAt) >= cb.cfg.OpenTimeout {
			cb.transition(StateHalfOpen)
			cb.halfOpenInFlight = true
			return nil
		}
		return ErrCircuitOpen{OpenedAt: cb.openedAt}
	case StateHalfOpen:
		if cb.halfOpenInFlight {
			return ErrCircuitOpen{OpenedAt: cb.openedAt}
		}
		cb.halfOpenInFlight = true
		return nil
	}
	return nil
}

func (cb *CircuitBreaker) record(resp *http.Response, err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	failed := cb.cfg.IsFailure(resp, err)

	switch cb.state {
	case StateClosed:
		if failed {
			cb.failures++
			if cb.failures >= cb.cfg.MaxFailures {
				cb.transition(StateOpen)
			}
		} else {
			cb.failures = 0
		}
	case StateHalfOpen:
		cb.halfOpenInFlight = false
		if failed {
			cb.transition(StateOpen)
		} else {
			cb.successes++
			if cb.successes >= cb.cfg.SuccessThreshold {
				cb.transition(StateClosed)
			}
		}
	}
}

func (cb *CircuitBreaker) transition(next State) {
	prev := cb.state
	cb.state = next
	switch next {
	case StateOpen:
		cb.openedAt = time.Now()
		cb.failures = 0
		cb.successes = 0
	case StateClosed:
		cb.failures = 0
		cb.successes = 0
	case StateHalfOpen:
		cb.successes = 0
	}
	if cb.cfg.OnStateChange != nil && prev != next {
		cb.cfg.OnStateChange(prev, next)
	}
}

// Middleware returns an http.RoundTripper middleware that applies the circuit breaker.
func (cb *CircuitBreaker) Middleware() func(http.RoundTripper) http.RoundTripper {
	return func(next http.RoundTripper) http.RoundTripper {
		return roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if err := cb.allow(); err != nil {
				return nil, err
			}
			resp, err := next.RoundTrip(req)
			cb.record(resp, err)
			return resp, err
		})
	}
}

// State returns the current circuit state.
func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// Reset forces the circuit back to Closed.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.transition(StateClosed)
}
