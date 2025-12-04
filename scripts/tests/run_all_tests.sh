#!/bin/bash

# Master Test Runner - Runs all test scripts in this directory
# Usage: ./scripts/tests/run_all_tests.sh [-v|--verbose]

set -e

VERBOSE_FLAG=""
if [ "$1" = "-v" ] || [ "$1" = "--verbose" ]; then
    VERBOSE_FLAG="-v"
fi

echo "========================================="
echo "Forum Complete Test Suite"
echo "Running All Tests: Unit, Integration, and E2E Scripts"
echo "========================================="
echo ""

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Initialize results tracking
declare -A TEST_RESULTS
declare -A TEST_EXITS
TOTAL_TESTS=0
FAILED_TESTS=0

# Helper function to run a test and track results
run_test() {
    local test_name="$1"
    local test_command="$2"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo "Step $TOTAL_TESTS: Running $test_name..."
    echo "========================================="
    
    if eval "$test_command"; then
        TEST_RESULTS["$test_name"]="✓ PASSED"
        TEST_EXITS["$test_name"]=0
    else
        TEST_RESULTS["$test_name"]="✗ FAILED"
        TEST_EXITS["$test_name"]=1
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    
    echo ""
    echo "$test_name: ${TEST_RESULTS[$test_name]}"
    echo ""
    echo "========================================="
    echo ""
    
    # Small delay between test runs
    sleep 1
}

# Run Go Unit Tests
cd "$PROJECT_ROOT"
run_test "Go Unit Tests" "go test $VERBOSE_FLAG ./tests/unit/..."

# Run Go Integration Tests
run_test "Go Integration Tests" "go test $VERBOSE_FLAG ./tests/integration/..."

# Discover and run all bash test scripts in the scripts/tests directory
# Excludes: run_all_tests.sh (this script), README.md, and non-.sh files
for script in "$SCRIPT_DIR"/*.sh; do
    script_name=$(basename "$script")
    
    # Skip this script itself
    if [ "$script_name" = "run_all_tests.sh" ]; then
        continue
    fi
    
    # Skip if not executable
    if [ ! -x "$script" ]; then
        echo "Skipping $script_name (not executable)"
        continue
    fi
    
    run_test "E2E: $script_name" "\"$script\" $VERBOSE_FLAG"
done

# Final summary
echo "========================================="
echo "FINAL SUMMARY"
echo "========================================="
for test_name in "${!TEST_RESULTS[@]}"; do
    printf "%-30s %s\n" "$test_name:" "${TEST_RESULTS[$test_name]}"
done
echo "========================================="
echo "Total: $TOTAL_TESTS tests, $FAILED_TESTS failed"
echo "========================================="

# Exit with failure if any failed
if [ $FAILED_TESTS -ne 0 ]; then
    echo ""
    echo "Some tests failed. Please review the output above."
    exit 1
else
    echo ""
    echo "✓ All tests passed successfully!"
    exit 0
fi
