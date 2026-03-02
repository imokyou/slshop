package shopline

import (
	"fmt"
	"sync"
	"time"
)

// cbState represents the state of a circuit breaker.
type cbState int

const (
	// cbClosed is the normal state — requests flow through.
	cbClosed cbState = iota
	// cbOpen means the circuit is tripped — requests are rejected immediately.
	cbOpen
	// cbHalfOpen allows one probe request to test if the service recovered.
	cbHalfOpen
)

// CircuitBreaker implements a three-state circuit breaker pattern:
//
//	Closed → (threshold failures) → Open → (cooldown) → Half-Open → (success) → Closed
//	                                                               → (failure)  → Open
//
// It is safe for concurrent use.
type CircuitBreaker struct {
	threshold int
	cooldown  time.Duration

	mu           sync.Mutex
	state        cbState
	failures     int
	lastFailTime time.Time
	probing      bool // true while a half-open probe is in flight
}

// newCircuitBreaker creates a CircuitBreaker.
//
//   - threshold: consecutive failures before opening (e.g. 5)
//   - cooldown: how long to stay Open before transitioning to Half-Open (e.g. 30s)
func newCircuitBreaker(threshold int, cooldown time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		threshold: threshold,
		cooldown:  cooldown,
		state:     cbClosed,
	}
}

// Allow checks whether a request is allowed to proceed.
// Returns an error if the circuit is Open (and cooldown has not elapsed) or
// if a Half-Open probe is already in flight.
func (cb *CircuitBreaker) Allow() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case cbClosed:
		return nil

	case cbOpen:
		// Check if cooldown has elapsed — if so, move to Half-Open
		if time.Since(cb.lastFailTime) >= cb.cooldown {
			cb.state = cbHalfOpen
			cb.probing = true
			return nil // allow the probe request
		}
		remaining := cb.cooldown - time.Since(cb.lastFailTime)
		return fmt.Errorf("shopline: circuit breaker is open, retry after %.1fs", remaining.Seconds())

	case cbHalfOpen:
		if cb.probing {
			// Another goroutine is already probing — reject
			return fmt.Errorf("shopline: circuit breaker is half-open, probe in progress")
		}
		cb.probing = true
		return nil
	}

	return nil
}

// RecordSuccess records a successful request outcome.
// In Half-Open state, this closes the circuit and resets the failure counter.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0
	cb.probing = false
	cb.state = cbClosed
}

// RecordFailure records a failed request outcome.
// In Closed state, accumulates failures. When the threshold is reached, opens the circuit.
// In Half-Open state, immediately re-opens the circuit.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.probing = false
	cb.lastFailTime = time.Now()

	switch cb.state {
	case cbClosed:
		cb.failures++
		if cb.failures >= cb.threshold {
			cb.state = cbOpen
		}
	case cbHalfOpen:
		cb.state = cbOpen
	}
}

// State returns the current circuit breaker state as a string (for logging/metrics).
func (cb *CircuitBreaker) State() string {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	switch cb.state {
	case cbClosed:
		return "closed"
	case cbOpen:
		return "open"
	case cbHalfOpen:
		return "half-open"
	}
	return "unknown"
}
