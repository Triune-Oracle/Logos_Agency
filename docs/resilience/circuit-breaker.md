# Circuit Breaker for Non-Differentiable Fractal Operations

## Overview

The Box-Counting Dimension (D₀) calculation used by the LogosTalisman fractal
loss is non-differentiable and numerically unstable.  Without fault-tolerance
the training loop can crash when:

* The BCD algorithm encounters a **mathematical singularity** (e.g. a
  degenerate latent batch produces `NaN`).
* The calculation **times out** (> 5 seconds on slow hardware or large batches).
* A **gradient overflow** is detected (gradient magnitude > 1.0).

The circuit breaker in `pkg/resilience` wraps the BCD function and guarantees
**100% training continuity** by falling back to the target dimension
(D_F,target) during failures.  This makes the fractal loss component zero
during recovery, preserving loss-trajectory continuity.

---

## State Machine

```
         ┌──────────────────────────┐
         │                          │
  5+ failures                   success
         │                          │
         ▼                          │
  ┌────────────┐             ┌──────────────┐
  │   CLOSED   │──────────── │  HALF_OPEN   │
  └────────────┘             └──────────────┘
         │                          │
  5+ failures                   failure
         │                          │
         ▼                          ▼
  ┌────────────┐             ┌──────────────┐
  │    OPEN    │─backoff────▶│  HALF_OPEN   │
  └────────────┘             └──────────────┘
```

| Transition         | Trigger                                   |
|--------------------|-------------------------------------------|
| CLOSED → OPEN      | `FailureThreshold` consecutive failures   |
| OPEN → HALF_OPEN   | Backoff window expires (15s / 30s / 60s)  |
| HALF_OPEN → CLOSED | Probe call succeeds                       |
| HALF_OPEN → OPEN   | Probe call fails (backoff advances)       |

---

## Configuration

```go
cfg := resilience.Config{
    FailureThreshold:     5,                   // failures before OPEN
    BackoffIntervals:     []time.Duration{     // exponential retry windows
        15 * time.Second,
        30 * time.Second,
        60 * time.Second,
    },
    CallTimeout:          5 * time.Second,     // per-call timeout
    MaxGradientMagnitude: 1.0,                 // overflow threshold
    FallbackD0:           1.7,                 // D_F,target (MNIST default)
}
cb := resilience.NewCircuitBreaker(cfg)
```

---

## Usage

### Standalone

```go
d0, err := cb.Execute(ctx, gradientMagnitude, func(ctx context.Context) (float64, error) {
    return myBCDCalculation(ctx, latentBatch)
})
// d0 is always a valid float64 (never NaN/Inf):
// • actual D₀ when the circuit is CLOSED and the call succeeds
// • FallbackD0 when the circuit is OPEN or the call fails
```

### With FractalLoss

```go
import (
    "github.com/Triune-Oracle/Logos_Agency/pkg/fractal"
    "github.com/Triune-Oracle/Logos_Agency/pkg/resilience"
)

cb := resilience.NewCircuitBreaker(resilience.Config{FallbackD0: 1.7})
fl := fractal.NewFractalLoss(cb, 1.7 /* D_target */)

// In each training step:
result := fl.Compute(ctx, gradMagnitude, func(ctx context.Context) (float64, error) {
    return computeBoxCountingDimension(ctx, latentBatch)
})

fmt.Printf("fractal loss=%.4f  D₀=%.4f  fallback=%v\n",
    result.Loss, result.D0Used, result.Fallback)
```

---

## Failure Scenarios

### Scenario 1 – Mathematical Singularity

```go
// Epoch 30: inject NaN by returning math.NaN()
func injectSingularity(_ context.Context) (float64, error) {
    return math.NaN(), nil
}
```

**Behaviour:** NaN is detected in `isFailure`, failure counter increments.
After 5 consecutive NaN results the circuit opens.  Training continues with
`FallbackD0` returned for each subsequent call until the backoff expires.

### Scenario 2 – Computation Timeout

```go
// Epoch 60: simulate 6-second computation
func injectTimeout(ctx context.Context) (float64, error) {
    time.Sleep(6 * time.Second)
    return 1.5, nil
}
```

**Behaviour:** `Execute` wraps the call in a `context.WithTimeout(ctx, 5s)`.
The goroutine is cancelled after 5 s; the timeout is recorded as a failure.
After 5 timeouts the circuit opens.  MTTR ≤ 60 s (first backoff is 15 s,
successful probe closes the circuit).

### Scenario 3 – Gradient Overflow

```go
// Epoch 80: gradient magnitude >> 1.0
gradientMagnitude := 1.5e10
cb.Execute(ctx, gradientMagnitude, bcdFn)
```

**Behaviour:** `isFailure` checks `gradientMagnitude > MaxGradientMagnitude`
before inspecting the BCD result.  Even if the BCD calculation itself
succeeds, the overflow is counted as a failure and `FallbackD0` is returned,
effectively clipping the gradient contribution.

---

## Prometheus Metrics

Call `cb.Metrics.PrometheusText(cb.State())` to obtain a Prometheus
text-format snapshot:

```
# HELP fractal_circuit_breaker_state Current state (0=CLOSED,1=OPEN,2=HALF_OPEN)
# TYPE fractal_circuit_breaker_state gauge
fractal_circuit_breaker_state 0

# HELP fractal_circuit_breaker_fallbacks_total Total fallback values returned
# TYPE fractal_circuit_breaker_fallbacks_total counter
fractal_circuit_breaker_fallbacks_total 42

# HELP fractal_circuit_breaker_consecutive_failures Current consecutive failure count
# TYPE fractal_circuit_breaker_consecutive_failures gauge
fractal_circuit_breaker_consecutive_failures 0

fractal_circuit_breaker_transitions_total{transition="CLOSED->OPEN"} 3
fractal_circuit_breaker_transitions_total{transition="OPEN->HALF_OPEN"} 3
fractal_circuit_breaker_transitions_total{transition="HALF_OPEN->CLOSED"} 2
fractal_circuit_breaker_transitions_total{transition="HALF_OPEN->OPEN"} 1
```

Expose this endpoint from the existing `FractalExporter.handleMetrics` handler
to include circuit-breaker state in the same Prometheus scrape.

---

## Success Criteria

| Criterion                                   | Status |
|---------------------------------------------|--------|
| Zero training interruptions (fault injection) | ✅ Fallback always returns valid `float64` |
| MTTR < 60 s                                  | ✅ Max backoff is 60 s; probe closes circuit in one step |
| State transitions logged in Prometheus       | ✅ `Metrics.StateTransitions` map |
| Graceful degradation (loss continuity)       | ✅ Fallback loss = (D_F,target − D_target)² = 0 |

---

## References

* [Circuit Breaker Pattern – Martin Fowler](https://martinfowler.com/bliki/CircuitBreaker.html)
* [sony/gobreaker](https://github.com/sony/gobreaker)
* `pkg/resilience/circuit_breaker.go`
* `pkg/fractal/fallback_loss.go`
* `pkg/resilience/circuit_breaker_test.go`
