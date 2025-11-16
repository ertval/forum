#!/bin/bash

# Master Test Runner - Runs both API and Page tests in sequence
# Usage: ./scripts/run_all_tests.sh [-v|--verbose]

set -e

VERBOSE_FLAG=""
if [ "$1" = "-v" ] || [ "$1" = "--verbose" ]; then
    VERBOSE_FLAG="-v"
fi

echo "========================================="
echo "Forum Complete Test Suite"
echo "Running API and Page Tests"
echo "========================================="
echo ""

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Run API tests
echo "Step 1/2: Running API Tests..."
echo "========================================="
if "$SCRIPT_DIR/test_api.sh" $VERBOSE_FLAG; then
    API_RESULT="✓ PASSED"
    API_EXIT=0
else
    API_RESULT="✗ FAILED"
    API_EXIT=1
fi

echo ""
echo "API Tests: $API_RESULT"
echo ""
echo "========================================="
echo ""

# Small delay between test runs
sleep 2

# Run Page tests
echo "Step 2/2: Running Page Tests..."
echo "========================================="
if "$SCRIPT_DIR/test_pages.sh" $VERBOSE_FLAG; then
    PAGE_RESULT="✓ PASSED"
    PAGE_EXIT=0
else
    PAGE_RESULT="✗ FAILED"
    PAGE_EXIT=1
fi

echo ""
echo "Page Tests: $PAGE_RESULT"
echo ""

# Final summary
echo "========================================="
echo "FINAL SUMMARY"
echo "========================================="
echo "API Tests:  $API_RESULT"
echo "Page Tests: $PAGE_RESULT"
echo "========================================="

# Exit with failure if either failed
if [ $API_EXIT -ne 0 ] || [ $PAGE_EXIT -ne 0 ]; then
    echo ""
    echo "Some tests failed. Please review the output above."
    exit 1
else
    echo ""
    echo "✓ All tests passed successfully!"
    exit 0
fi
