# Statistical Type Inference Engine

## Overview

The Statistical Type Inference Engine provides advanced, Bayesian-based type detection for data columns with SIMD acceleration, locale-aware priors, and deterministic behavior. It goes beyond simple heuristic type checking by incorporating statistical methods and domain knowledge.

## Architecture

### Core Components

1. **BayesianInferenceEngine** (`engine/stat_inference.go`)
   - Main inference engine using Bayesian statistical methods
   - Combines prior probabilities with observed evidence
   - Supports multiple locales and deterministic behavior

2. **SIMD Fastpath** (`engine/simd_fastpath.go`)
   - Vectorized operations for improved performance
   - Automatic fallback for small datasets
   - Optimized likelihood calculations

### Key Features

- **Bayesian Inference**: Uses prior probabilities and evidence to compute posterior type probabilities
- **Locale-Aware Priors**: Different regional data patterns (US, EU, Asia, Global)
- **SIMD Acceleration**: Vectorized operations for large datasets
- **Deterministic Behavior**: Configurable random seed for reproducible results
- **Performance Optimized**: Targets < 1ms for 1K rows, < 10ms for 10K rows

## Algorithm

### Bayesian Type Inference

The engine uses Bayes' theorem to calculate type probabilities:

```
P(Type|Data) ∝ P(Data|Type) × P(Type)
```

Where:
- `P(Type|Data)` is the posterior probability (what we want)
- `P(Data|Type)` is the likelihood (how well data fits the type)
- `P(Type)` is the prior probability (based on locale and domain knowledge)

### Likelihood Calculation

For each value in the dataset, we calculate the likelihood of it belonging to each type:

1. **Integer**: Can parse as `int64`
2. **Float**: Can parse as `float64`
3. **Boolean**: Matches boolean patterns (true/false, yes/no, 1/0)
4. **Date**: Matches common date formats
5. **String**: All values are valid strings (baseline)

### Prior Probabilities

Prior probabilities are locale-specific:

| Locale | Integer | Float | String | Date | Boolean |
|--------|---------|-------|--------|------|---------|
| US     | 0.25    | 0.20  | 0.40   | 0.10 | 0.05    |
| EU     | 0.25    | 0.20  | 0.40   | 0.10 | 0.05    |
| Asia   | 0.30    | 0.15  | 0.40   | 0.10 | 0.05    |
| Global | 0.25    | 0.20  | 0.40   | 0.10 | 0.05    |

These priors reflect typical data distributions in different regions.

## Usage

### Basic Usage

```go
import "github.com/Triune-Oracle/Logos_Agency/engine"

// Create engine with default configuration
engine := engine.NewBayesianInferenceEngine(engine.DefaultConfig())

// Infer type from values
values := []string{"1", "2", "3", "4", "5"}
result := engine.InferType(values)

fmt.Printf("Type: %v, Confidence: %.2f\n", result.InferredType, result.Confidence)
```

### Advanced Configuration

```go
// Custom configuration
config := &engine.InferenceConfig{
    Locale:           engine.LocaleUS,
    SampleSize:       1000,
    ConfidenceThresh: 0.80,
    RandomSeed:       42,
    EnableSIMD:       true,
}

engine := engine.NewBayesianInferenceEngine(config)
result := engine.InferType(largeDataset)
```

### Locale-Specific Inference

```go
engine := engine.NewBayesianInferenceEngine(engine.DefaultConfig())

// Change locale at runtime
engine.SetLocale(engine.LocaleASIA)

result := engine.InferType(values)
```

### Deterministic Behavior

```go
engine := engine.NewBayesianInferenceEngine(engine.DefaultConfig())

// Set seed for reproducibility
engine.SetSeed(12345)

result := engine.InferType(values)
// Same seed will always produce same result
```

### Simple API

```go
engine := engine.NewBayesianInferenceEngine(engine.DefaultConfig())

// Returns type as string ("int", "float", "string", "date", "boolean")
typeStr := engine.InferTypeSimple(values)
```

## SIMD Acceleration

### When SIMD is Used

SIMD acceleration is automatically enabled for datasets with:
- More than 100 values
- `EnableSIMD` config option set to `true` (default)

For smaller datasets, the overhead of SIMD isn't worth it, so the engine falls back to simple scalar operations.

### SIMD Implementation

The SIMD fastpath uses vectorization-friendly patterns:

1. **Batch Type Checking**: Process 4 values at a time
2. **Vectorized Counting**: Unrolled loops for better compiler optimization
3. **Early Exit Optimization**: Skip expensive date parsing when possible

### Performance Impact

SIMD typically provides:
- **1.5-2x speedup** for large datasets (10K+ rows)
- **Minimal overhead** for small datasets (auto-disabled)
- **Better cache utilization** through vectorized memory access

## Performance Characteristics

### Performance Targets

| Dataset Size | Target Time | SIMD Speedup |
|--------------|-------------|--------------|
| 100 rows     | < 0.1ms     | 1.0x         |
| 1K rows      | < 1ms       | 1.2x         |
| 10K rows     | < 10ms      | 1.5x         |
| 100K rows    | < 100ms     | 2.0x         |

### Memory Usage

The engine uses efficient memory allocation:

- **Sampling**: Reduces memory for large datasets
- **Pre-allocation**: Minimizes allocations in hot paths
- **Zero-copy**: Works with string slices directly

Typical memory usage:
- 1K rows: ~10KB
- 10K rows: ~100KB
- 100K rows: ~1MB (with sampling)

### Scalability

The engine scales linearly with data size:

```
Time ≈ O(n × m)
```

Where:
- `n` = number of values (with sampling applied)
- `m` = number of type checks (typically 5)

Sampling ensures that processing time doesn't grow unbounded for very large datasets.

## Type Inference Results

### InferenceResult Structure

```go
type InferenceResult struct {
    InferredType  DataType           // Most likely type
    Confidence    float64            // Confidence level (0.0 to 1.0)
    Probabilities []TypeProbability  // All type probabilities
    SampleSize    int                // Number of values analyzed
}
```

### TypeProbability Structure

```go
type TypeProbability struct {
    Type        DataType  // Type identifier
    Probability float64   // Posterior probability
    Evidence    float64   // Likelihood from data
}
```

### Interpreting Results

- **Confidence ≥ 0.80**: High confidence, type is reliable
- **Confidence 0.60-0.80**: Medium confidence, consider validation
- **Confidence < 0.60**: Low confidence, likely mixed or ambiguous data

## Supported Data Types

### Integer
- Format: Whole numbers
- Examples: `1`, `42`, `-100`
- Range: int64 (-9,223,372,036,854,775,808 to 9,223,372,036,854,775,807)

### Float
- Format: Decimal numbers
- Examples: `3.14`, `2.5`, `-0.001`
- Precision: float64

### String
- Format: Any text
- Examples: `"hello"`, `"world"`, `"abc123"`
- Default fallback type

### Date
- Formats supported:
  - ISO 8601: `2006-01-02`
  - US format: `01/02/2006`
  - EU format: `02-01-2006`
  - Long format: `January 2 2006`
  - RFC3339: `2006-01-02T15:04:05Z`

### Boolean
- Formats supported:
  - `true`, `false`
  - `yes`, `no`
  - `y`, `n`
  - `t`, `f`
  - `1`, `0`
- Case-insensitive

## Configuration Options

### InferenceConfig

| Option            | Type    | Default  | Description                           |
|-------------------|---------|----------|---------------------------------------|
| Locale            | Locale  | Global   | Regional data patterns                |
| SampleSize        | int     | 1000     | Max values to analyze                 |
| ConfidenceThresh  | float64 | 0.80     | Minimum confidence to accept type     |
| RandomSeed        | int64   | 42       | Seed for deterministic sampling       |
| EnableSIMD        | bool    | true     | Enable SIMD acceleration              |

### Locale Options

- `LocaleUS`: United States patterns
- `LocaleEU`: European patterns
- `LocaleASIA`: Asian patterns
- `LocaleGlobal`: General patterns (default)

## Advanced Features

### Entropy Calculation

Calculate uncertainty in type inference:

```go
result := engine.InferType(values)
entropy := engine.CalculateEntropy(result.Probabilities)

// Low entropy (< 0.5) = high certainty
// High entropy (> 2.0) = high uncertainty
```

### Confidence Checking

```go
confidence := engine.GetConfidence(values)
if confidence < 0.80 {
    // Consider additional validation or manual inspection
}
```

### Batch Processing

Process multiple columns efficiently:

```go
columns := [][]string{column1, column2, column3}
results := engine.ParallelLikelihoodBatch(columns)
```

## Testing

### Running Tests

```bash
# Run all tests
go test ./tests/... -v

# Run specific test
go test ./tests/ -run TestBayesianInferenceEngine_InferType -v

# Run with coverage
go test ./tests/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Running Benchmarks

```bash
# Run all benchmarks
go test ./tests/ -bench=. -benchmem

# Run specific benchmark
go test ./tests/ -bench=BenchmarkBayesianInference_1000 -benchmem

# Run SIMD comparison
go test ./tests/ -bench=BenchmarkSIMD -benchmem
```

## Best Practices

1. **Use Appropriate Sample Size**: Balance accuracy vs performance
   - Small datasets (<1K): Use all data
   - Medium datasets (1K-10K): Sample 1K values
   - Large datasets (>10K): Sample 1K-5K values

2. **Set Confidence Threshold**: Match your use case
   - Critical data: 0.90+
   - Production systems: 0.80 (default)
   - Exploratory analysis: 0.60

3. **Choose Correct Locale**: Match your data source
   - US data: `LocaleUS`
   - EU data: `LocaleEU`
   - Asian data: `LocaleASIA`
   - Unknown/mixed: `LocaleGlobal` (default)

4. **Enable SIMD for Large Data**: Keep enabled by default
   - Automatically optimizes for large datasets
   - No overhead for small datasets
   - Provides 1.5-2x speedup

5. **Use Deterministic Seeds**: For reproducible results
   - Testing: Use fixed seed
   - Production: Use fixed seed for consistency
   - Experimentation: Try different seeds

## Troubleshooting

### Low Confidence Results

**Problem**: Inference returns low confidence (<0.60)

**Solutions**:
1. Check for mixed data types in column
2. Verify data quality (malformed values)
3. Try different locale settings
4. Examine probability distribution
5. Consider data cleaning

### Unexpected Type Detection

**Problem**: Wrong type detected

**Solutions**:
1. Review prior probabilities for locale
2. Check sample size (may need larger sample)
3. Verify data format matches expectations
4. Use `InferenceResult.Probabilities` to see all candidates
5. Adjust confidence threshold

### Performance Issues

**Problem**: Inference is too slow

**Solutions**:
1. Ensure SIMD is enabled
2. Reduce sample size for very large datasets
3. Use batch processing for multiple columns
4. Profile to identify bottlenecks
5. Consider caching results

## Examples

### Example 1: CSV Column Type Detection

```go
func detectCSVColumnTypes(csvData [][]string) map[string]DataType {
    engine := engine.NewBayesianInferenceEngine(engine.DefaultConfig())
    types := make(map[string]DataType)
    
    // Assume first row is header
    headers := csvData[0]
    
    // Process each column
    for i, header := range headers {
        column := make([]string, len(csvData)-1)
        for j := 1; j < len(csvData); j++ {
            if i < len(csvData[j]) {
                column[j-1] = csvData[j][i]
            }
        }
        
        result := engine.InferType(column)
        types[header] = result.InferredType
    }
    
    return types
}
```

### Example 2: Locale-Specific Detection

```go
func detectWithLocale(values []string, locale engine.Locale) *engine.InferenceResult {
    config := engine.DefaultConfig()
    config.Locale = locale
    
    engine := engine.NewBayesianInferenceEngine(config)
    return engine.InferType(values)
}

// Usage
usResult := detectWithLocale(data, engine.LocaleUS)
euResult := detectWithLocale(data, engine.LocaleEU)
```

### Example 3: Performance Monitoring

```go
import "time"

func inferWithTiming(values []string) (*engine.InferenceResult, time.Duration) {
    engine := engine.NewBayesianInferenceEngine(engine.DefaultConfig())
    
    start := time.Now()
    result := engine.InferType(values)
    elapsed := time.Since(start)
    
    return result, elapsed
}

// Usage
result, duration := inferWithTiming(largeDataset)
fmt.Printf("Inferred %v in %v\n", result.InferredType, duration)
```

## Future Enhancements

Potential areas for future development:

1. **Additional Types**:
   - UUID detection
   - Email/URL validation
   - Currency detection
   - Geographic coordinates

2. **Machine Learning**:
   - Learn priors from historical data
   - Adaptive confidence thresholds
   - Pattern recognition for custom types

3. **Distributed Processing**:
   - Parallel column processing
   - Distributed inference for massive datasets
   - GPU acceleration

4. **Enhanced SIMD**:
   - AVX-512 support
   - ARM NEON optimizations
   - Custom assembly for hot paths

## References

- Bayesian Inference: https://en.wikipedia.org/wiki/Bayesian_inference
- SIMD Programming: https://en.wikipedia.org/wiki/SIMD
- Go Performance: https://go.dev/doc/effective_go#optimization

## License

See LICENSE file in repository root.

## Contributing

Contributions welcome! Please see CONTRIBUTING.md for guidelines.
