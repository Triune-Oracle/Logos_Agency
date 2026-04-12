package main

import (
	"fmt"
	"strings"
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

// ---------------------------------------------------------------------------
// HeuristicScanner comprehensive tests
// ---------------------------------------------------------------------------

// TestHeuristicScannerNew verifies that NewHeuristicScanner sets the default threshold.
func TestHeuristicScannerNew(t *testing.T) {
	hs := NewHeuristicScanner()
	if hs.ConfidenceThreshold != 0.95 {
		t.Errorf("expected default threshold 0.95, got %v", hs.ConfidenceThreshold)
	}
}

// TestHeuristicScanner_EdgeCases covers nulls, mixed types, boundary, overflow, sci notation,
// negatives, and boolean variants.
func TestHeuristicScanner_EdgeCases(t *testing.T) {
	hs := NewHeuristicScanner()

	tests := []struct {
		name        string
		values      []string
		wantType    string
		wantConfMin float64 // minimum acceptable confidence
		wantConfMax float64 // maximum acceptable confidence (0 = unchecked upper bound)
	}{
		// --- empty / all-null ---
		{
			name:        "all empty strings",
			values:      []string{"", "", "", ""},
			wantType:    "TEXT",
			wantConfMin: 0.0,
		},
		{
			name:        "all nulls",
			values:      []string{"null", "NULL", "Null"},
			wantType:    "TEXT",
			wantConfMin: 0.0,
		},
		{
			name:        "mix of empty and null",
			values:      []string{"", "null", "", "nil"},
			wantType:    "TEXT",
			wantConfMin: 0.0,
		},

		// --- pure integer ---
		{
			name:        "pure integers",
			values:      []string{"1", "2", "3", "100"},
			wantType:    "INTEGER",
			wantConfMin: 1.0,
		},
		{
			name:        "negative integers",
			values:      []string{"-5", "-100", "-1"},
			wantType:    "INTEGER",
			wantConfMin: 1.0,
		},

		// --- decimal / float ---
		{
			name:        "negative decimals",
			values:      []string{"-3.14", "-2.71", "-0.5"},
			wantType:    "DECIMAL(3,2)", // at least 3 sig digits, 2 after dot
			wantConfMin: 1.0,
		},
		{
			name:        "scientific notation",
			values:      []string{"1.23e10", "4.56e-3"},
			wantType:    "DECIMAL(3,2)",
			wantConfMin: 1.0,
		},

		// --- boolean variants ---
		{
			name:        "true/false boolean",
			values:      []string{"true", "false", "true"},
			wantType:    "BOOLEAN",
			wantConfMin: 1.0,
		},
		{
			name:        "yes/no boolean",
			values:      []string{"yes", "no", "yes"},
			wantType:    "BOOLEAN",
			wantConfMin: 1.0,
		},
		{
			name:        "Y/N boolean",
			values:      []string{"Y", "N", "Y"},
			wantType:    "BOOLEAN",
			wantConfMin: 1.0,
		},
		{
			name:        "T/F boolean",
			values:      []string{"T", "F", "T"},
			wantType:    "BOOLEAN",
			wantConfMin: 1.0,
		},
		{
			name:        "1/0 boolean",
			values:      []string{"1", "0", "1", "0"},
			wantType:    "BOOLEAN",
			wantConfMin: 1.0,
		},

		// --- date formats ---
		{
			name:        "ISO 8601 dates",
			values:      []string{"2024-01-15", "2023-12-31", "2022-06-01"},
			wantType:    "DATE",
			wantConfMin: 1.0,
		},
		{
			name:        "US format dates",
			values:      []string{"01/15/2024", "12/31/2023"},
			wantType:    "DATE",
			wantConfMin: 1.0,
		},
		{
			name:        "EU format dates",
			values:      []string{"15/01/2024", "31/12/2023"},
			wantType:    "DATE",
			wantConfMin: 1.0,
		},
		{
			name:        "named month dates",
			values:      []string{"15-Jan-2024", "31-Dec-2023"},
			wantType:    "DATE",
			wantConfMin: 1.0,
		},
		{
			name:        "timestamps with time component",
			values:      []string{"2024-01-15 14:30:00", "2023-12-31 23:59:59"},
			wantType:    "TIMESTAMP",
			wantConfMin: 1.0,
		},

		// --- mixed types fall through to TEXT ---
		{
			name:        "mixed integers and text",
			values:      []string{"1", "hello", "3", "world"},
			wantType:    "TEXT",
			wantConfMin: 0.0,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			gotType, gotConf := hs.InferColumnType(tc.values)
			// For DECIMAL types the exact DECIMAL(p,s) value varies; only check prefix.
			if strings.HasPrefix(tc.wantType, "DECIMAL") {
				if !strings.HasPrefix(gotType, "DECIMAL") {
					t.Errorf("expected DECIMAL type, got %s", gotType)
				}
			} else if gotType != tc.wantType {
				t.Errorf("expected type %s, got %s", tc.wantType, gotType)
			}
			if gotConf < tc.wantConfMin {
				t.Errorf("expected confidence >= %v, got %v", tc.wantConfMin, gotConf)
			}
			if tc.wantConfMax > 0 && gotConf > tc.wantConfMax {
				t.Errorf("expected confidence <= %v, got %v", tc.wantConfMax, gotConf)
			}
		})
	}
}

// TestHeuristicScanner_ConfidenceThreshold tests exactly-at and just-below 95% boundary.
func TestHeuristicScanner_ConfidenceThreshold(t *testing.T) {
	hs := NewHeuristicScanner() // threshold = 0.95

	// Build exactly 20 values: 19 integers + 1 text → 19/20 = 0.95 confidence
	exactlyAt := make([]string, 20)
	for i := 0; i < 19; i++ {
		exactlyAt[i] = fmt.Sprintf("%d", i+1)
	}
	exactlyAt[19] = "text"

	gotType, gotConf := hs.InferColumnType(exactlyAt)
	if gotType != "INTEGER" {
		t.Errorf("at 95%% confidence expected INTEGER, got %s (conf=%v)", gotType, gotConf)
	}
	if gotConf < 0.95 {
		t.Errorf("expected confidence >= 0.95, got %v", gotConf)
	}

	// 19 integers + 2 text = 19/21 ≈ 0.9047 confidence → below threshold → TEXT
	belowThreshold := append(exactlyAt[:], "text2") // 20 + 1 = 21
	gotType2, gotConf2 := hs.InferColumnType(belowThreshold)
	if gotType2 != "TEXT" {
		t.Errorf("below 95%% threshold expected TEXT, got %s (conf=%v)", gotType2, gotConf2)
	}
}

// TestHeuristicScanner_CustomThreshold tests that non-default thresholds (85%, 90%) work.
func TestHeuristicScanner_CustomThreshold(t *testing.T) {
	// 90 integers + 10 text → 90% confidence
	values := make([]string, 100)
	for i := 0; i < 90; i++ {
		values[i] = fmt.Sprintf("%d", i)
	}
	for i := 90; i < 100; i++ {
		values[i] = "text"
	}

	// At 85% threshold should detect INTEGER
	hs85 := &HeuristicScanner{ConfidenceThreshold: 0.85}
	if tp, _ := hs85.InferColumnType(values); tp != "INTEGER" {
		t.Errorf("at 85%% threshold expected INTEGER, got %s", tp)
	}

	// At 90% threshold (boundary) should detect INTEGER
	hs90 := &HeuristicScanner{ConfidenceThreshold: 0.90}
	if tp, _ := hs90.InferColumnType(values); tp != "INTEGER" {
		t.Errorf("at 90%% threshold boundary expected INTEGER, got %s", tp)
	}

	// At 95% threshold (default) should fall back to TEXT since confidence < 0.95
	hs95 := NewHeuristicScanner()
	if tp, _ := hs95.InferColumnType(values); tp != "TEXT" {
		t.Errorf("at 95%% threshold expected TEXT, got %s", tp)
	}
}

// TestHeuristicScanner_MixedWithNulls verifies that null/empty values are excluded from
// the confidence calculation, not counted as non-matching.
func TestHeuristicScanner_MixedWithNulls(t *testing.T) {
	hs := NewHeuristicScanner()
	// 5 integers + 5 nulls → 5/5 = 100% confidence on non-null values
	values := []string{"1", "", "3", "null", "5", "", "7", "null", "9", ""}
	gotType, gotConf := hs.InferColumnType(values)
	if gotType != "INTEGER" {
		t.Errorf("expected INTEGER with nulls excluded, got %s (conf=%v)", gotType, gotConf)
	}
	if gotConf < 0.95 {
		t.Errorf("expected confidence >= 0.95, got %v", gotConf)
	}
}

// TestHeuristicScanner_DecimalPrecision validates DECIMAL(p,s) precision/scale tracking.
func TestHeuristicScanner_DecimalPrecision(t *testing.T) {
	hs := NewHeuristicScanner()

	tests := []struct {
		name          string
		values        []string
		wantPrecision int
		wantScale     int
	}{
		{"pi to 5dp", []string{"3.14159"}, 6, 5},
		{"trailing zeros", []string{"1.50"}, 3, 2},
		{"large decimal", []string{"12345678901.23"}, 13, 2},
		{"negative decimal", []string{"-99.999"}, 5, 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotType, _ := hs.InferColumnType(tc.values)
			if !strings.HasPrefix(gotType, "DECIMAL") {
				t.Fatalf("expected DECIMAL type, got %s", gotType)
			}
			var p, s int
			if _, err := fmt.Sscanf(gotType, "DECIMAL(%d,%d)", &p, &s); err != nil {
				t.Fatalf("cannot parse type %q: %v", gotType, err)
			}
			if p < tc.wantPrecision {
				t.Errorf("precision: want >= %d, got %d", tc.wantPrecision, p)
			}
			if s < tc.wantScale {
				t.Errorf("scale: want >= %d, got %d", tc.wantScale, s)
			}
		})
	}
}

// TestHeuristicScanner_InvalidDates validates that invalid calendar dates are not typed as DATE.
func TestHeuristicScanner_InvalidDates(t *testing.T) {
	hs := NewHeuristicScanner()
	// Go's time.Parse rejects these automatically
	invalid := []string{"2024-02-30", "2023-13-01", "2024-00-15"}
	gotType, _ := hs.InferColumnType(invalid)
	if gotType == "DATE" || gotType == "TIMESTAMP" {
		t.Errorf("expected non-date type for invalid dates, got %s", gotType)
	}
}

// TestDecimalPrecisionScale unit-tests the helper in isolation.
func TestDecimalPrecisionScale(t *testing.T) {
	tests := []struct {
		v         string
		wantPrec  int
		wantScale int
	}{
		{"3.14159", 6, 5},
		{"1.50", 3, 2},
		{"42", 2, 0},
		{"-9.99", 3, 2},
		// Scientific notation: the exponent is stripped before counting digits,
		// so "1.23e10" is treated as mantissa "1.23" → precision 3, scale 2.
		{"1.23e10", 3, 2},
		{"0.001", 3, 3},
	}
	for _, tc := range tests {
		p, s := decimalPrecisionScale(tc.v)
		if p != tc.wantPrec || s != tc.wantScale {
			t.Errorf("decimalPrecisionScale(%q) = (%d,%d), want (%d,%d)",
				tc.v, p, s, tc.wantPrec, tc.wantScale)
		}
	}
}

// TestHeuristicScanner_VeryLargeNumbers checks that overflow values fall back to TEXT.
func TestHeuristicScanner_VeryLargeNumbers(t *testing.T) {
	hs := NewHeuristicScanner()
	// Values that exceed int64 range
	overflow := []string{"99999999999999999999999999999999", "88888888888888888888888888888888"}
	gotType, _ := hs.InferColumnType(overflow)
	// These should not be INTEGER (ParseInt64 overflows) nor DECIMAL (ParseFloat may Inf)
	if gotType == "INTEGER" {
		t.Errorf("overflow values should not be INTEGER, got %s", gotType)
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
