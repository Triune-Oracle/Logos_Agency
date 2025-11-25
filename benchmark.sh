'''#!/bin/bash

# benchmark.sh - Performance Comparison Suite

# --- Configuration ---
TEST_DURATION=5 # Duration in seconds for each test
CONCURRENCY=(1 5 10 20) # Number of concurrent users to simulate
API_ENDPOINT="http://localhost:3000/analyze/sentiment" # Placeholder for a real endpoint
PAYLOAD='{"text": "This is a test message for sentiment analysis."}'

# --- Dependencies Check ---
if ! command -v curl &> /dev/null
then
    echo "Error: curl is not installed. Please install it to run the benchmark."
    exit 1
fi

if ! command -v ab &> /dev/null
then
    echo "Warning: ApacheBench (ab) is not installed. Using basic curl timing."
    USE_AB=0
else
    USE_AB=1
fi

# --- Functions ---

run_ab_benchmark() {
    local users=$1
    echo "--- Running ApacheBench test with $users concurrent users for $TEST_DURATION seconds ---"
    ab -n 100000 -c $users -t $TEST_DURATION -p <(echo "$PAYLOAD") -T 'application/json' "$API_ENDPOINT"
}

run_curl_benchmark() {
    local users=$1
    echo "--- Running basic curl test with $users sequential requests ---"
    for i in $(seq 1 $users); do
        echo "Request $i/$users"
        time curl -s -X POST -H "Content-Type: application/json" -d "$PAYLOAD" "$API_ENDPOINT" > /dev/null
    done
}

# --- Main Execution ---

echo "Starting Performance Benchmark Suite for Logos_Agency"
echo "Target Endpoint: $API_ENDPOINT"
echo "Test Duration per concurrency level: $TEST_DURATION seconds"
echo "--------------------------------------------------------"

if [ $USE_AB -eq 1 ]; then
    for c in "${CONCURRENCY[@]}"; do
        run_ab_benchmark $c
    done
else
    echo "Falling back to sequential curl timing. Results will not reflect concurrency."
    run_curl_benchmark 5 # Run a small sequential test
fi

echo "--------------------------------------------------------"
echo "Benchmark Complete."
echo "Note: For accurate concurrency testing, please install ApacheBench (ab)."'''
