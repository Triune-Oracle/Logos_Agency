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
	time.RFC3339,
}

func InferColumnType(values []string) string {
	allInt := true
	allFloat := true
	allDate := true

	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			return "string"
		}

		if _, err := strconv.Atoi(v); err != nil {
			allInt = false
		}
		if _, err := strconv.ParseFloat(v, 64); err != nil {
			allFloat = false
		}
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
