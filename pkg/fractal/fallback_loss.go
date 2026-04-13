// Package fractal provides loss functions for fractal-constrained VAE training.
//
// FallbackLoss supplies a stable loss value when the Box-Counting Dimension
// calculation is unavailable (e.g., circuit breaker is OPEN).
package fractal

import (
	"context"
	"math"

	"github.com/Triune-Oracle/Logos_Agency/pkg/resilience"
)

// LossResult holds the outcome of a fractal-loss evaluation.
type LossResult struct {
	// Loss is the computed fractal loss value.
	Loss float64

	// D0Used is the D₀ value (measured or fallback) used to compute the loss.
	D0Used float64

	// Fallback is true when the circuit was OPEN and D_F,target was used instead
	// of a freshly computed D₀.
	Fallback bool
}

// FractalLoss wraps a Box-Counting Dimension function with a circuit breaker so
// that training continues uninterrupted when the BCD calculation fails.
//
// On success the loss is:
//
//	L = (D₀ - D_target)²
//
// When the circuit is OPEN the fallback loss is:
//
//	L = (D_F,target - D_target)² = 0   (because D_F,target == D_target)
//
// This guarantees that the fractal component of the total loss is zero (not
// noisy) during recovery, preserving loss-trajectory continuity.
type FractalLoss struct {
	cb      *resilience.CircuitBreaker
	dTarget float64
}

// NewFractalLoss creates a FractalLoss that uses cb to protect calls to the BCD
// function.  dTarget is the training target fractal dimension D_F,target.
func NewFractalLoss(cb *resilience.CircuitBreaker, dTarget float64) *FractalLoss {
	return &FractalLoss{cb: cb, dTarget: dTarget}
}

// Compute evaluates the fractal loss for one training step.
//
// bcdFn should compute the Box-Counting Dimension from the current latent batch
// and return it along with any error.  gradientMagnitude is the L2-norm of the
// gradient at the BCD calculation site (pass 0 if not tracked).
func (fl *FractalLoss) Compute(
	ctx context.Context,
	gradientMagnitude float64,
	bcdFn func(ctx context.Context) (float64, error),
) LossResult {
	d0, _ := fl.cb.Execute(ctx, gradientMagnitude, bcdFn)

	fallback := fl.cb.State() == resilience.StateOpen ||
		(math.IsNaN(d0) || math.IsInf(d0, 0))

	// Guard against any residual NaN/Inf that slipped through.
	if math.IsNaN(d0) || math.IsInf(d0, 0) {
		d0 = fl.dTarget
		fallback = true
	}

	diff := d0 - fl.dTarget
	loss := diff * diff

	return LossResult{
		Loss:     loss,
		D0Used:   d0,
		Fallback: fallback,
	}
}
