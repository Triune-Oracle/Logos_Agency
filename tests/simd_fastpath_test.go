package tests

import (
	"fmt"
	"math"
	"runtime"
	"testing"

	"github.com/Triune-Oracle/Logos_Agency/engine"
)

// ---- architecture detection ------------------------------------------------

// TestDetectSIMDCapabilities verifies that DetectSIMDCapabilities returns
// consistent values and that the fields are self-consistent.
func TestDetectSIMDCapabilities(t *testing.T) {
	caps := engine.DetectSIMDCapabilities()

	if caps.Arch == "" {
		t.Error("SIMDCapabilities.Arch must not be empty")
	}
	if caps.Arch != runtime.GOARCH {
		t.Errorf("Expected Arch %q, got %q", runtime.GOARCH, caps.Arch)
	}
	if caps.VectorWidth < 1 {
		t.Errorf("VectorWidth must be >= 1, got %d", caps.VectorWidth)
	}
	if caps.UnrollFactor < 1 {
		t.Errorf("UnrollFactor must be >= 1, got %d", caps.UnrollFactor)
	}

	switch runtime.GOARCH {
	case "amd64":
		if !caps.Available {
			t.Error("Expected Available=true on amd64")
		}
		if caps.VectorWidth != 4 {
			t.Errorf("Expected VectorWidth=4 on amd64, got %d", caps.VectorWidth)
		}
		if caps.UnrollFactor != 8 {
			t.Errorf("Expected UnrollFactor=8 on amd64, got %d", caps.UnrollFactor)
		}
	case "arm64":
		if !caps.Available {
			t.Error("Expected Available=true on arm64")
		}
		if caps.VectorWidth != 2 {
			t.Errorf("Expected VectorWidth=2 on arm64, got %d", caps.VectorWidth)
		}
		if caps.UnrollFactor != 4 {
			t.Errorf("Expected UnrollFactor=4 on arm64, got %d", caps.UnrollFactor)
		}
	default:
		if caps.Available {
			t.Errorf("Expected Available=false on %s, got true", runtime.GOARCH)
		}
		if caps.VectorWidth != 1 {
			t.Errorf("Expected VectorWidth=1 on %s, got %d", runtime.GOARCH, caps.VectorWidth)
		}
		if caps.UnrollFactor != 1 {
			t.Errorf("Expected UnrollFactor=1 on %s, got %d", runtime.GOARCH, caps.UnrollFactor)
		}
	}
}

// TestIsSIMDAvailable verifies that IsSIMDAvailable is consistent with
// DetectSIMDCapabilities.Available.
func TestIsSIMDAvailable(t *testing.T) {
	caps := engine.DetectSIMDCapabilities()
	got := engine.IsSIMDAvailable()
	if got != caps.Available {
		t.Errorf("IsSIMDAvailable()=%v but DetectSIMDCapabilities().Available=%v", got, caps.Available)
	}
}

// TestGetSIMDCapabilities verifies GetSIMDCapabilities returns the same value
// as DetectSIMDCapabilities.
func TestGetSIMDCapabilities(t *testing.T) {
	want := engine.DetectSIMDCapabilities()
	got := engine.GetSIMDCapabilities()
	if got != want {
		t.Errorf("GetSIMDCapabilities() = %+v, want %+v", got, want)
	}
}

// ---- CalculateStatsScalar correctness  ------------------------------------

func TestCalculateStatsScalar_Correctness(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		wantMean float64
		wantMin  float64
		wantMax  float64
		wantN    int
	}{
		{
			name:     "single element",
			values:   []float64{42.0},
			wantMean: 42.0,
			wantMin:  42.0,
			wantMax:  42.0,
			wantN:    1,
		},
		{
			name:     "two elements",
			values:   []float64{1.0, 3.0},
			wantMean: 2.0,
			wantMin:  1.0,
			wantMax:  3.0,
			wantN:    2,
		},
		{
			name:     "sequential integers",
			values:   []float64{1, 2, 3, 4, 5},
			wantMean: 3.0,
			wantMin:  1.0,
			wantMax:  5.0,
			wantN:    5,
		},
		{
			name:     "all same value",
			values:   []float64{7, 7, 7, 7},
			wantMean: 7.0,
			wantMin:  7.0,
			wantMax:  7.0,
			wantN:    4,
		},
		{
			name:     "negative values",
			values:   []float64{-3, -1, 0, 1, 3},
			wantMean: 0.0,
			wantMin:  -3.0,
			wantMax:  3.0,
			wantN:    5,
		},
		{
			name:     "non-divisible by unroll factor (9 elements)",
			values:   []float64{1, 2, 3, 4, 5, 6, 7, 8, 9},
			wantMean: 5.0,
			wantMin:  1.0,
			wantMax:  9.0,
			wantN:    9,
		},
		{
			name:     "size exactly equals unroll factor (8)",
			values:   []float64{1, 2, 3, 4, 5, 6, 7, 8},
			wantMean: 4.5,
			wantMin:  1.0,
			wantMax:  8.0,
			wantN:    8,
		},
	}

	const tol = 1e-9

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := engine.CalculateStatsScalar(tc.values)

			if got.Count != tc.wantN {
				t.Errorf("Count: got %d, want %d", got.Count, tc.wantN)
			}
			if math.Abs(got.Mean-tc.wantMean) > tol {
				t.Errorf("Mean: got %g, want %g", got.Mean, tc.wantMean)
			}
			if math.Abs(got.Min-tc.wantMin) > tol {
				t.Errorf("Min: got %g, want %g", got.Min, tc.wantMin)
			}
			if math.Abs(got.Max-tc.wantMax) > tol {
				t.Errorf("Max: got %g, want %g", got.Max, tc.wantMax)
			}
			if got.Variance < 0 {
				t.Errorf("Variance must be non-negative, got %g", got.Variance)
			}
			if got.StdDev < 0 {
				t.Errorf("StdDev must be non-negative, got %g", got.StdDev)
			}
			// StdDev = sqrt(Variance)
			if got.Variance > 0 {
				wantStdDev := math.Sqrt(got.Variance)
				if math.Abs(got.StdDev-wantStdDev) > tol {
					t.Errorf("StdDev: got %g, want %g", got.StdDev, wantStdDev)
				}
			}
		})
	}
}

// TestCalculateStatsScalar_EmptyInput verifies graceful handling of an empty slice.
func TestCalculateStatsScalar_EmptyInput(t *testing.T) {
	got := engine.CalculateStatsScalar([]float64{})
	if got.Count != 0 {
		t.Errorf("Expected Count=0 for empty input, got %d", got.Count)
	}
	if got.Mean != 0 || got.Variance != 0 || got.StdDev != 0 || got.Min != 0 || got.Max != 0 {
		t.Errorf("Expected zero-value stats for empty input, got %+v", got)
	}
}

// ---- CalculateStatsSIMD correctness  --------------------------------------

func TestCalculateStatsSIMD_Correctness(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		wantMean float64
		wantMin  float64
		wantMax  float64
		wantN    int
	}{
		{
			name:     "single element",
			values:   []float64{42.0},
			wantMean: 42.0,
			wantMin:  42.0,
			wantMax:  42.0,
			wantN:    1,
		},
		{
			name:     "sequential integers",
			values:   []float64{1, 2, 3, 4, 5},
			wantMean: 3.0,
			wantMin:  1.0,
			wantMax:  5.0,
			wantN:    5,
		},
		{
			name:     "all same value",
			values:   []float64{7, 7, 7, 7},
			wantMean: 7.0,
			wantMin:  7.0,
			wantMax:  7.0,
			wantN:    4,
		},
		{
			name:     "negative values",
			values:   []float64{-3, -1, 0, 1, 3},
			wantMean: 0.0,
			wantMin:  -3.0,
			wantMax:  3.0,
			wantN:    5,
		},
		{
			name:     "size exactly 8 (amd64 unroll boundary)",
			values:   []float64{1, 2, 3, 4, 5, 6, 7, 8},
			wantMean: 4.5,
			wantMin:  1.0,
			wantMax:  8.0,
			wantN:    8,
		},
		{
			name:     "9 elements (one scalar tail on amd64)",
			values:   []float64{1, 2, 3, 4, 5, 6, 7, 8, 9},
			wantMean: 5.0,
			wantMin:  1.0,
			wantMax:  9.0,
			wantN:    9,
		},
		{
			name:     "100 elements",
			values:   func() []float64 { s := make([]float64, 100); for i := range s { s[i] = float64(i + 1) }; return s }(),
			wantMean: 50.5,
			wantMin:  1.0,
			wantMax:  100.0,
			wantN:    100,
		},
	}

	const tol = 1e-9

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := engine.CalculateStatsSIMD(tc.values)

			if got.Count != tc.wantN {
				t.Errorf("Count: got %d, want %d", got.Count, tc.wantN)
			}
			if math.Abs(got.Mean-tc.wantMean) > tol {
				t.Errorf("Mean: got %g, want %g", got.Mean, tc.wantMean)
			}
			if math.Abs(got.Min-tc.wantMin) > tol {
				t.Errorf("Min: got %g, want %g", got.Min, tc.wantMin)
			}
			if math.Abs(got.Max-tc.wantMax) > tol {
				t.Errorf("Max: got %g, want %g", got.Max, tc.wantMax)
			}
			if got.Variance < 0 {
				t.Errorf("Variance must be non-negative, got %g", got.Variance)
			}
		})
	}
}

func TestCalculateStatsSIMD_EmptyInput(t *testing.T) {
	got := engine.CalculateStatsSIMD([]float64{})
	if got.Count != 0 {
		t.Errorf("Expected Count=0 for empty input, got %d", got.Count)
	}
}

// ---- SIMD / scalar numerical equivalence  ---------------------------------

// TestSIMDScalarEquivalence_Stats verifies that CalculateStatsSIMD and
// CalculateStatsScalar produce results that are within a small floating-point
// tolerance for a wide range of inputs. This satisfies the determinism
// requirement: both paths are reproducible and numerically equivalent.
//
// Note: SIMD paths use independent accumulators with tree reduction, so they
// may differ from the scalar sequential accumulation by a few ULPs — this is
// documented and expected behaviour for floating-point parallelism.
func TestSIMDScalarEquivalence_Stats(t *testing.T) {
	const relativeTol = 1e-10 // relative tolerance for FP equivalence

	inputs := []struct {
		name   string
		values []float64
	}{
		{"single", []float64{1.0}},
		{"two", []float64{1.0, 2.0}},
		{"four", []float64{1, 2, 3, 4}},
		{"eight", makeSeq(8)},
		{"nine", makeSeq(9)},
		{"16", makeSeq(16)},
		{"100", makeSeq(100)},
		{"1000", makeSeq(1000)},
		{"10000", makeSeq(10000)},
		{"negatives", []float64{-100, -50, 0, 50, 100}},
		{"all_same", makeConst(256, 3.14)},
		{"alternating", makeAlternating(200)},
	}

	for _, tc := range inputs {
		t.Run(tc.name, func(t *testing.T) {
			simd := engine.CalculateStatsSIMD(tc.values)
			scalar := engine.CalculateStatsScalar(tc.values)

			if simd.Count != scalar.Count {
				t.Errorf("Count mismatch: SIMD=%d scalar=%d", simd.Count, scalar.Count)
			}
			assertRelClose(t, "Mean", simd.Mean, scalar.Mean, relativeTol)
			assertRelClose(t, "Min", simd.Min, scalar.Min, relativeTol)
			assertRelClose(t, "Max", simd.Max, scalar.Max, relativeTol)
			assertRelClose(t, "Variance", simd.Variance, scalar.Variance, relativeTol)
			assertRelClose(t, "StdDev", simd.StdDev, scalar.StdDev, relativeTol)
		})
	}
}

// TestSIMDScalarEquivalence_Stats_Determinism verifies that calling the same
// path twice with the same input always returns the same result (no randomness
// or state leakage between calls).
func TestSIMDScalarEquivalence_Stats_Determinism(t *testing.T) {
	values := makeSeq(10000)

	simd1 := engine.CalculateStatsSIMD(values)
	simd2 := engine.CalculateStatsSIMD(values)
	scalar1 := engine.CalculateStatsScalar(values)
	scalar2 := engine.CalculateStatsScalar(values)

	if simd1 != simd2 {
		t.Error("CalculateStatsSIMD is not deterministic (results differ between calls)")
	}
	if scalar1 != scalar2 {
		t.Error("CalculateStatsScalar is not deterministic (results differ between calls)")
	}
}

// ---- CalculateLikelihoodsSIMD correctness  ---------------------------------

func TestCalculateLikelihoodsSIMD_Correctness(t *testing.T) {
	tests := []struct {
		name         string
		values       []string
		wantIntMin   float64 // minimum expected integer likelihood
		wantFloatMax float64 // maximum expected float likelihood (for pure int data)
	}{
		{
			name:       "all integers",
			values:     generateIntegerData(200),
			wantIntMin: 0.9,
		},
		{
			name:         "all floats",
			values:       generateFloatData(200),
			wantFloatMax: 1.0,
		},
		{
			name: "all strings",
			values: func() []string {
				s := make([]string, 200)
				for i := range s {
					s[i] = fmt.Sprintf("word_%d", i)
				}
				return s
			}(),
		},
		{
			name: "all booleans",
			values: func() []string {
				s := make([]string, 200)
				for i := range s {
					if i%2 == 0 {
						s[i] = "true"
					} else {
						s[i] = "false"
					}
				}
				return s
			}(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			likelihoods := engine.CalculateLikelihoodsSIMD(tc.values)

			if len(likelihoods) == 0 {
				t.Error("Expected non-empty likelihoods map")
			}

			// All likelihoods must be in [0, 1].
			for typ, v := range likelihoods {
				if v < 0 || v > 1+1e-9 {
					t.Errorf("Likelihood for type %v = %g, must be in [0,1]", typ, v)
				}
			}

			if tc.wantIntMin > 0 {
				got := likelihoods[engine.TypeInteger]
				if got < tc.wantIntMin {
					t.Errorf("Integer likelihood = %g, want >= %g", got, tc.wantIntMin)
				}
			}
		})
	}
}

// TestSIMDScalarEquivalence_Likelihoods verifies that CalculateLikelihoodsSIMD
// and the simple scalar path produce identical type rankings for the same input.
func TestSIMDScalarEquivalence_Likelihoods(t *testing.T) {
	// Use datasets larger than 100 values to ensure both paths are exercised.
	inputs := []struct {
		name   string
		values []string
	}{
		{"integers_200", generateIntegerData(200)},
		{"floats_200", generateFloatData(200)},
		{"strings_200", generateStringData(200)},
		{"dates_200", generateDateData(200)},
		{"mixed_200", generateMixedData(200)},
		{"integers_1000", generateIntegerData(1000)},
	}

	// We compare SIMD results with a BayesianInferenceEngine in SIMD-disabled
	// mode, which uses the scalar likelihood path directly.
	scalarConfig := engine.DefaultConfig()
	scalarConfig.EnableSIMD = false
	// Use a large SampleSize so no sampling occurs on our 1000-element inputs.
	scalarConfig.SampleSize = 10000

	simdConfig := engine.DefaultConfig()
	simdConfig.EnableSIMD = true
	simdConfig.SampleSize = 10000

	for _, tc := range inputs {
		t.Run(tc.name, func(t *testing.T) {
			scalarEng := engine.NewBayesianInferenceEngine(scalarConfig)
			simdEng := engine.NewBayesianInferenceEngine(simdConfig)

			scalarResult := scalarEng.InferType(tc.values)
			simdResult := simdEng.InferType(tc.values)

			// The inferred type must agree.
			if scalarResult.InferredType != simdResult.InferredType {
				t.Errorf("InferredType mismatch: scalar=%v SIMD=%v",
					scalarResult.InferredType, simdResult.InferredType)
			}
		})
	}
}

// ---- ParallelLikelihoodBatch  -----------------------------------------------

func TestParallelLikelihoodBatch_Correctness(t *testing.T) {
	columns := [][]string{
		generateIntegerData(200),
		generateFloatData(200),
		generateStringData(200),
	}

	results := engine.ParallelLikelihoodBatch(columns)

	if len(results) != len(columns) {
		t.Fatalf("Expected %d results, got %d", len(columns), len(results))
	}
	for i, r := range results {
		if len(r) == 0 {
			t.Errorf("Column %d: expected non-empty likelihoods map", i)
		}
		for typ, v := range r {
			if v < 0 || v > 1+1e-9 {
				t.Errorf("Column %d type %v: likelihood %g out of [0,1]", i, typ, v)
			}
		}
	}
}

// TestParallelLikelihoodBatch_SmallColumns verifies the scalar fallback path
// used for columns with <= 100 elements.
func TestParallelLikelihoodBatch_SmallColumns(t *testing.T) {
	small := generateIntegerData(50)
	results := engine.ParallelLikelihoodBatch([][]string{small})
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if len(results[0]) == 0 {
		t.Error("Expected non-empty likelihoods for small column")
	}
}

// ---- Variance / StdDev known values  ---------------------------------------

// TestCalculateStatsSIMD_KnownVariance checks variance against analytic values.
// For data [1, 2, 3, 4, 5], population variance = 2.0.
func TestCalculateStatsSIMD_KnownVariance(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5}
	got := engine.CalculateStatsSIMD(values)

	const wantVariance = 2.0
	const wantStdDev = math.Sqrt2 // sqrt(2) ≈ 1.41421356…
	const tol = 1e-9

	if math.Abs(got.Variance-wantVariance) > tol {
		t.Errorf("Variance: got %g, want %g", got.Variance, wantVariance)
	}
	if math.Abs(got.StdDev-wantStdDev) > tol {
		t.Errorf("StdDev: got %g, want %g", got.StdDev, wantStdDev)
	}
}

// TestCalculateStatsScalar_KnownVariance mirrors the above for the scalar path.
func TestCalculateStatsScalar_KnownVariance(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5}
	got := engine.CalculateStatsScalar(values)

	const wantVariance = 2.0
	const wantStdDev = math.Sqrt2
	const tol = 1e-9

	if math.Abs(got.Variance-wantVariance) > tol {
		t.Errorf("Variance: got %g, want %g", got.Variance, wantVariance)
	}
	if math.Abs(got.StdDev-wantStdDev) > tol {
		t.Errorf("StdDev: got %g, want %g", got.StdDev, wantStdDev)
	}
}

// ---- zero-variance edge case  -----------------------------------------------

// TestCalculateStatsSIMD_ZeroVariance verifies that all-equal inputs produce
// zero variance and zero standard deviation.
func TestCalculateStatsSIMD_ZeroVariance(t *testing.T) {
	for _, n := range []int{1, 4, 8, 9, 100} {
		t.Run(fmt.Sprintf("n=%d", n), func(t *testing.T) {
			values := makeConst(n, 5.5)
			simd := engine.CalculateStatsSIMD(values)
			scalar := engine.CalculateStatsScalar(values)

			for label, got := range map[string]engine.VectorizedStatistics{
				"simd": simd, "scalar": scalar,
			} {
				if got.Variance != 0 {
					t.Errorf("%s: Variance=%g, want 0 for all-equal input", label, got.Variance)
				}
				if got.StdDev != 0 {
					t.Errorf("%s: StdDev=%g, want 0 for all-equal input", label, got.StdDev)
				}
			}
		})
	}
}

// ---- helpers ---------------------------------------------------------------

// assertRelClose fails the test if the values differ by more than tol relative
// to |want|. For near-zero values (|want| < 1e-6), an absolute tolerance of
// 1e-6 is used instead to avoid false positives from floating-point rounding.
func assertRelClose(t *testing.T, label string, got, want, tol float64) {
	t.Helper()
	diff := math.Abs(got - want)
	// Absolute tolerance fallback: when both values are essentially zero,
	// small absolute differences are acceptable (FP rounding in variance
	// computation for constant-value inputs, for example).
	if math.Abs(want) < 1e-6 && diff < 1e-6 {
		return
	}
	denom := math.Abs(want)
	if denom < 1 {
		denom = 1
	}
	if diff/denom > tol {
		t.Errorf("%s: got %g, want %g (rel diff %g > tol %g)", label, got, want, diff/denom, tol)
	}
}

// makeSeq returns a slice [1, 2, …, n] as float64.
func makeSeq(n int) []float64 {
	s := make([]float64, n)
	for i := range s {
		s[i] = float64(i + 1)
	}
	return s
}

// makeConst returns a slice of n copies of value.
func makeConst(n int, value float64) []float64 {
	s := make([]float64, n)
	for i := range s {
		s[i] = value
	}
	return s
}

// makeAlternating returns a slice that alternates between +1 and -1.
func makeAlternating(n int) []float64 {
	s := make([]float64, n)
	for i := range s {
		if i%2 == 0 {
			s[i] = 1.0
		} else {
			s[i] = -1.0
		}
	}
	return s
}
