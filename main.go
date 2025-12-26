package main

import (
	"strconv"
	"time"
	"strings"
	"sync"
	"log"
	"os"
)

// EnablePerformanceLogging controls whether performance metrics are logged
var EnablePerformanceLogging = os.Getenv("ENABLE_PERF_LOGGING") == "1"

var dateFormats = []string{
	"2006-01-02",
	"01/02/2006",
	"02-01-2006",
	"Jan 2 2006",
	time.RFC3339,
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

	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			return "string"
		}

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
