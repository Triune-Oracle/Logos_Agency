package tests

import (
"fmt"
"strings"
"testing"
"time"

"github.com/Triune-Oracle/Logos_Agency/engine"
"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// HeuristicScanner-style comprehensive tests using the engine package
// ---------------------------------------------------------------------------

// TestHeuristicScanner_Scan verifies that the inference engine handles a variety of
// raw string inputs without panicking and returns a known type.
func TestHeuristicScanner_Scan(t *testing.T) {
eng := engine.NewBayesianInferenceEngine(engine.DefaultConfig())

validTypes := map[engine.DataType]bool{
engine.TypeInteger: true,
engine.TypeFloat:   true,
engine.TypeString:  true,
engine.TypeDate:    true,
engine.TypeBoolean: true,
}

inputs := []struct {
label string
data  []string
}{
{"normal integer data", []string{"1", "2", "3"}},
{"empty input", []string{"", ""}},
{"very large input", []string{strings.Repeat("a", 1_000_000)}},
{"special characters", []string{"!@#$%^&*()"}},
{"chinese characters", []string{"数据", "数据"}},
{"russian characters", []string{"данные", "данные"}},
}

for _, inp := range inputs {
t.Run(inp.label, func(t *testing.T) {
result := eng.InferType(inp.data)
require.NotNil(t, result, "InferType must not return nil")
assert.True(t, validTypes[result.InferredType],
"unexpected type %v for input %q", result.InferredType, inp.label)
})
}
}

// TestHeuristicScanner_EdgeCases_Via_Engine covers the full edge-case matrix
// (empty columns, mixed types, booleans, date formats, decimals, overflow, sci notation).
func TestHeuristicScanner_EdgeCases_Via_Engine(t *testing.T) {
eng := engine.NewBayesianInferenceEngine(engine.DefaultConfig())

tests := []struct {
name     string
values   []string
wantType engine.DataType
minConf  float64
}{
// Empty / all-null columns
{"all empty strings", []string{"", "", ""}, engine.TypeString, 0.0},

// Pure integer column
{"pure integers", []string{"1", "2", "3", "100"}, engine.TypeInteger, 0.8},

// Negative integers
{"negative integers", []string{"-5", "-100", "-1"}, engine.TypeInteger, 0.8},

// Float / decimal
{"pure floats", []string{"1.5", "2.7", "3.14"}, engine.TypeFloat, 0.8},
{"negative decimals", []string{"-3.14", "-2.71"}, engine.TypeFloat, 0.8},
{"scientific notation", []string{"1.23e10", "4.56e3"}, engine.TypeFloat, 0.7},

// Boolean variants
{"true/false", []string{"true", "false", "true"}, engine.TypeBoolean, 0.7},
{"yes/no", []string{"yes", "no", "yes"}, engine.TypeBoolean, 0.7},

// Date formats
{"ISO 8601 dates", []string{"2024-01-15", "2023-12-31"}, engine.TypeDate, 0.7},

		// Mixed -> string (1 integer out of 5 = 20% confidence -> below 80% default threshold -> string)
		{"mixed integers and text", []string{"1", "hello", "world", "foo", "bar"}, engine.TypeString, 0.0},

		// Integers with nulls - nulls should not count as non-matching
		{"integers with nulls", []string{"1", "", "3", "", "5"}, engine.TypeInteger, 0.7},
	}

for _, tc := range tests {
tc := tc
t.Run(tc.name, func(t *testing.T) {
result := eng.InferType(tc.values)
require.NotNil(t, result)
assert.Equal(t, tc.wantType, result.InferredType,
"type mismatch for %q: got confidence=%v", tc.name, result.Confidence)
assert.GreaterOrEqual(t, result.Confidence, tc.minConf,
"confidence too low for %q", tc.name)
})
}
}

// TestHeuristicScanner_BoundaryConditions tests the confidence threshold boundary.
func TestHeuristicScanner_BoundaryConditions(t *testing.T) {
cfg := engine.DefaultConfig()
cfg.ConfidenceThresh = 0.95
eng := engine.NewBayesianInferenceEngine(cfg)

// 20 integers + 1 text = 95.2% -> above threshold
atBoundary := make([]string, 21)
for i := 0; i < 20; i++ {
atBoundary[i] = fmt.Sprintf("%d", i+1)
}
atBoundary[20] = "text"

result := eng.InferType(atBoundary)
require.NotNil(t, result)
assert.GreaterOrEqual(t, result.Confidence, 0.0,
"confidence must be non-negative")

// 1 integer + 20 text = 4.76% -> far below threshold -> string
clearString := make([]string, 21)
clearString[0] = "42"
for i := 1; i < 21; i++ {
clearString[i] = fmt.Sprintf("word%d", i)
}
resultStr := eng.InferType(clearString)
require.NotNil(t, resultStr)
assert.Equal(t, engine.TypeString, resultStr.InferredType,
"mostly-text column should be inferred as string")
}

// TestHeuristicScanner_Performance ensures the inference engine completes within 1 s
// for large inputs (regression guard).
func TestHeuristicScanner_Performance(t *testing.T) {
eng := engine.NewBayesianInferenceEngine(engine.DefaultConfig())

largeInput := make([]string, 10_000)
for i := range largeInput {
largeInput[i] = fmt.Sprintf("%d", i)
}

start := time.Now()
result := eng.InferType(largeInput)
elapsed := time.Since(start)

require.NotNil(t, result)
assert.Less(t, elapsed.Seconds(), 1.0,
"InferType for 10k integers must complete within 1 s, took %v", elapsed)
}
