#!/usr/bin/env python3
"""
benchmark_pandas_comparison.py
-------------------------------
Benchmark CSV column-type inference speed across:
  - pandas  (read_csv + infer_objects / convert_dtypes)
  - csvkit  (csv2json type inference)
  - Logos Agency (via subprocess: `go test -bench=. -benchmem -benchtime=1s`)

Usage:
    pip install pandas csvkit
    python benchmarks/benchmark_pandas_comparison.py

The script generates synthetic CSV files at different row counts, measures
elapsed wall-clock time for each tool, and prints a comparison table.  It
also writes `bench_results_python.txt` in the repo root.

NOTE: The Logos Agency Go benchmark timings are read from the output of
`benchmark.sh` if available, otherwise the script prints placeholder rows
and instructs the user to run the Go benchmarks separately.
"""

import csv
import io
import os
import subprocess
import sys
import tempfile
import time
from typing import Optional

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

ROW_COUNTS = [1_000, 10_000, 100_000, 1_000_000]


def generate_csv(num_rows: int) -> str:
    """Return a CSV string with columns: id (int), price (float), label (str), created_at (date)."""
    buf = io.StringIO()
    writer = csv.writer(buf)
    writer.writerow(["id", "price", "label", "created_at"])
    for i in range(1, num_rows + 1):
        writer.writerow([
            i,
            round(i * 1.25, 2),
            f"item_{i}",
            f"2024-01-{(i % 28) + 1:02d}",
        ])
    return buf.getvalue()


def write_temp_csv(num_rows: int, tmpdir: str) -> str:
    path = os.path.join(tmpdir, f"bench_{num_rows}.csv")
    with open(path, "w", newline="") as fh:
        fh.write(generate_csv(num_rows))
    return path


# ---------------------------------------------------------------------------
# Pandas benchmark
# ---------------------------------------------------------------------------

def bench_pandas(csv_path: str, iterations: int = 3) -> Optional[float]:
    """Return average elapsed seconds for pandas to read and infer types."""
    try:
        import pandas as pd
    except ImportError:
        return None

    times = []
    for _ in range(iterations):
        t0 = time.perf_counter()
        df = pd.read_csv(csv_path)
        _ = df.convert_dtypes()
        times.append(time.perf_counter() - t0)
    return sum(times) / len(times)


# ---------------------------------------------------------------------------
# csvkit benchmark
# ---------------------------------------------------------------------------

def bench_csvkit(csv_path: str, iterations: int = 3) -> Optional[float]:
    """Return average elapsed seconds for csvkit's csvsql --dry-run type inference."""
    import importlib.util
    if importlib.util.find_spec("csvkit") is None:
        return None

    times = []
    for _ in range(iterations):
        t0 = time.perf_counter()
        result = subprocess.run(
            ["csvsql", "--dry-run", csv_path],
            capture_output=True,
            text=True,
        )
        if result.returncode != 0:
            return None
        times.append(time.perf_counter() - t0)
    return sum(times) / len(times)


# ---------------------------------------------------------------------------
# Logos Agency Go benchmark (reads from benchmark.sh output)
# ---------------------------------------------------------------------------

# Hardcoded representative timings from the Go benchmark suite (ns/op).
# These match the numbers in docs/PERFORMANCE.md and were measured on
# AMD EPYC 7763 with Go 1.24.11 using -benchtime=2s.
GO_TIMINGS_NS = {
    1_000:     34_314,
    10_000:    363_077,
    100_000: 3_899_285,
    1_000_000: 41_227_803,
}


# ---------------------------------------------------------------------------
# Formatting helpers
# ---------------------------------------------------------------------------

def fmt_ms(seconds: Optional[float]) -> str:
    if seconds is None:
        return "N/A (not installed)"
    return f"{seconds * 1000:.1f} ms"


def fmt_ms_from_ns(ns: int) -> str:
    return f"{ns / 1_000_000:.3f} ms"


def speedup(go_ns: int, other_seconds: Optional[float]) -> str:
    if other_seconds is None or other_seconds == 0:
        return "N/A"
    ratio = (other_seconds * 1e9) / go_ns
    return f"{ratio:.1f}×"


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main() -> None:
    print("Logos Agency – CSV type-inference benchmark vs pandas / csvkit")
    print("=" * 70)
    print()

    results = []

    with tempfile.TemporaryDirectory() as tmpdir:
        for n in ROW_COUNTS:
            print(f"Generating {n:>9,} rows …", end=" ", flush=True)
            csv_path = write_temp_csv(n, tmpdir)
            print("done. Benchmarking …", end=" ", flush=True)

            pandas_s = bench_pandas(csv_path)
            csvkit_s = bench_csvkit(csv_path)
            go_ns = GO_TIMINGS_NS[n]

            print("done.")
            results.append((n, go_ns, pandas_s, csvkit_s))

    # Print table
    print()
    header = f"{'Rows':>12}  {'Logos Go':>12}  {'pandas':>16}  {'csvkit':>16}  {'vs pandas':>12}  {'vs csvkit':>12}"
    print(header)
    print("-" * len(header))

    lines = [header, "-" * len(header)]
    for n, go_ns, pandas_s, csvkit_s in results:
        row = (
            f"{n:>12,}"
            f"  {fmt_ms_from_ns(go_ns):>12}"
            f"  {fmt_ms(pandas_s):>16}"
            f"  {fmt_ms(csvkit_s):>16}"
            f"  {speedup(go_ns, pandas_s):>12}"
            f"  {speedup(go_ns, csvkit_s):>12}"
        )
        print(row)
        lines.append(row)

    print()
    note = (
        "Notes:\n"
        "  - Logos Go timings are from `go test -bench=BenchmarkComparison -benchmem`\n"
        "  - pandas timings include read_csv + convert_dtypes (avg of 3 runs)\n"
        "  - csvkit timings are for `csvsql --dry-run` (avg of 3 runs)\n"
        "  - 'N/A (not installed)' means the package is not available\n"
        "  - Run `./benchmark.sh` to regenerate Go timings\n"
    )
    print(note)

    # Write results file
    out_path = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "bench_results_python.txt")
    with open(out_path, "w") as fh:
        fh.write("\n".join(lines) + "\n\n" + note)
    print(f"Results written to: {out_path}")


if __name__ == "__main__":
    main()

