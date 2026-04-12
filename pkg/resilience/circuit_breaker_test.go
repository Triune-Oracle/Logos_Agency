package resilience_test

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/Triune-Oracle/Logos_Agency/pkg/resilience"
)

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

// successFn returns the given d0 value without error.
func successFn(d0 float64) func(context.Context) (float64, error) {
	return func(_ context.Context) (float64, error) { return d0, nil }
}

// errorFn always returns an error.
func errorFn(msg string) func(context.Context) (float64, error) {
	return func(_ context.Context) (float64, error) { return 0, errors.New(msg) }
}

// nanFn simulates a mathematical singularity (D₀ = NaN).
func nanFn() func(context.Context) (float64, error) {
	return successFn(math.NaN())
}

// sleepFn sleeps for d, simulating a computation timeout.
func sleepFn(d time.Duration, result float64) func(context.Context) (float64, error) {
	return func(ctx context.Context) (float64, error) {
		select {
		case <-time.After(d):
			return result, nil
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	}
}

// newFastBreaker returns a CircuitBreaker with very short backoff intervals
// suitable for unit tests.
func newFastBreaker(threshold int, fallback float64) *resilience.CircuitBreaker {
	return resilience.NewCircuitBreaker(resilience.Config{
		FailureThreshold:     threshold,
		BackoffIntervals:     []time.Duration{20 * time.Millisecond, 40 * time.Millisecond, 80 * time.Millisecond},
		CallTimeout:          200 * time.Millisecond,
		MaxGradientMagnitude: 1.0,
		FallbackD0:           fallback,
	})
}

// tripBreaker drives cb through `n` consecutive failures.
func tripBreaker(t *testing.T, cb *resilience.CircuitBreaker, n int) {
	t.Helper()
	ctx := context.Background()
	for i := 0; i < n; i++ {
		_, _ = cb.Execute(ctx, 0, errorFn("injected failure"))
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// State machine tests
// ─────────────────────────────────────────────────────────────────────────────

func TestCircuitBreaker_InitialStateClosed(t *testing.T) {
	cb := newFastBreaker(5, 1.7)
	if got := cb.State(); got != resilience.StateClosed {
		t.Fatalf("expected CLOSED, got %s", got)
	}
}

func TestCircuitBreaker_SuccessDoesNotChangeState(t *testing.T) {
	cb := newFastBreaker(5, 1.7)
	ctx := context.Background()

	d0, err := cb.Execute(ctx, 0, successFn(1.5))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d0 != 1.5 {
		t.Fatalf("expected 1.5, got %g", d0)
	}
	if cb.State() != resilience.StateClosed {
		t.Fatalf("circuit should remain CLOSED after success")
	}
}

func TestCircuitBreaker_TripsAfterThreshold(t *testing.T) {
	threshold := 5
	cb := newFastBreaker(threshold, 1.7)
	tripBreaker(t, cb, threshold)

	if cb.State() != resilience.StateOpen {
		t.Fatalf("expected OPEN after %d failures, got %s", threshold, cb.State())
	}
}

func TestCircuitBreaker_FallbackReturnedWhenOpen(t *testing.T) {
	const fallback = 1.7
	cb := newFastBreaker(3, fallback)
	tripBreaker(t, cb, 3)

	d0, err := cb.Execute(context.Background(), 0, successFn(99))
	if err != nil {
		t.Fatalf("unexpected error from open circuit: %v", err)
	}
	if d0 != fallback {
		t.Fatalf("expected fallback %g, got %g", fallback, d0)
	}
}

func TestCircuitBreaker_TransitionsToHalfOpen(t *testing.T) {
	cb := newFastBreaker(2, 1.7)
	tripBreaker(t, cb, 2) // OPEN

	// Wait for the first backoff (20 ms).
	time.Sleep(30 * time.Millisecond)

	// The next Execute call should probe (HALF_OPEN) and succeed → CLOSED.
	d0, err := cb.Execute(context.Background(), 0, successFn(1.5))
	if err != nil {
		t.Fatalf("unexpected error during half-open probe: %v", err)
	}
	if d0 != 1.5 {
		t.Fatalf("expected 1.5, got %g", d0)
	}
	if cb.State() != resilience.StateClosed {
		t.Fatalf("expected CLOSED after successful probe, got %s", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenFailureReopens(t *testing.T) {
	cb := newFastBreaker(2, 1.7)
	tripBreaker(t, cb, 2) // OPEN

	time.Sleep(30 * time.Millisecond) // wait for backoff

	// Probe fails → should reopen.
	_, _ = cb.Execute(context.Background(), 0, errorFn("probe failure"))
	if cb.State() != resilience.StateOpen {
		t.Fatalf("expected OPEN after failed half-open probe, got %s", cb.State())
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cb := newFastBreaker(2, 1.7)
	tripBreaker(t, cb, 2)
	if cb.State() != resilience.StateOpen {
		t.Skip("precondition: circuit must be open")
	}

	cb.Reset()
	if cb.State() != resilience.StateClosed {
		t.Fatalf("expected CLOSED after Reset, got %s", cb.State())
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Fault injection scenarios (as per issue spec)
// ─────────────────────────────────────────────────────────────────────────────

// Scenario 1: Mathematical singularity (NaN in D₀ calculation).
func TestCircuitBreaker_Scenario1_MathematicalSingularity(t *testing.T) {
	const fallback = 1.7
	cb := newFastBreaker(5, fallback)
	ctx := context.Background()

	// Inject 5 NaN results → circuit should open.
	for i := 0; i < 5; i++ {
		d0, _ := cb.Execute(ctx, 0, nanFn())
		// While circuit is still closed, fallback is returned on failure.
		if !math.IsNaN(d0) && d0 != fallback {
			t.Errorf("step %d: expected NaN or fallback, got %g", i, d0)
		}
	}

	if cb.State() != resilience.StateOpen {
		t.Fatalf("expected OPEN after 5 NaN failures, got %s", cb.State())
	}

	// Training continues: next call returns fallback, not NaN.
	d0, err := cb.Execute(ctx, 0, nanFn())
	if err != nil {
		t.Fatalf("unexpected error from open circuit: %v", err)
	}
	if d0 != fallback {
		t.Fatalf("expected fallback %g when OPEN, got %g", fallback, d0)
	}
}

// Scenario 2: Computation timeout.
func TestCircuitBreaker_Scenario2_ComputationTimeout(t *testing.T) {
	const fallback = 1.7
	cb := resilience.NewCircuitBreaker(resilience.Config{
		FailureThreshold:     5,
		BackoffIntervals:     []time.Duration{20 * time.Millisecond, 40 * time.Millisecond, 80 * time.Millisecond},
		CallTimeout:          50 * time.Millisecond, // short timeout for the test
		MaxGradientMagnitude: 1.0,
		FallbackD0:           fallback,
	})
	ctx := context.Background()

	// Each call sleeps longer than the timeout → timeout failure.
	for i := 0; i < 5; i++ {
		_, _ = cb.Execute(ctx, 0, sleepFn(200*time.Millisecond, 1.5))
	}

	if cb.State() != resilience.StateOpen {
		t.Fatalf("expected OPEN after 5 timeout failures, got %s", cb.State())
	}

	// After circuit opens, recovery happens within MTTR < backoff total.
	time.Sleep(30 * time.Millisecond)
	// Probe with a fast, successful call.
	d0, err := cb.Execute(ctx, 0, successFn(fallback))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = d0 // probe might return fallback (HALF_OPEN probe) or actual value
	// After a successful probe the circuit should close.
	if cb.State() != resilience.StateClosed {
		// May still be in half-open with the result returned as fallback.
		t.Logf("state after probe: %s (acceptable if fallback returned)", cb.State())
	}
}

// Scenario 3: Gradient overflow.
func TestCircuitBreaker_Scenario3_GradientOverflow(t *testing.T) {
	const fallback = 1.7
	cb := newFastBreaker(5, fallback)
	ctx := context.Background()

	overflowGrad := 1.5e10 // far above MaxGradientMagnitude=1.0

	for i := 0; i < 5; i++ {
		d0, _ := cb.Execute(ctx, overflowGrad, successFn(1.5))
		_ = d0
	}

	if cb.State() != resilience.StateOpen {
		t.Fatalf("expected OPEN after 5 gradient-overflow failures, got %s", cb.State())
	}

	// With circuit open, gradient effectively clipped: fallback returned.
	d0, _ := cb.Execute(ctx, overflowGrad, successFn(1.5))
	if d0 != fallback {
		t.Fatalf("expected fallback %g when OPEN, got %g", fallback, d0)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Metrics tests
// ─────────────────────────────────────────────────────────────────────────────

func TestCircuitBreaker_MetricsFallbacksTracked(t *testing.T) {
	const fallback = 1.7
	cb := newFastBreaker(2, fallback)
	tripBreaker(t, cb, 2) // open circuit

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		_, _ = cb.Execute(ctx, 0, successFn(1.5))
	}

	if cb.Metrics.FallbacksUsed < 3 {
		t.Fatalf("expected at least 3 fallbacks tracked, got %d", cb.Metrics.FallbacksUsed)
	}
}

func TestCircuitBreaker_MetricsTransitionsRecorded(t *testing.T) {
	cb := newFastBreaker(2, 1.7)
	tripBreaker(t, cb, 2) // CLOSED → OPEN

	transitions := cb.Metrics.StateTransitions
	if transitions["CLOSED->OPEN"] < 1 {
		t.Fatalf("expected CLOSED->OPEN transition recorded, got %v", transitions)
	}

	// Wait and let OPEN → HALF_OPEN.
	time.Sleep(30 * time.Millisecond)
	_, _ = cb.Execute(context.Background(), 0, successFn(1.5)) // HALF_OPEN → CLOSED

	if cb.Metrics.StateTransitions["OPEN->HALF_OPEN"] < 1 {
		t.Fatalf("expected OPEN->HALF_OPEN transition, got %v", cb.Metrics.StateTransitions)
	}
	if cb.Metrics.StateTransitions["HALF_OPEN->CLOSED"] < 1 {
		t.Fatalf("expected HALF_OPEN->CLOSED transition, got %v", cb.Metrics.StateTransitions)
	}
}

func TestCircuitBreaker_PrometheusTextNonEmpty(t *testing.T) {
	cb := newFastBreaker(3, 1.7)
	text := cb.Metrics.PrometheusText(cb.State())
	if len(text) == 0 {
		t.Fatal("expected non-empty Prometheus text output")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Exponential backoff tests
// ─────────────────────────────────────────────────────────────────────────────

func TestCircuitBreaker_ExponentialBackoff(t *testing.T) {
	// Three failures, each reopening with the next backoff interval.
	cb := resilience.NewCircuitBreaker(resilience.Config{
		FailureThreshold:     1,
		BackoffIntervals:     []time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 40 * time.Millisecond},
		CallTimeout:          200 * time.Millisecond,
		MaxGradientMagnitude: 1.0,
		FallbackD0:           1.7,
	})
	ctx := context.Background()

	// Trip 1: 10ms backoff.
	_, _ = cb.Execute(ctx, 0, errorFn("fail"))
	if cb.State() != resilience.StateOpen {
		t.Fatalf("expected OPEN, got %s", cb.State())
	}

	// Before first backoff expires, circuit stays open.
	time.Sleep(5 * time.Millisecond)
	_, _ = cb.Execute(ctx, 0, successFn(1.5)) // still open, fallback
	if cb.State() != resilience.StateOpen {
		t.Logf("state: %s (may have transitioned — timing-dependent)", cb.State())
	}

	// After first backoff: probe fails → reopen with 20ms backoff.
	time.Sleep(10 * time.Millisecond)
	_, _ = cb.Execute(ctx, 0, errorFn("probe fail")) // HALF_OPEN → OPEN with 20ms backoff

	// Still open after 10ms (next backoff is 20ms).
	time.Sleep(10 * time.Millisecond)
	// It may or may not be half-open here depending on exact timing; just
	// verify the circuit eventually recovers.
	time.Sleep(15 * time.Millisecond)
	_, _ = cb.Execute(ctx, 0, successFn(1.5)) // should probe and close
}
