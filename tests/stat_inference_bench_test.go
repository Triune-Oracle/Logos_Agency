package tests

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/Triune-Oracle/Logos_Agency/engine"
)

// Benchmark data generators
func generateIntegerData(n int) []string {
	data := make([]string, n)
	for i := 0; i < n; i++ {
		data[i] = strconv.Itoa(i)
	}
	return data
}

func generateFloatData(n int) []string {
	data := make([]string, n)
	for i := 0; i < n; i++ {
		data[i] = fmt.Sprintf("%.2f", float64(i)*1.5)
	}
	return data
}

func generateStringData(n int) []string {
	data := make([]string, n)
	for i := 0; i < n; i++ {
		data[i] = fmt.Sprintf("string_%d", i)
	}
	return data
}

func generateDateData(n int) []string {
	data := make([]string, n)
	for i := 0; i < n; i++ {
		data[i] = fmt.Sprintf("2024-01-%02d", (i%28)+1)
	}
	return data
}

func generateMixedData(n int) []string {
	data := make([]string, n)
	for i := 0; i < n; i++ {
		switch i % 4 {
		case 0:
			data[i] = strconv.Itoa(i)
		case 1:
			data[i] = fmt.Sprintf("%.2f", float64(i)*1.5)
		case 2:
			data[i] = fmt.Sprintf("string_%d", i)
		case 3:
			data[i] = "2024-01-01"
		}
	}
	return data
}

// Benchmarks for different data sizes

func BenchmarkBayesianInference_100(b *testing.B) {
	data := generateIntegerData(100)
	eng := engine.NewBayesianInferenceEngine(engine.DefaultConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eng.InferType(data)
	}
}

func BenchmarkBayesianInference_1000(b *testing.B) {
	data := generateIntegerData(1000)
	eng := engine.NewBayesianInferenceEngine(engine.DefaultConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eng.InferType(data)
	}
}

func BenchmarkBayesianInference_10000(b *testing.B) {
	data := generateIntegerData(10000)
	eng := engine.NewBayesianInferenceEngine(engine.DefaultConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eng.InferType(data)
	}
}

func BenchmarkBayesianInference_100000(b *testing.B) {
	data := generateIntegerData(100000)
	eng := engine.NewBayesianInferenceEngine(engine.DefaultConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eng.InferType(data)
	}
}

// Benchmarks for different data types

func BenchmarkBayesianInference_Integer(b *testing.B) {
	data := generateIntegerData(1000)
	eng := engine.NewBayesianInferenceEngine(engine.DefaultConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eng.InferType(data)
	}
}

func BenchmarkBayesianInference_Float(b *testing.B) {
	data := generateFloatData(1000)
	eng := engine.NewBayesianInferenceEngine(engine.DefaultConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eng.InferType(data)
	}
}

func BenchmarkBayesianInference_String(b *testing.B) {
	data := generateStringData(1000)
	eng := engine.NewBayesianInferenceEngine(engine.DefaultConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eng.InferType(data)
	}
}

func BenchmarkBayesianInference_Date(b *testing.B) {
	data := generateDateData(1000)
	eng := engine.NewBayesianInferenceEngine(engine.DefaultConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eng.InferType(data)
	}
}

func BenchmarkBayesianInference_Mixed(b *testing.B) {
	data := generateMixedData(1000)
	eng := engine.NewBayesianInferenceEngine(engine.DefaultConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eng.InferType(data)
	}
}

// SIMD vs non-SIMD benchmarks

func BenchmarkSIMD_Enabled_1000(b *testing.B) {
	data := generateIntegerData(1000)
	config := engine.DefaultConfig()
	config.EnableSIMD = true
	eng := engine.NewBayesianInferenceEngine(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eng.InferType(data)
	}
}

func BenchmarkSIMD_Disabled_1000(b *testing.B) {
	data := generateIntegerData(1000)
	config := engine.DefaultConfig()
	config.EnableSIMD = false
	eng := engine.NewBayesianInferenceEngine(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eng.InferType(data)
	}
}

func BenchmarkSIMD_Enabled_10000(b *testing.B) {
	data := generateIntegerData(10000)
	config := engine.DefaultConfig()
	config.EnableSIMD = true
	eng := engine.NewBayesianInferenceEngine(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eng.InferType(data)
	}
}

func BenchmarkSIMD_Disabled_10000(b *testing.B) {
	data := generateIntegerData(10000)
	config := engine.DefaultConfig()
	config.EnableSIMD = false
	eng := engine.NewBayesianInferenceEngine(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eng.InferType(data)
	}
}

// SIMD fastpath direct benchmarks

func BenchmarkCalculateLikelihoodsSIMD_100(b *testing.B) {
	data := generateIntegerData(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.CalculateLikelihoodsSIMD(data)
	}
}

func BenchmarkCalculateLikelihoodsSIMD_1000(b *testing.B) {
	data := generateIntegerData(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.CalculateLikelihoodsSIMD(data)
	}
}

func BenchmarkCalculateLikelihoodsSIMD_10000(b *testing.B) {
	data := generateIntegerData(10000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.CalculateLikelihoodsSIMD(data)
	}
}

// Locale-specific benchmarks

func BenchmarkBayesianInference_LocaleUS(b *testing.B) {
	data := generateIntegerData(1000)
	config := engine.DefaultConfig()
	config.Locale = engine.LocaleUS
	eng := engine.NewBayesianInferenceEngine(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eng.InferType(data)
	}
}

func BenchmarkBayesianInference_LocaleEU(b *testing.B) {
	data := generateIntegerData(1000)
	config := engine.DefaultConfig()
	config.Locale = engine.LocaleEU
	eng := engine.NewBayesianInferenceEngine(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eng.InferType(data)
	}
}

func BenchmarkBayesianInference_LocaleASIA(b *testing.B) {
	data := generateIntegerData(1000)
	config := engine.DefaultConfig()
	config.Locale = engine.LocaleASIA
	eng := engine.NewBayesianInferenceEngine(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eng.InferType(data)
	}
}

// Sampling benchmarks

func BenchmarkBayesianInference_Sampling_Large(b *testing.B) {
	data := generateIntegerData(100000)
	config := engine.DefaultConfig()
	config.SampleSize = 1000
	eng := engine.NewBayesianInferenceEngine(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eng.InferType(data)
	}
}

func BenchmarkBayesianInference_NoSampling_Small(b *testing.B) {
	data := generateIntegerData(500)
	config := engine.DefaultConfig()
	config.SampleSize = 1000 // Larger than data
	eng := engine.NewBayesianInferenceEngine(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eng.InferType(data)
	}
}

// Parallel batch processing benchmark

func BenchmarkParallelLikelihoodBatch_10Columns(b *testing.B) {
	columns := make([][]string, 10)
	for i := 0; i < 10; i++ {
		columns[i] = generateIntegerData(1000)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.ParallelLikelihoodBatch(columns)
	}
}

func BenchmarkParallelLikelihoodBatch_100Columns(b *testing.B) {
	columns := make([][]string, 100)
	for i := 0; i < 100; i++ {
		columns[i] = generateIntegerData(1000)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.ParallelLikelihoodBatch(columns)
	}
}

// Vectorized statistics benchmarks

func BenchmarkCalculateStatsSIMD_1000(b *testing.B) {
	data := make([]float64, 1000)
	for i := 0; i < 1000; i++ {
		data[i] = float64(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.CalculateStatsSIMD(data)
	}
}

func BenchmarkCalculateStatsSIMD_10000(b *testing.B) {
	data := make([]float64, 10000)
	for i := 0; i < 10000; i++ {
		data[i] = float64(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.CalculateStatsSIMD(data)
	}
}

// Scalar fallback benchmarks — paired with the SIMD benchmarks above so the
// speedup ratio can be observed directly. On amd64/arm64 the SIMD-dispatched
// path should outperform the scalar path by ≥2x on large inputs.

func BenchmarkCalculateStatsScalar_1000(b *testing.B) {
	data := make([]float64, 1000)
	for i := 0; i < 1000; i++ {
		data[i] = float64(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.CalculateStatsScalar(data)
	}
}

func BenchmarkCalculateStatsScalar_10000(b *testing.B) {
	data := make([]float64, 10000)
	for i := 0; i < 10000; i++ {
		data[i] = float64(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.CalculateStatsScalar(data)
	}
}

// Memory allocation benchmarks

func BenchmarkBayesianInference_Memory_1000(b *testing.B) {
	data := generateIntegerData(1000)
	eng := engine.NewBayesianInferenceEngine(engine.DefaultConfig())

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eng.InferType(data)
	}
}

func BenchmarkBayesianInference_Memory_10000(b *testing.B) {
	data := generateIntegerData(10000)
	eng := engine.NewBayesianInferenceEngine(engine.DefaultConfig())

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eng.InferType(data)
	}
}

// Regression test - ensure performance stays within bounds
func BenchmarkRegressionTest_1000(b *testing.B) {
	data := generateIntegerData(1000)
	eng := engine.NewBayesianInferenceEngine(engine.DefaultConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eng.InferType(data)
	}

	// Performance target: < 1ms per operation for 1000 rows
	avgNs := b.Elapsed().Nanoseconds() / int64(b.N)
	targetNs := int64(1_000_000) // 1ms in nanoseconds

	if avgNs > targetNs {
		b.Errorf("Performance regression: avg %dns exceeds target %dns", avgNs, targetNs)
	}
}

func BenchmarkRegressionTest_10000(b *testing.B) {
	data := generateIntegerData(10000)
	eng := engine.NewBayesianInferenceEngine(engine.DefaultConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eng.InferType(data)
	}

	// Performance target: < 10ms per operation for 10000 rows
	avgNs := b.Elapsed().Nanoseconds() / int64(b.N)
	targetNs := int64(10_000_000) // 10ms in nanoseconds

	if avgNs > targetNs {
		b.Errorf("Performance regression: avg %dns exceeds target %dns", avgNs, targetNs)
	}
}
