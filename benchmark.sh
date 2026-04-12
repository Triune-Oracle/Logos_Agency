#!/usr/bin/env bash
# benchmark.sh – Logos Agency comprehensive benchmark harness
#
# Usage:
#   ./benchmark.sh               # run all Go benchmarks, save results
#   ./benchmark.sh --compare     # also run the Python pandas/csvkit comparison
#   ./benchmark.sh --help        # show this message
#
# Output files:
#   bench_results.txt            – raw Go benchmark output
#   bench_results_python.txt     – Python comparison table (--compare only)
#
# Requirements:
#   - Go 1.21+ in PATH
#   - (optional) pip install pandas csvkit   for --compare mode

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RESULTS_FILE="${REPO_ROOT}/bench_results.txt"
PYTHON_SCRIPT="${REPO_ROOT}/benchmarks/benchmark_pandas_comparison.py"

# ---------------------------------------------------------------------------
print_header() {
    echo "=================================================="
    echo " Logos Agency – Benchmark Suite"
    echo " $(date -u '+%Y-%m-%dT%H:%M:%SZ')"
    echo "=================================================="
}

run_go_benchmarks() {
    echo ""
    echo "── Go Benchmarks ───────────────────────────────"
    echo "Running: go test -bench=. -benchmem -benchtime=3s -count=3 ."
    echo ""

    cd "${REPO_ROOT}"
    go test \
        -bench=. \
        -benchmem \
        -benchtime=3s \
        -count=3 \
        -run="^$" \
        . | tee "${RESULTS_FILE}"

    echo ""
    echo "Results saved to: ${RESULTS_FILE}"
}

run_comparison_benchmarks() {
    echo ""
    echo "── Comparison Benchmarks (Optimised vs Naive) ──"
    echo "Running: go test -bench=BenchmarkComparison -benchmem -benchtime=2s ."
    echo ""

    cd "${REPO_ROOT}"
    go test \
        -bench=BenchmarkComparison \
        -benchmem \
        -benchtime=2s \
        -count=1 \
        -run="^$" \
        . | tee -a "${RESULTS_FILE}"
}

run_speedup_test() {
    echo ""
    echo "── Speedup Assertion Test ───────────────────────"
    echo "Running: go test -run TestSpeedup -v ."
    echo ""

    cd "${REPO_ROOT}"
    go test -run TestSpeedup -v . | tee -a "${RESULTS_FILE}"
}

run_python_comparison() {
    echo ""
    echo "── Python Comparison (pandas / csvkit) ──────────"

    if ! command -v python3 &>/dev/null; then
        echo "python3 not found – skipping Python comparison"
        return
    fi

    python3 "${PYTHON_SCRIPT}" | tee -a "${RESULTS_FILE}"
}

print_summary() {
    echo ""
    echo "=================================================="
    echo " Summary"
    echo "=================================================="
    if [[ -f "${RESULTS_FILE}" ]]; then
        # Extract key lines for a quick summary
        echo ""
        echo "Key single-column benchmarks (integer data):"
        grep -E "BenchmarkComparison_(Optimised|Naive).*Int" "${RESULTS_FILE}" | tail -8 || true
        echo ""
        echo "Full results: ${RESULTS_FILE}"
    fi
    echo ""
    echo "See docs/PERFORMANCE.md for analysis and comparison tables."
}

# ---------------------------------------------------------------------------
# Entry point
# ---------------------------------------------------------------------------

COMPARE_MODE=false

for arg in "$@"; do
    case "$arg" in
        --compare) COMPARE_MODE=true ;;
        --help)
            grep "^#" "$0" | sed 's/^# \?//'
            exit 0
            ;;
    esac
done

print_header
run_go_benchmarks
run_comparison_benchmarks
run_speedup_test

if [[ "${COMPARE_MODE}" == "true" ]]; then
    run_python_comparison
fi

print_summary

