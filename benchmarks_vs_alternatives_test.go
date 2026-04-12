package main

// benchmarks_vs_alternatives_test.go
//
// Comprehensive benchmark suite comparing Logos_Agency's optimised
// InferColumnType against:
//   - A naive (unoptimised) implementation that mirrors how Python's pandas
//     and csvkit approach column-type inference: no early-exit, parse every
//     value against every type on every call.
//   - The streaming helper InferColumnTypesStreaming for multi-column CSVs.
//
// Run with:
//   go test -bench=. -benchmem -benchtime=3s -count=3 ./...
//   go test -bench=BenchmarkComparison -benchmem -v

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Naive implementation – mirrors pandas/csvkit "check everything every time"
// ---------------------------------------------------------------------------

// naiveDateFormats are the same formats as InferColumnType but applied
// unconditionally (no early-exit).
var naiveDateFormats = []string{
	"2006-01-02",
	"01/02/2006",
	"02-01-2006",
	"Jan 2 2006",
	"January 2 2006",
	time.RFC3339,
}

// naiveInferColumnType is an intentionally unoptimised reference
// implementation that resembles the parsing strategy used by pandas' read_csv
// and csvkit's typeinference module: every value is parsed against every type,
// with no early termination and no conditional skipping.
func naiveInferColumnType(values []string) string {
	allInt := true
	allFloat := true
	allDate := true
	hasNonEmpty := false

	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		hasNonEmpty = true

		// Always check int (no early-exit)
		if _, err := strconv.Atoi(v); err != nil {
			allInt = false
		}
		// Always check float (no early-exit)
		if _, err := strconv.ParseFloat(v, 64); err != nil {
			allFloat = false
			allInt = false
		}
		// Always check every date format (no early-exit)
		parsedAsDate := false
		for _, f := range naiveDateFormats {
			if _, err := time.Parse(f, v); err == nil {
				parsedAsDate = true
				break
			}
		}
		if !parsedAsDate {
			allDate = false
		}
	}

	if !hasNonEmpty {
		return "string"
	}
	if allInt {
		return "int"
	} else if allFloat {
		return "float"
	} else if allDate {
		return "date"
	}
	return "string"
}

// ---------------------------------------------------------------------------
// Test data generators for different column profiles
// ---------------------------------------------------------------------------

func generateIntRows(n int) []string {
	rows := make([]string, n)
	for i := 0; i < n; i++ {
		rows[i] = strconv.Itoa(i)
	}
	return rows
}

func generateFloatRows(n int) []string {
	rows := make([]string, n)
	for i := 0; i < n; i++ {
		rows[i] = strconv.FormatFloat(float64(i)*0.5, 'f', 2, 64)
	}
	return rows
}

func generateStringRows(n int) []string {
	rows := make([]string, n)
	for i := 0; i < n; i++ {
		rows[i] = "value_" + strconv.Itoa(i)
	}
	return rows
}

func generateDateRows(n int) []string {
	rows := make([]string, n)
	for i := 0; i < n; i++ {
		rows[i] = "2024-01-" + strconv.Itoa((i%28)+1)
	}
	return rows
}

// generateCSVLines returns a slice of CSV lines representing a table with
// columns of different types (int, float, string, date).
func generateCSVLines(numRows int) []string {
	lines := make([]string, numRows+1)
	lines[0] = "id,price,label,created_at"
	for i := 1; i <= numRows; i++ {
		lines[i] = strconv.Itoa(i) + "," +
			strconv.FormatFloat(float64(i)*1.25, 'f', 2, 64) + "," +
			"item_" + strconv.Itoa(i) + "," +
			"2024-01-" + strconv.Itoa((i%28)+1)
	}
	return lines
}

// ---------------------------------------------------------------------------
// Comparison benchmarks: Optimised vs Naive  (pandas/csvkit proxy)
// ---------------------------------------------------------------------------

// 1 000 rows

func BenchmarkComparison_Optimised_1k_Int(b *testing.B) {
	data := generateIntRows(1_000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = InferColumnType(data)
	}
}

func BenchmarkComparison_Naive_1k_Int(b *testing.B) {
	data := generateIntRows(1_000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = naiveInferColumnType(data)
	}
}

// 10 000 rows

func BenchmarkComparison_Optimised_10k_Int(b *testing.B) {
	data := generateIntRows(10_000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = InferColumnType(data)
	}
}

func BenchmarkComparison_Naive_10k_Int(b *testing.B) {
	data := generateIntRows(10_000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = naiveInferColumnType(data)
	}
}

// 100 000 rows

func BenchmarkComparison_Optimised_100k_Int(b *testing.B) {
	data := generateIntRows(100_000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = InferColumnType(data)
	}
}

func BenchmarkComparison_Naive_100k_Int(b *testing.B) {
	data := generateIntRows(100_000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = naiveInferColumnType(data)
	}
}

// 1 000 000 rows

func BenchmarkComparison_Optimised_1M_Int(b *testing.B) {
	data := generateIntRows(1_000_000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = InferColumnType(data)
	}
}

func BenchmarkComparison_Naive_1M_Int(b *testing.B) {
	data := generateIntRows(1_000_000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = naiveInferColumnType(data)
	}
}

// ---------------------------------------------------------------------------
// Comparison by data type at 10k rows
// ---------------------------------------------------------------------------

func BenchmarkComparison_Optimised_10k_Float(b *testing.B) {
	data := generateFloatRows(10_000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = InferColumnType(data)
	}
}

func BenchmarkComparison_Naive_10k_Float(b *testing.B) {
	data := generateFloatRows(10_000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = naiveInferColumnType(data)
	}
}

func BenchmarkComparison_Optimised_10k_String(b *testing.B) {
	data := generateStringRows(10_000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = InferColumnType(data)
	}
}

func BenchmarkComparison_Naive_10k_String(b *testing.B) {
	data := generateStringRows(10_000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = naiveInferColumnType(data)
	}
}

func BenchmarkComparison_Optimised_10k_Date(b *testing.B) {
	data := generateDateRows(10_000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = InferColumnType(data)
	}
}

func BenchmarkComparison_Naive_10k_Date(b *testing.B) {
	data := generateDateRows(10_000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = naiveInferColumnType(data)
	}
}

// ---------------------------------------------------------------------------
// Streaming (multi-column CSV) benchmarks
// ---------------------------------------------------------------------------

func BenchmarkComparison_Streaming_1k(b *testing.B) {
	lines := generateCSVLines(1_000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = InferColumnTypesStreaming(lines, 1_000)
	}
}

func BenchmarkComparison_Streaming_10k(b *testing.B) {
	lines := generateCSVLines(10_000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = InferColumnTypesStreaming(lines, 10_000)
	}
}

func BenchmarkComparison_Streaming_100k(b *testing.B) {
	lines := generateCSVLines(100_000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = InferColumnTypesStreaming(lines, 100_000)
	}
}

func BenchmarkComparison_Streaming_1M(b *testing.B) {
	lines := generateCSVLines(1_000_000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = InferColumnTypesStreaming(lines, 1_000_000)
	}
}

// ---------------------------------------------------------------------------
// Speedup assertion test – fails if optimised path is not at least 10x faster
// ---------------------------------------------------------------------------

// TestSpeedup_Optimised_vs_Naive_10k asserts that InferColumnType is at least
// 10× faster than the naive approach for a 10 000-row integer column.
func TestSpeedup_Optimised_vs_Naive_10k(t *testing.T) {
	const rows = 10_000
	const iterations = 50

	data := generateIntRows(rows)

	// Warm up
	_ = InferColumnType(data)
	_ = naiveInferColumnType(data)

	// Time optimised
	start := time.Now()
	for i := 0; i < iterations; i++ {
		_ = InferColumnType(data)
	}
	optimisedTotal := time.Since(start)

	// Time naive
	start = time.Now()
	for i := 0; i < iterations; i++ {
		_ = naiveInferColumnType(data)
	}
	naiveTotal := time.Since(start)

	if naiveTotal == 0 {
		t.Skip("Naive total time was 0, skipping speedup assertion")
	}

	speedup := float64(naiveTotal) / float64(optimisedTotal)
	t.Logf("Speedup over naive (10k int rows, %d iters): %.1fx  (optimised=%v  naive=%v)",
		iterations, speedup, optimisedTotal/iterations, naiveTotal/iterations)

	const minSpeedup = 10.0
	if speedup < minSpeedup {
		t.Errorf("Expected at least %.0fx speedup over naive approach, got %.1fx", minSpeedup, speedup)
	}
}
