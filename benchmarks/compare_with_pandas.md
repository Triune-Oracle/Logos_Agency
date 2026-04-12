# Benchmarks: Logos Agency vs pandas / csvkit

## Quick results (AMD EPYC 7763, Go 1.24.11)

| Rows | Logos Agency (Go) | pandas | csvkit | vs pandas | vs csvkit |
|------|-------------------|--------|--------|-----------|-----------|
| 1 000 | 0.034 ms | ~0.60 ms | ~1.2 ms | **17.6×** | **35×** |
| 10 000 | 0.363 ms | ~6.83 ms | ~12 ms | **18.8×** | **33×** |
| 100 000 | 3.90 ms | ~85.5 ms | ~150 ms | **21.9×** | **38×** |
| 1 000 000 | 41.2 ms | ~804 ms | ~1 400 ms | **19.5×** | **34×** |

> pandas and csvkit timings are estimates based on representative measurements.
> See `docs/PERFORMANCE.md` for full methodology and reproduction steps.

## How to reproduce

```bash
# Go benchmarks
go test -bench=BenchmarkComparison -benchmem -benchtime=2s .

# Python comparison (requires pip install pandas csvkit)
python benchmarks/benchmark_pandas_comparison.py

# Full automated harness
./benchmark.sh
```

## Why the difference is so large

| Factor | pandas/csvkit | Logos Agency |
|--------|--------------|--------------|
| Language | CPython (interpreted) | Compiled Go |
| Parsing strategy | Parse every value × every type | Early-exit on first type mismatch |
| Memory | O(rows) intermediate objects | O(1) constant per column |
| Multi-column | Sequential per column | Parallel-ready streaming |

For a detailed technical write-up see `docs/PERFORMANCE.md`.

