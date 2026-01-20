package engine

import (
	"math"
	"strconv"
	"strings"
	"time"
)

// CalculateLikelihoodsSIMD calculates likelihoods using SIMD-friendly vectorized operations
// Note: Go doesn't have direct SIMD intrinsics, but this implementation uses patterns
// that the Go compiler can optimize with auto-vectorization
func CalculateLikelihoodsSIMD(values []string) map[DataType]float64 {
	n := len(values)
	
	// Pre-allocate result vectors for SIMD-friendly processing
	isInt := make([]bool, n)
	isFloat := make([]bool, n)
	isBool := make([]bool, n)
	isDate := make([]bool, n)
	isEmpty := make([]bool, n)
	
	// Vectorized type checking - batch operations
	// This allows the compiler to potentially vectorize the loops
	for i := 0; i < n; i++ {
		v := strings.TrimSpace(values[i])
		if v == "" {
			isEmpty[i] = true
			continue
		}
		
		// Integer check
		_, err := strconv.ParseInt(v, 10, 64)
		isInt[i] = (err == nil)
		
		// Float check
		_, err = strconv.ParseFloat(v, 64)
		isFloat[i] = (err == nil)
		
		// Boolean check
		isBool[i] = isBooleanSIMD(v)
		
		// Date check
		isDate[i] = isDateSIMD(v)
	}
	
	// Parallel reduction - count successes
	// Using SIMD-friendly reduction pattern
	counts := countSuccessesSIMD(isInt, isFloat, isBool, isDate, isEmpty)
	
	// Calculate likelihoods
	nonEmpty := counts.total - counts.empty
	likelihoods := make(map[DataType]float64)
	
	if nonEmpty > 0 {
		likelihoods[TypeInteger] = float64(counts.integer) / float64(nonEmpty)
		likelihoods[TypeFloat] = float64(counts.float) / float64(nonEmpty)
		likelihoods[TypeBoolean] = float64(counts.boolean) / float64(nonEmpty)
		likelihoods[TypeDate] = float64(counts.date) / float64(nonEmpty)
		likelihoods[TypeString] = 1.0 // All non-empty values are valid strings
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

// countSuccessesSIMD performs vectorized counting
func countSuccessesSIMD(isInt, isFloat, isBool, isDate, isEmpty []bool) typeCounts {
	counts := typeCounts{
		total: len(isInt),
	}
	
	// Unrolled loop for better vectorization
	// Process 4 elements at a time when possible
	i := 0
	n := len(isInt)
	
	// Main loop - process 4 at a time
	for ; i+3 < n; i += 4 {
		// Batch process 4 elements
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
	
	// Handle remaining elements
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

// isBooleanSIMD checks if value is boolean using SIMD-friendly approach
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

// VectorizedStatistics calculates statistical measures using vectorized operations
type VectorizedStatistics struct {
	Mean     float64
	Variance float64
	StdDev   float64
	Min      float64
	Max      float64
	Count    int
}

// CalculateStatsSIMD computes statistics using SIMD-friendly vectorization
func CalculateStatsSIMD(values []float64) VectorizedStatistics {
	if len(values) == 0 {
		return VectorizedStatistics{}
	}
	
	n := len(values)
	
	// Initialize aggregators
	sum := 0.0
	sumSq := 0.0
	min := values[0]
	max := values[0]
	
	// Vectorized accumulation - process 4 at a time
	i := 0
	for ; i+3 < n; i += 4 {
		v0 := values[i]
		v1 := values[i+1]
		v2 := values[i+2]
		v3 := values[i+3]
		
		sum += v0 + v1 + v2 + v3
		sumSq += v0*v0 + v1*v1 + v2*v2 + v3*v3
		
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
	
	// Handle remaining elements
	for ; i < n; i++ {
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
	
	// Calculate statistics
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
