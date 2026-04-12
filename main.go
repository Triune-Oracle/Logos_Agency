package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

// EnablePerformanceLogging controls whether performance metrics are logged
var EnablePerformanceLogging = os.Getenv("ENABLE_PERF_LOGGING") == "1"

var dateFormats = []string{
	"2006-01-02",
	"01/02/2006",
	"02-01-2006",
	"Jan 2 2006",
	"January 2 2006",
	time.RFC3339,
}

// extendedDateFormats includes additional formats for HeuristicScanner
var extendedDateFormats = []string{
	"2006-01-02",
	"01/02/2006",
	"02/01/2006",
	"02-Jan-2006",
	"02-January-2006",
	"Jan 2 2006",
	"January 2 2006",
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05",
	time.RFC3339,
}

// booleanValues maps lowercase string representations to bool
var booleanValues = map[string]bool{
	"true": true, "false": true,
	"yes": true, "no": true,
	"t": true, "f": true,
	"y": true, "n": true,
	"1": true, "0": true,
}

// HeuristicScanner provides probabilistic column type inference with configurable confidence thresholds.
type HeuristicScanner struct {
	// ConfidenceThreshold is the minimum fraction of non-null values that must match a type
	// for that type to be returned (default 0.95).
	ConfidenceThreshold float64
}

// NewHeuristicScanner creates a HeuristicScanner with the default 95% confidence threshold.
func NewHeuristicScanner() *HeuristicScanner {
	return &HeuristicScanner{ConfidenceThreshold: 0.95}
}

// InferColumnType infers the SQL type of a column and returns the type name together
// with the fraction of non-null values that matched the inferred type.
//
// Returned type names: "INTEGER", "DECIMAL(p,s)", "TIMESTAMP", "DATE", "BOOLEAN", "TEXT".
// If the best-matched type's confidence is below ConfidenceThreshold the function
// returns "TEXT" with that confidence score.
func (hs *HeuristicScanner) InferColumnType(values []string) (string, float64) {
	type counts struct{ boolean, integer, decimal, timestamp, date int }
	var c counts
	var maxPrecision, maxScale int

	nonNull := 0
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" || strings.EqualFold(v, "null") || strings.EqualFold(v, "nil") {
			continue
		}
		nonNull++
		lower := strings.ToLower(v)

		if booleanValues[lower] {
			c.boolean++
		}
		if _, err := strconv.ParseInt(v, 10, 64); err == nil {
			c.integer++
		}
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			// Reject values too large to represent without overflow
			if !math.IsInf(f, 0) && !math.IsNaN(f) {
				p, s := decimalPrecisionScale(v)
				c.decimal++
				if p > maxPrecision {
					maxPrecision = p
				}
				if s > maxScale {
					maxScale = s
				}
			}
		}

		// Check timestamp (date + time) before plain date
		for _, fmt := range extendedDateFormats {
			if t, err := time.Parse(fmt, v); err == nil {
				h, m, s := t.Clock()
				if h != 0 || m != 0 || s != 0 {
					c.timestamp++
				} else {
					c.date++
				}
				break
			}
		}
	}

	if nonNull == 0 {
		return "TEXT", 0.0
	}

	// Build candidates in specificity order so the most descriptive type wins
	type candidate struct {
		typeName   string
		matchCount int
	}
	decimalType := "DECIMAL"
	if maxPrecision > 0 {
		decimalType = fmt.Sprintf("DECIMAL(%d,%d)", maxPrecision, maxScale)
	}
	candidates := []candidate{
		{"BOOLEAN", c.boolean},
		{"INTEGER", c.integer},
		{decimalType, c.decimal},
		{"TIMESTAMP", c.timestamp},
		{"DATE", c.date},
	}

	// Pick the first candidate whose confidence meets the threshold
	for _, cand := range candidates {
		conf := float64(cand.matchCount) / float64(nonNull)
		if conf >= hs.ConfidenceThreshold {
			return cand.typeName, conf
		}
	}

	// No type met threshold; return TEXT with the confidence of the best non-boolean match
	bestConf := 0.0
	for _, cand := range candidates[1:] { // skip boolean
		conf := float64(cand.matchCount) / float64(nonNull)
		if conf > bestConf {
			bestConf = conf
		}
	}
	return "TEXT", bestConf
}

// decimalPrecisionScale returns (total significant digits, digits after decimal point) for a
// numeric string.  It handles sign, leading zeros, and trailing zeros.
func decimalPrecisionScale(v string) (precision, scale int) {
	v = strings.TrimSpace(v)
	// Strip sign
	if len(v) > 0 && (v[0] == '+' || v[0] == '-') {
		v = v[1:]
	}
	// Remove exponent part for precision counting
	if idx := strings.IndexAny(v, "eE"); idx >= 0 {
		v = v[:idx]
	}
	dotIdx := strings.Index(v, ".")
	if dotIdx < 0 {
		// Integer-like – count significant digits (strip leading zeros)
		digits := strings.TrimLeft(v, "0")
		if digits == "" {
			digits = "0"
		}
		return len(digits), 0
	}
	intPart := v[:dotIdx]
	fracPart := v[dotIdx+1:]

	// Count significant digits in integer part (strip leading zeros)
	intSig := strings.TrimLeft(intPart, "0")
	intDigits := len(intSig)

	// Fractional digits (keep trailing zeros as they indicate precision)
	fracDigits := len(fracPart)

	// Filter out non-digit characters just in case
	intDigits = countDigits(intSig)
	fracDigits = countDigits(fracPart)

	scale = fracDigits
	precision = intDigits + fracDigits
	if precision == 0 {
		precision = 1 // at least "0"
	}
	return precision, scale
}

// countDigits counts the number of digit characters in s.
func countDigits(s string) int {
	n := 0
	for _, r := range s {
		if unicode.IsDigit(r) {
			n++
		}
	}
	return n
}

// TypeCache provides thread-safe caching for column type inferences
type TypeCache struct {
	sync.RWMutex
	cache map[string]string
}

// NewTypeCache creates a new type cache
func NewTypeCache() *TypeCache {
	return &TypeCache{
		cache: make(map[string]string),
	}
}

// GetOrInfer gets cached type or computes it
func (c *TypeCache) GetOrInfer(key string, inferFn func() string) string {
	// Try read lock first
	c.RLock()
	if typ, ok := c.cache[key]; ok {
		c.RUnlock()
		return typ
	}
	c.RUnlock()
	
	// Compute with write lock
	c.Lock()
	defer c.Unlock()
	
	// Double-check pattern
	if typ, ok := c.cache[key]; ok {
		return typ
	}
	
	typ := inferFn()
	c.cache[key] = typ
	return typ
}

// Clear removes all cached entries
func (c *TypeCache) Clear() {
	c.Lock()
	defer c.Unlock()
	c.cache = make(map[string]string)
}

func InferColumnType(values []string) string {
	start := time.Now()
	allInt := true
	allFloat := true
	allDate := true
	hasNonEmpty := false

	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue // Skip empty values instead of early return
		}
		hasNonEmpty = true

		// Early termination: if all types are ruled out, return immediately
		if !allInt && !allFloat && !allDate {
			return "string"
		}

		// Only check int if we haven't ruled it out
		if allInt {
			if _, err := strconv.Atoi(v); err != nil {
				allInt = false
				// All ints are valid floats, so we still need to check float
			}
		}

		// Only check float if we haven't ruled it out
		if allFloat {
			if _, err := strconv.ParseFloat(v, 64); err != nil {
				allFloat = false
				allInt = false // If not a valid float, can't be an int either
			}
		}

		// Only check date if we haven't ruled it out
		if allDate {
			parsedAsDate := false
			for _, f := range dateFormats {
				if _, err := time.Parse(f, v); err == nil {
					parsedAsDate = true
					break
				}
			}
			if !parsedAsDate {
				allDate = false
			}
		}
	}

	// If all values are empty, return string
	if !hasNonEmpty {
		return "string"
	}

	var result string
	if allInt {
		result = "int"
	} else if allFloat {
		result = "float"
	} else if allDate {
		result = "date"
	} else {
		result = "string"
	}
	
	// Structured logging with performance context
	duration := time.Since(start)
	if EnablePerformanceLogging && len(values) > 1000 { // Only log for larger datasets
		log.Printf("CSV type inference: rows=%d, type=%s, duration_ms=%d", 
			len(values), result, duration.Milliseconds())
	}
	
	return result
}

func generateRows(n int) []string {
	out := make([]string, n)
	for i := 0; i < n; i++ {
		out[i] = strconv.Itoa(i)
	}
	return out
}

// InferColumnTypesStreaming infers types from a reader line by line
// This is more memory-efficient for large CSV files
func InferColumnTypesStreaming(lines []string, sampleSize int) []string {
	if len(lines) == 0 {
		return []string{}
	}
	
	// Determine number of columns from first line
	firstLine := strings.Split(lines[0], ",")
	numCols := len(firstLine)
	
	// Collect samples for each column
	samples := make([][]string, numCols)
	for i := range samples {
		samples[i] = make([]string, 0, sampleSize)
	}
	
	// Sample rows (skip header if present)
	startIdx := 1
	if len(lines) > 0 {
		// Simple heuristic: if first row has non-numeric values, treat as header
		firstCell := strings.TrimSpace(firstLine[0])
		if _, err := strconv.ParseFloat(firstCell, 64); err != nil {
			startIdx = 1
		} else {
			startIdx = 0
		}
	}
	
	for i := startIdx; i < len(lines) && i < startIdx+sampleSize; i++ {
		fields := strings.Split(lines[i], ",")
		for j := 0; j < numCols && j < len(fields); j++ {
			samples[j] = append(samples[j], strings.TrimSpace(fields[j]))
		}
	}
	
	// Infer type for each column
	types := make([]string, numCols)
	for i, sample := range samples {
		types[i] = InferColumnType(sample)
	}
	
	return types
}
