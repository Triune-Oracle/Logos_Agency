package main

import "testing"

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
