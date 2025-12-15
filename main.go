package main

import (
	"strconv"
	"time"
	"strings"
)

var dateFormats = []string{
	"2006-01-02",
	"01/02/2006",
	"02-01-2006",
	"Jan 2 2006",
	"January 2 2006",
	time.RFC3339,
}

func InferColumnType(values []string) string {
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

		// Optimization: int parsing is redundant since ParseFloat covers integers
		// Only parse as int if we haven't ruled it out yet
		if allInt {
			if _, err := strconv.Atoi(v); err != nil {
				allInt = false
			}
		}

		// Check float only if not already ruled out and int failed
		if allFloat && !allInt {
			if _, err := strconv.ParseFloat(v, 64); err != nil {
				allFloat = false
			}
		} else if allInt {
			// If it's an int, it's also a valid float
			allFloat = true
		}

		// Only check dates if not already ruled out
		// This is the most expensive check, so do it last
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

		// Early exit optimization: if all types are ruled out, it's a string
		if !allInt && !allFloat && !allDate {
			return "string"
		}
	}

	// If all values were empty, treat as string
	if !hasNonEmpty {
		return "string"
	}

	if allInt {
		return "int"
	}
	if allFloat {
		return "float"
	}
	if allDate {
		return "date"
	}
	return "string"
}

func generateRows(n int) []string {
	out := make([]string, n)
	for i := 0; i < n; i++ {
		out[i] = strconv.Itoa(i)
	}
	return out
}
