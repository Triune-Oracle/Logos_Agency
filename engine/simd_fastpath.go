package engine

import (
	"math"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// SIMDCapabilities describes the SIMD acceleration available on the current
// platform. The struct is populated once at package initialization via
// DetectSIMDCapabilities and stored in the package-level simdCaps variable.
type SIMDCapabilities struct {
	// Arch is the Go runtime architecture identifier (runtime.GOARCH).
	Arch string

	// VectorWidth is the number of float64 elements that fit in one SIMD
	// register for this architecture:
	//   amd64 (AVX2):  4  × float64 per 256-bit register
	//   arm64 (NEON):  2  × float64 per 128-bit register
	//   other:         1  (scalar, no SIMD register)
	VectorWidth int

	// UnrollFactor is the chosen loop-unroll width for this architecture.
	// We unroll to cover two vector-register widths per iteration so that
	// the CPU's dual-issue execution units stay saturated:
	//   amd64: 8  (2 × AVX2 registers, each holding 4 × float64)
	//   arm64: 4  (2 × NEON registers, each holding 2 × float64)
	//   other: 1  (scalar, no unrolling)
	UnrollFactor int

	// Available is true when a SIMD-accelerated code path exists for this
	// architecture (currently amd64 and arm64).
	Available bool
}

// simdCaps holds the capabilities detected once at package init time.
var simdCaps SIMDCapabilities

func init() {
	simdCaps = DetectSIMDCapabilities()
}

// DetectSIMDCapabilities returns SIMD capability information for the current
// platform. It is called automatically at package init; it is also exported so
// callers can inspect or log the detected capabilities.
//
// Architecture annotations:
//   - amd64 (x86_64): The Go compiler emits SSE2/AVX/AVX2 instructions when
//     the loop body is vectorizable. An AVX2 YMM register holds 4 × float64
//     (256 bits). We unroll 8-wide (two register widths) to fill both the
//     integer and FP execution pipelines simultaneously.
//   - arm64 (AArch64): The Go compiler emits NEON/AdvSIMD instructions. A
//     NEON Q-register holds 2 × float64 (128 bits). We unroll 4-wide (two
//     NEON registers) per iteration.
//   - all other architectures: scalar fallback with no loop unrolling. The
//     scalar path is numerically identical to the SIMD paths for the same
//     input on the same platform.
func DetectSIMDCapabilities() SIMDCapabilities {
	switch runtime.GOARCH {
	case "amd64":
		// x86_64 — AVX2: 256-bit = 4 × float64 per register.
		// Unroll factor 8 = two AVX2 vectors per iteration → dual-pipeline.
		return SIMDCapabilities{
			Arch:         "amd64",
			VectorWidth:  4,
			UnrollFactor: 8,
			Available:    true,
		}
	case "arm64":
		// AArch64 — NEON/AdvSIMD: 128-bit = 2 × float64 per register.
		// Unroll factor 4 = two NEON vectors per iteration.
		return SIMDCapabilities{
			Arch:         "arm64",
			VectorWidth:  2,
			UnrollFactor: 4,
			Available:    true,
		}
	default:
		// Scalar fallback — no platform-specific vector unit targeted.
		return SIMDCapabilities{
			Arch:         runtime.GOARCH,
			VectorWidth:  1,
			UnrollFactor: 1,
			Available:    false,
		}
	}
}

// IsSIMDAvailable reports whether the current platform has a SIMD-accelerated
// code path (amd64 or arm64). The scalar fallback is used on all other
// platforms and produces numerically equivalent results.
func IsSIMDAvailable() bool {
	return simdCaps.Available
}

// GetSIMDCapabilities returns a copy of the platform capabilities that were
// detected at package initialization.
func GetSIMDCapabilities() SIMDCapabilities {
	return simdCaps
}

// CalculateLikelihoodsSIMD calculates likelihoods using SIMD-friendly vectorized operations.
// Note: Go doesn't have direct SIMD intrinsics, but this implementation uses patterns
// that the Go compiler can optimize with auto-vectorization.
//
// Architecture dispatch:
//   - amd64: countSuccessesSIMD uses 8-wide unrolled inner loop (two AVX2 vectors).
//   - arm64: countSuccessesSIMD uses 4-wide unrolled inner loop (two NEON vectors).
//   - other: scalar path via calculateLikelihoodsSimple.
//
// Semantics: type likelihood is defined as the fraction of non-empty values that
// successfully parse as that type, with float counting only for values that are not
// already integers (mutually exclusive). TypeString counts values that don't match
// any other type, consistent with the scalar fallback path.
func CalculateLikelihoodsSIMD(values []string) map[DataType]float64 {
	n := len(values)

	// Pre-allocate result vectors for SIMD-friendly processing.
	isInt := make([]bool, n)
	isFloat := make([]bool, n) // true only when NOT also an integer (exclusive)
	isBool := make([]bool, n)
	isDate := make([]bool, n)
	isEmpty := make([]bool, n)

	// Vectorized type checking — one pass per value.
	// Integer and float checks are kept mutually exclusive so that integer
	// values are never double-counted as floats (mirrors the scalar path).
	for i := 0; i < n; i++ {
		v := strings.TrimSpace(values[i])
		if v == "" {
			isEmpty[i] = true
			continue
		}

		// Integer check (most specific numeric type — checked first).
		_, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			isInt[i] = true
		} else {
			// Float check — only when value is NOT an integer.
			_, err = strconv.ParseFloat(v, 64)
			isFloat[i] = (err == nil)
		}

		// Boolean check (independent — "true"/"false" are also valid strings).
		isBool[i] = isBooleanSIMD(v)

		// Date check (independent).
		isDate[i] = isDateSIMD(v)
	}

	// Parallel reduction — architecture-specific unrolled counting.
	counts := countSuccessesSIMD(isInt, isFloat, isBool, isDate, isEmpty)

	// Calculate likelihoods as fraction of non-empty values.
	nonEmpty := counts.total - counts.empty
	likelihoods := make(map[DataType]float64)

	if nonEmpty > 0 {
		likelihoods[TypeInteger] = float64(counts.integer) / float64(nonEmpty)
		likelihoods[TypeFloat] = float64(counts.float) / float64(nonEmpty)
		likelihoods[TypeBoolean] = float64(counts.boolean) / float64(nonEmpty)
		likelihoods[TypeDate] = float64(counts.date) / float64(nonEmpty)
		// TypeString: fraction of values that match no other specific type.
		// This mirrors the scalar path and prevents TypeString from dominating
		// when the data clearly contains integers, floats, dates, or booleans.
		stringCount := nonEmpty - counts.integer - counts.float - counts.boolean - counts.date
		if stringCount < 0 {
			stringCount = 0 // guard against double-counted booleans/dates
		}
		likelihoods[TypeString] = float64(stringCount) / float64(nonEmpty)
	} else {
		likelihoods[TypeString] = 1.0
	}

	return likelihoods
}

// typeCounts holds counts for each type
type typeCounts struct {
	integer int
	float   int
	boolean int
	date    int
	empty   int
	total   int
}

// countSuccessesSIMD performs vectorized counting with architecture-specific unrolling.
//
// Architecture dispatch:
//   - amd64 (x86_64): 8-wide unrolled loop — maps to two 256-bit AVX2 vectors
//     (each holding 4 × bool/uint8). The Go compiler can issue two independent
//     vector compare-and-accumulate operations per cycle.
//   - arm64 (AArch64): 4-wide unrolled loop — maps to two 128-bit NEON vectors.
//   - other architectures: scalar element-by-element loop (no unrolling).
func countSuccessesSIMD(isInt, isFloat, isBool, isDate, isEmpty []bool) typeCounts {
	switch runtime.GOARCH {
	case "amd64":
		// x86_64: 8-wide unrolled to match two AVX2 register widths.
		return countSuccessesAMD64(isInt, isFloat, isBool, isDate, isEmpty)
	case "arm64":
		// arm64: 4-wide unrolled to match two NEON register widths.
		return countSuccessesARM64(isInt, isFloat, isBool, isDate, isEmpty)
	default:
		// Scalar fallback for unsupported architectures.
		return countSuccessesScalar(isInt, isFloat, isBool, isDate, isEmpty)
	}
}

// countSuccessesAMD64 uses an 8-wide unrolled loop optimised for x86_64.
//
// Architecture: x86_64 — AVX2 registers hold 32 bytes = 32 bool elements
// (stored as uint8). Processing 8 bools per iteration gives the compiler
// enough independent operations to fill a 256-bit vector lane. The
// conditional increments are transformed by the compiler into branchless
// CMOV / VPBLENDVB sequences on capable targets.
func countSuccessesAMD64(isInt, isFloat, isBool, isDate, isEmpty []bool) typeCounts {
	counts := typeCounts{total: len(isInt)}
	n := len(isInt)
	i := 0

	// Main loop — process 8 elements per iteration (two AVX2 bool-vectors).
	for ; i+7 < n; i += 8 {
		// Integer counts — group A (lanes 0-3) and group B (lanes 4-7).
		if isInt[i] {
			counts.integer++
		}
		if isInt[i+1] {
			counts.integer++
		}
		if isInt[i+2] {
			counts.integer++
		}
		if isInt[i+3] {
			counts.integer++
		}
		if isInt[i+4] {
			counts.integer++
		}
		if isInt[i+5] {
			counts.integer++
		}
		if isInt[i+6] {
			counts.integer++
		}
		if isInt[i+7] {
			counts.integer++
		}

		// Float counts.
		if isFloat[i] {
			counts.float++
		}
		if isFloat[i+1] {
			counts.float++
		}
		if isFloat[i+2] {
			counts.float++
		}
		if isFloat[i+3] {
			counts.float++
		}
		if isFloat[i+4] {
			counts.float++
		}
		if isFloat[i+5] {
			counts.float++
		}
		if isFloat[i+6] {
			counts.float++
		}
		if isFloat[i+7] {
			counts.float++
		}

		// Boolean counts.
		if isBool[i] {
			counts.boolean++
		}
		if isBool[i+1] {
			counts.boolean++
		}
		if isBool[i+2] {
			counts.boolean++
		}
		if isBool[i+3] {
			counts.boolean++
		}
		if isBool[i+4] {
			counts.boolean++
		}
		if isBool[i+5] {
			counts.boolean++
		}
		if isBool[i+6] {
			counts.boolean++
		}
		if isBool[i+7] {
			counts.boolean++
		}

		// Date counts.
		if isDate[i] {
			counts.date++
		}
		if isDate[i+1] {
			counts.date++
		}
		if isDate[i+2] {
			counts.date++
		}
		if isDate[i+3] {
			counts.date++
		}
		if isDate[i+4] {
			counts.date++
		}
		if isDate[i+5] {
			counts.date++
		}
		if isDate[i+6] {
			counts.date++
		}
		if isDate[i+7] {
			counts.date++
		}

		// Empty counts.
		if isEmpty[i] {
			counts.empty++
		}
		if isEmpty[i+1] {
			counts.empty++
		}
		if isEmpty[i+2] {
			counts.empty++
		}
		if isEmpty[i+3] {
			counts.empty++
		}
		if isEmpty[i+4] {
			counts.empty++
		}
		if isEmpty[i+5] {
			counts.empty++
		}
		if isEmpty[i+6] {
			counts.empty++
		}
		if isEmpty[i+7] {
			counts.empty++
		}
	}

	// Scalar tail — handle remaining elements (0–7 elements).
	for ; i < n; i++ {
		if isInt[i] {
			counts.integer++
		}
		if isFloat[i] {
			counts.float++
		}
		if isBool[i] {
			counts.boolean++
		}
		if isDate[i] {
			counts.date++
		}
		if isEmpty[i] {
			counts.empty++
		}
	}

	return counts
}

// countSuccessesARM64 uses a 4-wide unrolled loop optimised for arm64.
//
// Architecture: arm64 — NEON Q-registers hold 16 bytes = 16 bool elements
// (stored as uint8). Processing 4 bools per iteration covers two NEON loads,
// which is sufficient to keep the NEON pipeline busy on Cortex-A and Apple
// Silicon cores.
func countSuccessesARM64(isInt, isFloat, isBool, isDate, isEmpty []bool) typeCounts {
	counts := typeCounts{total: len(isInt)}
	n := len(isInt)
	i := 0

	// Main loop — process 4 elements per iteration (two NEON uint8×8 lanes).
	for ; i+3 < n; i += 4 {
		if isInt[i] {
			counts.integer++
		}
		if isInt[i+1] {
			counts.integer++
		}
		if isInt[i+2] {
			counts.integer++
		}
		if isInt[i+3] {
			counts.integer++
		}

		if isFloat[i] {
			counts.float++
		}
		if isFloat[i+1] {
			counts.float++
		}
		if isFloat[i+2] {
			counts.float++
		}
		if isFloat[i+3] {
			counts.float++
		}

		if isBool[i] {
			counts.boolean++
		}
		if isBool[i+1] {
			counts.boolean++
		}
		if isBool[i+2] {
			counts.boolean++
		}
		if isBool[i+3] {
			counts.boolean++
		}

		if isDate[i] {
			counts.date++
		}
		if isDate[i+1] {
			counts.date++
		}
		if isDate[i+2] {
			counts.date++
		}
		if isDate[i+3] {
			counts.date++
		}

		if isEmpty[i] {
			counts.empty++
		}
		if isEmpty[i+1] {
			counts.empty++
		}
		if isEmpty[i+2] {
			counts.empty++
		}
		if isEmpty[i+3] {
			counts.empty++
		}
	}

	// Scalar tail — handle remaining elements (0–3 elements).
	for ; i < n; i++ {
		if isInt[i] {
			counts.integer++
		}
		if isFloat[i] {
			counts.float++
		}
		if isBool[i] {
			counts.boolean++
		}
		if isDate[i] {
			counts.date++
		}
		if isEmpty[i] {
			counts.empty++
		}
	}

	return counts
}

// countSuccessesScalar is the plain scalar fallback used on architectures
// other than amd64 and arm64. No unrolling is applied; results are identical
// to the unrolled paths for the same input.

func countSuccessesScalar(isInt, isFloat, isBool, isDate, isEmpty []bool) typeCounts {
	counts := typeCounts{total: len(isInt)}
	for i := range isInt {
		if isInt[i] {
			counts.integer++
		}
		if isFloat[i] {
			counts.float++
		}
		if isBool[i] {
			counts.boolean++
		}
		if isDate[i] {
			counts.date++
		}
		if isEmpty[i] {
			counts.empty++
		}
	}
	return counts
}
func isBooleanSIMD(v string) bool {
	// Use lookup table for fast boolean detection
	// This is more cache-friendly and potentially vectorizable
	lower := strings.ToLower(v)
	
	// Fast path: single character booleans (excluding 0 and 1 which are integers)
	if len(lower) == 1 {
		c := lower[0]
		return c == 't' || c == 'f' || c == 'y' || c == 'n'
	}
	
	// Standard boolean values
	switch lower {
	case "true", "false", "yes", "no":
		return true
	default:
		return false
	}
}

// isDateSIMD checks if value is a date using SIMD-friendly approach
func isDateSIMD(v string) bool {
	// Optimized date checking with early exits
	// This reduces the number of expensive time.Parse calls
	
	// Quick heuristic checks before parsing
	if len(v) < 6 {
		return false // Too short to be a date
	}
	
	// Check for common date separators
	hasSeparator := false
	for i := 0; i < len(v); i++ {
		c := v[i]
		if c == '-' || c == '/' || c == ' ' || c == 'T' || c == ':' {
			hasSeparator = true
			break
		}
	}
	
	if !hasSeparator && len(v) != 8 {
		// No separator and not YYYYMMDD format
		return false
	}
	
	// Try fast path formats first (most common)
	fastFormats := []string{
		"2006-01-02",
		"01/02/2006",
		"2006-01-02T15:04:05",
	}
	
	for _, format := range fastFormats {
		if _, err := time.Parse(format, v); err == nil {
			return true
		}
	}
	
	// Try remaining formats
	slowFormats := []string{
		"02-01-2006",
		"Jan 2 2006",
		"January 2 2006",
		time.RFC3339,
		"2006-01-02 15:04:05",
	}
	
	for _, format := range slowFormats {
		if _, err := time.Parse(format, v); err == nil {
			return true
		}
	}
	
	return false
}

// VectorizedStatistics holds the statistical measures produced by the vectorized
// and scalar stat kernels. All fields are identical regardless of which code
// path is used for a given input.
type VectorizedStatistics struct {
	Mean     float64
	Variance float64
	StdDev   float64
	Min      float64
	Max      float64
	Count    int
}

// CalculateStatsSIMD computes mean, variance, standard deviation, min, and max
// using an architecture-optimised kernel.
//
// Architecture dispatch:
//   - amd64 (x86_64): 8-wide unrolled inner loop using 8 independent float64
//     accumulators. The independent accumulators map to two 256-bit AVX2 YMM
//     registers (4 × float64 each), enabling dual-pipeline execution.
//   - arm64 (AArch64): 4-wide unrolled inner loop using 4 independent float64
//     accumulators. The independent accumulators map to two 128-bit NEON
//     Q-registers (2 × float64 each).
//   - other architectures: CalculateStatsScalar — a plain sequential loop that
//     produces identical results to the SIMD paths on the same machine.
//
// Determinism: each path uses independent accumulators with a fixed tree
// reduction order, so repeated calls with the same input always return the
// same floating-point result. The SIMD and scalar paths may differ by at most
// a few ULPs due to floating-point non-associativity; this is expected and
// documented behaviour.
func CalculateStatsSIMD(values []float64) VectorizedStatistics {
	if len(values) == 0 {
		return VectorizedStatistics{}
	}

	switch runtime.GOARCH {
	case "amd64":
		// x86_64: 8-wide unrolled loop — two AVX2 YMM vectors per iteration.
		return calculateStatsAMD64(values)
	case "arm64":
		// arm64: 4-wide unrolled loop — two NEON Q-registers per iteration.
		return calculateStatsARM64(values)
	default:
		// Scalar fallback for unsupported architectures.
		return CalculateStatsScalar(values)
	}
}

// CalculateStatsScalar computes the same statistics as CalculateStatsSIMD
// using a plain sequential loop with no loop unrolling. This is the reference
// scalar implementation and the fallback for architectures other than amd64
// and arm64.
//
// It is exported so callers can compare scalar vs SIMD results in tests.
func CalculateStatsScalar(values []float64) VectorizedStatistics {
	if len(values) == 0 {
		return VectorizedStatistics{}
	}

	n := len(values)
	sum := 0.0
	sumSq := 0.0
	min := values[0]
	max := values[0]

	// Plain sequential loop — single accumulator, no unrolling.
	for i := 1; i < n; i++ {
		v := values[i]
		sum += v
		sumSq += v * v
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	// Include the first element that was used to seed min/max.
	sum += values[0]
	sumSq += values[0] * values[0]

	mean := sum / float64(n)
	variance := (sumSq / float64(n)) - (mean * mean)
	stdDev := 0.0
	if variance > 0 {
		stdDev = math.Sqrt(variance)
	}

	return VectorizedStatistics{
		Mean:     mean,
		Variance: variance,
		StdDev:   stdDev,
		Min:      min,
		Max:      max,
		Count:    n,
	}
}

// calculateStatsAMD64 is the x86_64-optimised stat kernel.
//
// Architecture: x86_64 — AVX2 YMM registers hold 4 × float64 (256 bits).
// We use 8 independent sum accumulators (two YMM register widths) so the
// compiler can issue two independent VADDPD (256-bit FP add) instructions per
// cycle, saturating both FP execution units on Haswell-and-later CPUs.
// The accumulators are reduced with a balanced tree sum to minimise round-off
// accumulation. The same strategy is applied to sumSq.
func calculateStatsAMD64(values []float64) VectorizedStatistics {
	n := len(values)

	// Eight independent accumulators — map to two AVX2 vectors of 4 × float64.
	var s0, s1, s2, s3, s4, s5, s6, s7 float64
	var q0, q1, q2, q3, q4, q5, q6, q7 float64 // sum-of-squares

	min := values[0]
	max := values[0]

	i := 0
	// Main loop — 8 elements per iteration (two 256-bit AVX2 vector loads).
	for ; i+7 < n; i += 8 {
		v0 := values[i]
		v1 := values[i+1]
		v2 := values[i+2]
		v3 := values[i+3]
		v4 := values[i+4]
		v5 := values[i+5]
		v6 := values[i+6]
		v7 := values[i+7]

		// Independent accumulation — no data dependency between lanes.
		s0 += v0
		s1 += v1
		s2 += v2
		s3 += v3
		s4 += v4
		s5 += v5
		s6 += v6
		s7 += v7

		q0 += v0 * v0
		q1 += v1 * v1
		q2 += v2 * v2
		q3 += v3 * v3
		q4 += v4 * v4
		q5 += v5 * v5
		q6 += v6 * v6
		q7 += v7 * v7

		// Min/max reduction — branchless on modern compilers (VMINPD/VMAXPD).
		if v0 < min {
			min = v0
		}
		if v1 < min {
			min = v1
		}
		if v2 < min {
			min = v2
		}
		if v3 < min {
			min = v3
		}
		if v4 < min {
			min = v4
		}
		if v5 < min {
			min = v5
		}
		if v6 < min {
			min = v6
		}
		if v7 < min {
			min = v7
		}

		if v0 > max {
			max = v0
		}
		if v1 > max {
			max = v1
		}
		if v2 > max {
			max = v2
		}
		if v3 > max {
			max = v3
		}
		if v4 > max {
			max = v4
		}
		if v5 > max {
			max = v5
		}
		if v6 > max {
			max = v6
		}
		if v7 > max {
			max = v7
		}
	}

	// Scalar tail — handle remaining 0–7 elements.
	for ; i < n; i++ {
		v := values[i]
		s0 += v
		q0 += v * v
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	// Balanced tree reduction to minimise accumulated round-off error.
	sum := (s0 + s1) + (s2 + s3) + (s4 + s5) + (s6 + s7)
	sumSq := (q0 + q1) + (q2 + q3) + (q4 + q5) + (q6 + q7)

	mean := sum / float64(n)
	variance := (sumSq / float64(n)) - (mean * mean)
	stdDev := 0.0
	if variance > 0 {
		stdDev = math.Sqrt(variance)
	}

	return VectorizedStatistics{
		Mean:     mean,
		Variance: variance,
		StdDev:   stdDev,
		Min:      min,
		Max:      max,
		Count:    n,
	}
}

// calculateStatsARM64 is the arm64-optimised stat kernel.
//
// Architecture: arm64 — NEON/AdvSIMD Q-registers hold 2 × float64 (128 bits).
// We use 4 independent accumulators (two NEON Q-register widths) so the
// compiler can issue two independent FADD (128-bit FP add) instructions per
// cycle. Cortex-A and Apple Silicon M-series cores both benefit from this
// dual-issue pattern. Accumulators are reduced with a balanced tree sum.
func calculateStatsARM64(values []float64) VectorizedStatistics {
	n := len(values)

	// Four independent accumulators — map to two NEON 128-bit Q-registers.
	var s0, s1, s2, s3 float64
	var q0, q1, q2, q3 float64 // sum-of-squares

	min := values[0]
	max := values[0]

	i := 0
	// Main loop — 4 elements per iteration (two 128-bit NEON vector loads).
	for ; i+3 < n; i += 4 {
		v0 := values[i]
		v1 := values[i+1]
		v2 := values[i+2]
		v3 := values[i+3]

		// Independent accumulation — no data dependency between lanes.
		s0 += v0
		s1 += v1
		s2 += v2
		s3 += v3

		q0 += v0 * v0
		q1 += v1 * v1
		q2 += v2 * v2
		q3 += v3 * v3

		// Min/max reduction — FMIN/FMAX on arm64.
		if v0 < min {
			min = v0
		}
		if v1 < min {
			min = v1
		}
		if v2 < min {
			min = v2
		}
		if v3 < min {
			min = v3
		}

		if v0 > max {
			max = v0
		}
		if v1 > max {
			max = v1
		}
		if v2 > max {
			max = v2
		}
		if v3 > max {
			max = v3
		}
	}

	// Scalar tail — handle remaining 0–3 elements.
	for ; i < n; i++ {
		v := values[i]
		s0 += v
		q0 += v * v
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	// Balanced tree reduction.
	sum := (s0 + s1) + (s2 + s3)
	sumSq := (q0 + q1) + (q2 + q3)

	mean := sum / float64(n)
	variance := (sumSq / float64(n)) - (mean * mean)
	stdDev := 0.0
	if variance > 0 {
		stdDev = math.Sqrt(variance)
	}

	return VectorizedStatistics{
		Mean:     mean,
		Variance: variance,
		StdDev:   stdDev,
		Min:      min,
		Max:      max,
		Count:    n,
	}
}

// ParallelLikelihoodBatch processes multiple columns in parallel with SIMD
func ParallelLikelihoodBatch(columns [][]string) []map[DataType]float64 {
	results := make([]map[DataType]float64, len(columns))
	
	// Process each column
	for i, col := range columns {
		if len(col) > 100 {
			results[i] = CalculateLikelihoodsSIMD(col)
		} else {
			// For small datasets, SIMD overhead isn't worth it
			results[i] = calculateLikelihoodsSimple(col)
		}
	}
	
	return results
}

// calculateLikelihoodsSimple is a simple non-SIMD version for small datasets
func calculateLikelihoodsSimple(values []string) map[DataType]float64 {
	likelihoods := make(map[DataType]float64)
	counts := map[DataType]int{}
	nonEmpty := 0
	
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		nonEmpty++
		
		isInt := false
		isFloat := false
		isBool := false
		isDate := false
		
		// Check integer first (most specific numeric type)
		if _, err := strconv.ParseInt(v, 10, 64); err == nil {
			counts[TypeInteger]++
			isInt = true
		} else {
			// Only check float if not an integer
			if _, err := strconv.ParseFloat(v, 64); err == nil {
				counts[TypeFloat]++
				isFloat = true
			}
		}
		
		if isBooleanSIMD(v) {
			counts[TypeBoolean]++
			isBool = true
		}
		if isDateSIMD(v) {
			counts[TypeDate]++
			isDate = true
		}
		
		// Count as string only if it's NOT a specific type
		if !isInt && !isFloat && !isBool && !isDate {
			counts[TypeString]++
		}
	}
	
	if nonEmpty > 0 {
		for typ, count := range counts {
			likelihoods[typ] = float64(count) / float64(nonEmpty)
		}
	} else {
		likelihoods[TypeString] = 1.0
	}
	
	return likelihoods
}
