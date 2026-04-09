#!/usr/bin/env bash
# simple local benchmark harness (skeleton)
go test -bench=. -benchmem ./... | tee bench_results.txt
