package main

import (
	"testing"
	"time"
)

func TestHeuristicScanner_InferColumnType(t *testing.T) {
	tests := []struct {
		name     string
		values   []string
		expected string
	}{
		{"Integer Detection", []string{"1", "2", "3"}, "int"},
		{"Float Detection", []string{"1.0", "2.5", "3.14"}, "float"},
		{"Mixed Int/Float", []string{"1", "2.5"}, "float"},
		{"String Detection", []string{"hello", "world"}, "string"},
		{"Date ISO", []string{"2024-01-01", "2023-02-05"}, "date"},
		{"Date Slash", []string{"12/31/2024", "01/01/2025"}, "date"},
		{"Empty Cells", []string{"", "", ""}, "string"},
		{"Malformed Rows",
			[]string{"7", "???", "9"}, "string"},
		{"Mixed Types", []string{"1", "hello", "3.14"}, "string"},
		{"Long Month", []string{"Jan 1 2020", "Feb 2 2021"}, "date"},
		// Test early termination optimization
		{"Early String Detection", []string{"text", "more text", "100", "200"}, "string"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := InferColumnType(tc.values)
			if got != tc.expected {
				t.Fatalf("expected %s, got %s", tc.expected, got)
			}
		})
	}
}

func BenchmarkHeuristicScanner_10k(b *testing.B) {
	data := generateRows(10000)
	for i := 0; i < b.N; i++ {
		InferColumnType(data)
	}
}

func BenchmarkHeuristicScanner_100k(b *testing.B) {
	data := generateRows(100000)
	for i := 0; i < b.N; i++ {
		InferColumnType(data)
	}
}

func BenchmarkHeuristicScanner_1M(b *testing.B) {
	data := generateRows(1000000)
	for i := 0; i < b.N; i++ {
		InferColumnType(data)
	}
}

// Benchmark regression test to ensure performance stays within acceptable bounds
func BenchmarkColumnTypeInferenceRegression(b *testing.B) {
	data := generateRows(10000)
	
	b.ResetTimer()
	start := time.Now()
	for i := 0; i < b.N; i++ {
		InferColumnType(data)
	}
	elapsed := time.Since(start)
	
	// Performance should stay under 50ms for 10k rows (averaged over iterations)
	avgTime := elapsed / time.Duration(b.N)
	maxAcceptable := 50 * time.Millisecond
	
	if avgTime > maxAcceptable {
		b.Errorf("Performance regression detected: avg %v exceeds threshold %v", avgTime, maxAcceptable)
	}
}

// Fuzz testing to ensure InferColumnType handles all inputs safely
func FuzzInferColumnType(f *testing.F) {
	// Add seed corpus with various edge cases
	f.Add("123")
	f.Add("3.14")
	f.Add("2024-01-01")
	f.Add("")
	f.Add("hello")
	f.Add("   ")
	f.Add("\n\t")
	f.Add("12/31/2024")
	
	f.Fuzz(func(t *testing.T, data string) {
		// Create a slice with the fuzzed data
		values := []string{data}
		
		// Should not panic or error - just return a valid type
		result := InferColumnType(values)
		
		// Result must be one of the valid types
		validTypes := map[string]bool{
			"int":    true,
			"float":  true,
			"date":   true,
			"string": true,
		}
		
		if !validTypes[result] {
			t.Errorf("InferColumnType returned invalid type: %s", result)
		}
	})
}

// Test streaming CSV type inference
func TestInferColumnTypesStreaming(t *testing.T) {
	tests := []struct {
		name       string
		lines      []string
		sampleSize int
		expected   []string
	}{
		{
			name: "Simple CSV with header",
			lines: []string{
				"id,name,age",
				"1,Alice,30",
				"2,Bob,25",
			},
			sampleSize: 10,
			expected:   []string{"int", "string", "int"},
		},
		{
			name: "Mixed types",
			lines: []string{
				"100,3.14,hello",
				"200,2.71,world",
			},
			sampleSize: 10,
			expected:   []string{"int", "float", "string"},
		},
		{
			name: "Empty input",
			lines: []string{},
			sampleSize: 10,
			expected:   []string{},
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := InferColumnTypesStreaming(tc.lines, tc.sampleSize)
			if len(got) != len(tc.expected) {
				t.Fatalf("expected %d columns, got %d", len(tc.expected), len(got))
			}
			for i := range got {
				if got[i] != tc.expected[i] {
					t.Errorf("column %d: expected %s, got %s", i, tc.expected[i], got[i])
				}
			}
		})
	}
}
