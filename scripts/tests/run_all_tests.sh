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
echo "Running Unit, Integration, API and Page Tests"
echo "========================================="
echo ""

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Run Unit Tests (including user stats tests)
echo "Step 1/4: Running Unit Tests..."
echo "========================================="
cd "$PROJECT_ROOT"
if go test $VERBOSE_FLAG ./tests/integration/ -run "UserStats|UserCard"; then
    UNIT_RESULT="✓ PASSED"
    UNIT_EXIT=0
else
    UNIT_RESULT="✗ FAILED"
    UNIT_EXIT=1
fi

echo ""
echo "Unit Tests: $UNIT_RESULT"
echo ""
echo "========================================="
echo ""

# Small delay between test runs
sleep 1

# Run API tests
echo "Step 2/4: Running API Tests..."
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

# Run Integration tests
echo "Step 3/4: Running Integration Tests..."
echo "========================================="
cd "$PROJECT_ROOT"
if go test $VERBOSE_FLAG ./tests/integration/; then
    INTEGRATION_RESULT="✓ PASSED"
    INTEGRATION_EXIT=0
else
    INTEGRATION_RESULT="✗ FAILED"
    INTEGRATION_EXIT=1
fi

echo ""
echo "Integration Tests: $INTEGRATION_RESULT"
echo ""
echo "========================================="
echo ""

# Small delay between test runs
sleep 2

# Run Page tests
echo "Step 4/4: Running Page Tests..."
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
echo "Unit Tests:        $UNIT_RESULT"
echo "Integration Tests: $INTEGRATION_RESULT"
echo "API Tests:         $API_RESULT"
echo "Page Tests:        $PAGE_RESULT"
echo "========================================="

# Exit with failure if any failed
if [ $UNIT_EXIT -ne 0 ] || [ $INTEGRATION_EXIT -ne 0 ] || [ $API_EXIT -ne 0 ] || [ $PAGE_EXIT -ne 0 ]; then
    echo ""
    echo "Some tests failed. Please review the output above."
    exit 1
else
    echo ""
    echo "✓ All tests passed successfully!"
    exit 0
fi
