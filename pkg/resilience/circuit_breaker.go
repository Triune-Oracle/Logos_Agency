// Package resilience provides fault-tolerance primitives for non-differentiable
// fractal operations (Box-Counting Dimension calculations).
//
// The CircuitBreaker implements the classic three-state machine:
//
//	CLOSED  → normal operation; failures are counted.
//	OPEN    → too many failures; all calls are rejected and the fallback
//	          value (D_F,target) is returned immediately.
//	HALF_OPEN → a probe call is allowed after the backoff window expires;
//	          success closes the circuit, failure reopens it.
package resilience

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

// State represents the circuit-breaker state.
type State int

const (
	// StateClosed is the normal operating state.
	StateClosed State = iota
	// StateOpen rejects all requests and returns the fallback value.
	StateOpen
	// StateHalfOpen allows one probe request to test recovery.
	StateHalfOpen
)

// String returns a human-readable state label, used in log/metric output.
func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// Metrics records Prometheus-compatible counters and gauges for circuit-breaker
// events.  The zero value is ready to use.
type Metrics struct {
	mu sync.Mutex

	// StateTransitions counts each CLOSED→OPEN, OPEN→HALF_OPEN, HALF_OPEN→CLOSED
	// and HALF_OPEN→OPEN transition.
	StateTransitions map[string]int64

	// FallbacksUsed counts how many times the fallback value was returned.
	FallbacksUsed int64

	// ConsecutiveFailures is the current run of consecutive failures.
	ConsecutiveFailures int64

	// LastTransitionTime records when the state last changed.
	LastTransitionTime time.Time
}

// recordTransition increments the named counter and notes the time.
func (m *Metrics) recordTransition(label string, t time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.StateTransitions == nil {
		m.StateTransitions = make(map[string]int64)
	}
	m.StateTransitions[label]++
	m.LastTransitionTime = t
}

// PrometheusText returns a minimal Prometheus text-format snapshot of the
// circuit-breaker metrics.
func (m *Metrics) PrometheusText(stateVal State) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	out := "# HELP fractal_circuit_breaker_state Current state (0=CLOSED,1=OPEN,2=HALF_OPEN)\n"
	out += "# TYPE fractal_circuit_breaker_state gauge\n"
	out += fmt.Sprintf("fractal_circuit_breaker_state %d\n", stateVal)

	out += "# HELP fractal_circuit_breaker_fallbacks_total Total fallback values returned\n"
	out += "# TYPE fractal_circuit_breaker_fallbacks_total counter\n"
	out += fmt.Sprintf("fractal_circuit_breaker_fallbacks_total %d\n", m.FallbacksUsed)

	out += "# HELP fractal_circuit_breaker_consecutive_failures Current consecutive failure count\n"
	out += "# TYPE fractal_circuit_breaker_consecutive_failures gauge\n"
	out += fmt.Sprintf("fractal_circuit_breaker_consecutive_failures %d\n", m.ConsecutiveFailures)

	for label, count := range m.StateTransitions {
		out += fmt.Sprintf("fractal_circuit_breaker_transitions_total{transition=%q} %d\n", label, count)
	}
	return out
}

// ErrCircuitOpen is returned by Execute when the circuit is OPEN and no
// fallback has been provided.
var ErrCircuitOpen = errors.New("circuit breaker is OPEN")

// Config holds tunable parameters for the CircuitBreaker.
type Config struct {
	// FailureThreshold is the number of consecutive failures that trips the
	// circuit from CLOSED to OPEN (default 5).
	FailureThreshold int

	// BackoffIntervals is the sequence of durations to wait between retries
	// when the circuit is OPEN.  Each OPEN→probe attempt advances through
	// the slice; the last entry is reused once exhausted.
	// Default: [15s, 30s, 60s].
	BackoffIntervals []time.Duration

	// CallTimeout is the maximum duration allowed for a single wrapped call.
	// Exceeding this is treated as a failure.  Default 5s.
	CallTimeout time.Duration

	// MaxGradientMagnitude is the gradient-magnitude threshold above which a
	// result is considered an overflow failure.  Default 1.0.
	MaxGradientMagnitude float64

	// FallbackD0 is the target fractal dimension (D_F,target) returned when
	// the circuit is OPEN.
	FallbackD0 float64
}

func (c *Config) applyDefaults() {
	if c.FailureThreshold <= 0 {
		c.FailureThreshold = 5
	}
	if len(c.BackoffIntervals) == 0 {
		c.BackoffIntervals = []time.Duration{15 * time.Second, 30 * time.Second, 60 * time.Second}
	}
	if c.CallTimeout <= 0 {
		c.CallTimeout = 5 * time.Second
	}
	if c.MaxGradientMagnitude <= 0 {
		c.MaxGradientMagnitude = 1.0
	}
}

// CircuitBreaker protects a Box-Counting Dimension calculation from
// mathematical singularities, computation timeouts, and gradient overflows.
type CircuitBreaker struct {
	mu     sync.Mutex
	cfg    Config
	state  State
	Metrics Metrics

	consecutiveFailures int
	backoffIndex        int
	openedAt            time.Time
	nextRetryAt         time.Time
}

// NewCircuitBreaker creates a CircuitBreaker with the given configuration.
func NewCircuitBreaker(cfg Config) *CircuitBreaker {
	cfg.applyDefaults()
	return &CircuitBreaker{cfg: cfg}
}

// State returns the current circuit state (thread-safe).
func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// Execute calls fn inside the circuit breaker.
//
//   - If the circuit is CLOSED or HALF_OPEN, fn is invoked with a timeout-
//     bounded context derived from ctx.
//   - A result is classified as a failure when fn returns a non-nil error, the
//     returned D₀ is NaN/Inf, or the supplied gradientMagnitude exceeds
//     MaxGradientMagnitude.
//   - When the circuit is OPEN it immediately returns (fallbackD0, nil) and
//     records a fallback event.
//
// The gradientMagnitude parameter should be the L2-norm (or max-norm) of the
// loss gradient at the D₀ calculation site.  Pass 0 if not applicable.
func (cb *CircuitBreaker) Execute(
	ctx context.Context,
	gradientMagnitude float64,
	fn func(ctx context.Context) (float64, error),
) (float64, error) {
	cb.mu.Lock()
	state := cb.maybeTransitionFromOpen(time.Now())
	cb.mu.Unlock()

	if state == StateOpen {
		cb.mu.Lock()
		cb.Metrics.mu.Lock()
		cb.Metrics.FallbacksUsed++
		cb.Metrics.mu.Unlock()
		cb.mu.Unlock()
		return cb.cfg.FallbackD0, nil
	}

	// Derive a timeout-bounded context.
	callCtx, cancel := context.WithTimeout(ctx, cb.cfg.CallTimeout)
	defer cancel()

	resultCh := make(chan struct {
		val float64
		err error
	}, 1)

	go func() {
		val, err := fn(callCtx)
		resultCh <- struct {
			val float64
			err error
		}{val, err}
	}()

	var d0 float64
	var callErr error

	select {
	case r := <-resultCh:
		d0, callErr = r.val, r.err
	case <-callCtx.Done():
		callErr = fmt.Errorf("box-counting dimension calculation timed out after %s", cb.cfg.CallTimeout)
	}

	// Classify the result.
	failure := cb.isFailure(d0, gradientMagnitude, callErr)

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if failure {
		cb.recordFailure(time.Now())
		// Return fallback when circuit just opened or is already open.
		if cb.state == StateOpen {
			cb.Metrics.mu.Lock()
			cb.Metrics.FallbacksUsed++
			cb.Metrics.mu.Unlock()
			return cb.cfg.FallbackD0, nil
		}
		return cb.cfg.FallbackD0, callErr
	}

	// Success path.
	if state == StateHalfOpen {
		cb.transitionTo(StateClosed, time.Now())
	}
	cb.consecutiveFailures = 0
	cb.Metrics.mu.Lock()
	cb.Metrics.ConsecutiveFailures = 0
	cb.Metrics.mu.Unlock()
	return d0, nil
}

// isFailure returns true when the result should be treated as a failure.
func (cb *CircuitBreaker) isFailure(d0, gradientMag float64, err error) bool {
	if err != nil {
		return true
	}
	if math.IsNaN(d0) || math.IsInf(d0, 0) {
		return true
	}
	if gradientMag > cb.cfg.MaxGradientMagnitude {
		return true
	}
	return false
}

// recordFailure increments the failure counter and potentially trips the circuit.
// Caller must hold cb.mu.
func (cb *CircuitBreaker) recordFailure(now time.Time) {
	cb.consecutiveFailures++
	cb.Metrics.mu.Lock()
	cb.Metrics.ConsecutiveFailures = int64(cb.consecutiveFailures)
	cb.Metrics.mu.Unlock()

	if cb.state == StateHalfOpen || (cb.state == StateClosed && cb.consecutiveFailures >= cb.cfg.FailureThreshold) {
		cb.transitionTo(StateOpen, now)
	}
}

// maybeTransitionFromOpen checks whether the backoff window has elapsed and
// advances the state from OPEN to HALF_OPEN if so.
// Caller must hold cb.mu.
func (cb *CircuitBreaker) maybeTransitionFromOpen(now time.Time) State {
	if cb.state == StateOpen && !now.Before(cb.nextRetryAt) {
		cb.transitionTo(StateHalfOpen, now)
	}
	return cb.state
}

// transitionTo updates the state and records the transition in Metrics.
// Caller must hold cb.mu.
func (cb *CircuitBreaker) transitionTo(newState State, now time.Time) {
	label := fmt.Sprintf("%s->%s", cb.state.String(), newState.String())
	cb.state = newState
	cb.Metrics.recordTransition(label, now)

	if newState == StateOpen {
		cb.openedAt = now
		// Advance the backoff index, clamped to the last interval.
		intervals := cb.cfg.BackoffIntervals
		if cb.backoffIndex >= len(intervals) {
			cb.backoffIndex = len(intervals) - 1
		}
		cb.nextRetryAt = now.Add(intervals[cb.backoffIndex])
		if cb.backoffIndex < len(intervals)-1 {
			cb.backoffIndex++
		}
	}
	if newState == StateClosed {
		cb.consecutiveFailures = 0
		cb.backoffIndex = 0
	}
}

// Reset forces the circuit to the CLOSED state and clears all counters.
// Intended for testing and manual recovery scenarios.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = StateClosed
	cb.consecutiveFailures = 0
	cb.backoffIndex = 0
	cb.openedAt = time.Time{}
	cb.nextRetryAt = time.Time{}
	cb.Metrics = Metrics{}
}
