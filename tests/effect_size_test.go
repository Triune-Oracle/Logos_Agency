package tests

import (
	"math"
	"testing"

	"github.com/Triune-Oracle/Logos_Agency/engine"
)

// Phase II validation datasets used throughout the statistical supplement.
var (
	mnistPSNRLogos    = []float64{24.5, 24.7, 24.4, 24.8, 24.6}
	mnistPSNRBaseline = []float64{19.1, 19.3, 19.0, 19.4, 19.2}

	mnistSSIMLogos    = []float64{0.90, 0.92, 0.89, 0.93, 0.91}
	mnistSSIMBaseline = []float64{0.81, 0.83, 0.80, 0.84, 0.82}

	cifar10PSNRLogos    = []float64{26.8, 27.3, 26.9, 27.5, 27.0}
	cifar10PSNRBaseline = []float64{20.9, 21.5, 21.1, 21.8, 21.2}

	cifar10SSIMLogos    = []float64{0.86, 0.90, 0.86, 0.90, 0.88}
	cifar10SSIMBaseline = []float64{0.73, 0.79, 0.73, 0.79, 0.76}

	cifar10FIDLogos    = []float64{49.3, 54.8, 51.6, 55.9, 51.0}
	cifar10FIDBaseline = []float64{83.5, 92.1, 86.7, 96.3, 87.6}
)

// approxEqual returns true when a and b are within tol of each other.
func approxEqual(a, b, tol float64) bool {
	return math.Abs(a-b) <= tol
}

// ── SliceMean ────────────────────────────────────────────────────────────────

func TestSliceMean_Basic(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5}
	got := engine.SliceMean(data)
	if !approxEqual(got, 3.0, 1e-9) {
		t.Errorf("SliceMean: expected 3.0, got %v", got)
	}
}

func TestSliceMean_Empty(t *testing.T) {
	got := engine.SliceMean([]float64{})
	if got != 0 {
		t.Errorf("SliceMean(empty): expected 0, got %v", got)
	}
}

func TestSliceMean_MNISTPSNRBaseline(t *testing.T) {
	got := engine.SliceMean(mnistPSNRBaseline)
	if !approxEqual(got, 19.2, 1e-9) {
		t.Errorf("Mean of MNIST baseline PSNR: expected 19.2, got %v", got)
	}
}

// ── SliceVariance ─────────────────────────────────────────────────────────────

func TestSliceVariance_Known(t *testing.T) {
	// Variance of {2, 4, 4, 4, 5, 5, 7, 9} = 4.571...
	data := []float64{2, 4, 4, 4, 5, 5, 7, 9}
	got := engine.SliceVariance(data)
	// Population variance is 4.0; sample variance (ddof=1) = 32/7 ≈ 4.5714
	expected := 32.0 / 7.0
	if !approxEqual(got, expected, 1e-9) {
		t.Errorf("SliceVariance: expected %v, got %v", expected, got)
	}
}

func TestSliceVariance_SingleElement(t *testing.T) {
	got := engine.SliceVariance([]float64{42})
	if got != 0 {
		t.Errorf("SliceVariance(single): expected 0, got %v", got)
	}
}

// ── SliceStdDev ───────────────────────────────────────────────────────────────

func TestSliceStdDev_MNISTPSNRLogos(t *testing.T) {
	got := engine.SliceStdDev(mnistPSNRLogos)
	// Manual: deviations from 24.6 are ±0.1, ±0.2; SS = 0.10; s² = 0.025; s = 0.15811
	if !approxEqual(got, 0.15811, 1e-4) {
		t.Errorf("StdDev MNIST Logos PSNR: expected ~0.1581, got %v", got)
	}
}

// ── SlicePooledStdDev ─────────────────────────────────────────────────────────

func TestSlicePooledStdDev_EqualVariances(t *testing.T) {
	g1 := []float64{1, 3, 5}
	g2 := []float64{2, 4, 6}
	// Variance of g1 = 4, variance of g2 = 4; pooled = sqrt(4) = 2
	got := engine.SlicePooledStdDev(g1, g2)
	if !approxEqual(got, 2.0, 1e-9) {
		t.Errorf("PooledStdDev equal variance: expected 2.0, got %v", got)
	}
}

// ── CohensD ───────────────────────────────────────────────────────────────────

func TestCohensD_MNIST_PSNR(t *testing.T) {
	d := engine.CohensD(mnistPSNRLogos, mnistPSNRBaseline)
	// mean diff = 5.4, pooled std ≈ 0.15811 → d ≈ 34.15
	if d < 30 {
		t.Errorf("CohensD MNIST PSNR: expected large positive effect, got %v", d)
	}
}

func TestCohensD_CIFAR10_PSNR(t *testing.T) {
	d := engine.CohensD(cifar10PSNRLogos, cifar10PSNRBaseline)
	if d < 10 {
		t.Errorf("CohensD CIFAR-10 PSNR: expected large positive effect, got %v", d)
	}
}

func TestCohensD_MNIST_SSIM(t *testing.T) {
	d := engine.CohensD(mnistSSIMLogos, mnistSSIMBaseline)
	if d < 5 {
		t.Errorf("CohensD MNIST SSIM: expected large positive effect, got %v", d)
	}
}

func TestCohensD_CIFAR10_SSIM(t *testing.T) {
	d := engine.CohensD(cifar10SSIMLogos, cifar10SSIMBaseline)
	if d < 3 {
		t.Errorf("CohensD CIFAR-10 SSIM: expected large positive effect, got %v", d)
	}
}

func TestCohensD_CIFAR10_FID(t *testing.T) {
	// Logos FID < Baseline FID (lower is better), so d should be negative
	d := engine.CohensD(cifar10FIDLogos, cifar10FIDBaseline)
	if d > -5 {
		t.Errorf("CohensD CIFAR-10 FID: expected large negative effect, got %v", d)
	}
}

func TestCohensD_ZeroPooledStd(t *testing.T) {
	g1 := []float64{5, 5, 5}
	g2 := []float64{5, 5, 5}
	d := engine.CohensD(g1, g2)
	if d != 0 {
		t.Errorf("CohensD identical groups: expected 0, got %v", d)
	}
}

// ── WelchTTest ────────────────────────────────────────────────────────────────

func TestWelchTTest_MNIST_PSNR(t *testing.T) {
	res := engine.WelchTTest(mnistPSNRLogos, mnistPSNRBaseline)

	// t-statistic should be large and positive
	if res.TStatistic < 20 {
		t.Errorf("WelchTTest MNIST PSNR t-stat: expected > 20, got %v", res.TStatistic)
	}

	// Degrees of freedom: both groups have equal variance → df ≈ 8
	if res.DegreesOfFreedom < 6 || res.DegreesOfFreedom > 10 {
		t.Errorf("WelchTTest MNIST PSNR df: expected ~8, got %v", res.DegreesOfFreedom)
	}

	// p-value must be highly significant
	if res.PValueTwoSided >= 0.001 {
		t.Errorf("WelchTTest MNIST PSNR p-value: expected < 0.001, got %v", res.PValueTwoSided)
	}

	// Cohen's d must match direct computation
	direct := engine.CohensD(mnistPSNRLogos, mnistPSNRBaseline)
	if !approxEqual(res.CohensD, direct, 1e-9) {
		t.Errorf("WelchTTest CohensD mismatch: %v vs %v", res.CohensD, direct)
	}
}

func TestWelchTTest_CIFAR10_PSNR(t *testing.T) {
	res := engine.WelchTTest(cifar10PSNRLogos, cifar10PSNRBaseline)

	if res.TStatistic < 10 {
		t.Errorf("WelchTTest CIFAR-10 PSNR t-stat: expected > 10, got %v", res.TStatistic)
	}
	if res.PValueTwoSided >= 0.001 {
		t.Errorf("WelchTTest CIFAR-10 PSNR p-value: expected < 0.001, got %v", res.PValueTwoSided)
	}
}

func TestWelchTTest_MNIST_SSIM(t *testing.T) {
	res := engine.WelchTTest(mnistSSIMLogos, mnistSSIMBaseline)
	if res.PValueTwoSided >= 0.001 {
		t.Errorf("WelchTTest MNIST SSIM p-value: expected < 0.001, got %v", res.PValueTwoSided)
	}
}

func TestWelchTTest_CIFAR10_SSIM(t *testing.T) {
	res := engine.WelchTTest(cifar10SSIMLogos, cifar10SSIMBaseline)
	if res.PValueTwoSided >= 0.001 {
		t.Errorf("WelchTTest CIFAR-10 SSIM p-value: expected < 0.001, got %v", res.PValueTwoSided)
	}
}

func TestWelchTTest_CIFAR10_FID(t *testing.T) {
	// FID: LogosTalisman lower (better), so t-stat is negative
	res := engine.WelchTTest(cifar10FIDLogos, cifar10FIDBaseline)
	if res.TStatistic > 0 {
		t.Errorf("WelchTTest CIFAR-10 FID: expected negative t-stat, got %v", res.TStatistic)
	}
	if res.PValueTwoSided >= 0.001 {
		t.Errorf("WelchTTest CIFAR-10 FID p-value: expected < 0.001, got %v", res.PValueTwoSided)
	}
}

func TestWelchTTest_InsufficientData(t *testing.T) {
	res := engine.WelchTTest([]float64{1}, []float64{2, 3})
	if res.TStatistic != 0 || res.PValueTwoSided != 0 {
		t.Errorf("WelchTTest insufficient data: expected zero result, got %+v", res)
	}
}

func TestWelchTTest_MeanDiff(t *testing.T) {
	res := engine.WelchTTest(mnistPSNRLogos, mnistPSNRBaseline)
	if !approxEqual(res.MeanDiff, 5.4, 1e-9) {
		t.Errorf("WelchTTest MeanDiff: expected 5.4, got %v", res.MeanDiff)
	}
}

func TestWelchTTest_OneSidedPValue(t *testing.T) {
	res := engine.WelchTTest(mnistPSNRLogos, mnistPSNRBaseline)
	// One-sided p should be half of two-sided p for a symmetric t-distribution
	if !approxEqual(res.PValueOneSided*2, res.PValueTwoSided, 1e-12) {
		t.Errorf("One-sided p should be half of two-sided p: one=%v two=%v",
			res.PValueOneSided, res.PValueTwoSided)
	}
}

// ── regularizedIncompleteBeta (via t-distribution) ───────────────────────────

func TestTDistPValue_LargeT(t *testing.T) {
	// For very large t, p-value should approach 0
	res := engine.WelchTTest(
		[]float64{100, 101, 100, 101, 100},
		[]float64{1, 2, 1, 2, 1},
	)
	if res.PValueTwoSided > 1e-10 {
		t.Errorf("Large t-stat should yield near-zero p-value, got %v", res.PValueTwoSided)
	}
}

func TestTDistPValue_SmallT(t *testing.T) {
	// Groups with overlapping distributions → small t, large p-value
	g1 := []float64{10.0, 10.1, 9.9, 10.0, 10.1}
	g2 := []float64{10.0, 10.0, 10.1, 9.9, 10.0}
	res := engine.WelchTTest(g1, g2)
	if res.PValueTwoSided < 0.1 {
		t.Errorf("Similar groups should yield large p-value, got %v", res.PValueTwoSided)
	}
}

// ── Bonferroni correction sanity check ───────────────────────────────────────

func TestBonferroniCorrection_AllSignificant(t *testing.T) {
	// All 6 quality metric comparisons should remain significant
	// after Bonferroni correction (α' = 0.05/6 ≈ 0.0083)
	alpha := 0.05 / 6.0
	benchmarks := []struct {
		name   string
		logos  []float64
		base   []float64
	}{
		{"MNIST PSNR", mnistPSNRLogos, mnistPSNRBaseline},
		{"MNIST SSIM", mnistSSIMLogos, mnistSSIMBaseline},
		{"CIFAR-10 PSNR", cifar10PSNRLogos, cifar10PSNRBaseline},
		{"CIFAR-10 SSIM", cifar10SSIMLogos, cifar10SSIMBaseline},
	}

	for _, bm := range benchmarks {
		res := engine.WelchTTest(bm.logos, bm.base)
		if res.PValueTwoSided >= alpha {
			t.Errorf("%s: p=%v is NOT significant after Bonferroni correction (α'=%v)",
				bm.name, res.PValueTwoSided, alpha)
		}
	}
}
