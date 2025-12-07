#!/bin/bash
# =============================================================================
# SIMPLE TEST RUNNER
# Discovers and runs all test scripts in scripts/tests/
# Usage: ./run_all_tests.sh [--quiet|-q]
# =============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
DB_PATH="${FORUM_DB_PATH:-$PROJECT_ROOT/data/forum.db}"
SEED_SCRIPT="$PROJECT_ROOT/scripts/seed/seed.sh"

# Parse quiet mode
QUIET=false
[[ "$*" =~ (--quiet|-q) ]] && QUIET=true

# Colors
RED='\033[0;31m' GREEN='\033[0;32m' YELLOW='\033[1;33m' BLUE='\033[0;34m' NC='\033[0m'

# Results tracking
declare -A RESULTS
PASSED=0 FAILED=0

ensure_db_ready() {
    if [ ! -f "$DB_PATH" ]; then
        echo -e "${YELLOW}Database not found at $DB_PATH${NC}"
        echo -e "Run ${GREEN}$SEED_SCRIPT${NC} to create the database and seed the required data."
        exit 1
    fi

    REQUIRED_TABLES=(users categories posts comments reactions notifications reports)
    for table in "${REQUIRED_TABLES[@]}"; do
        count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM $table;" 2>/dev/null)
        if [ $? -ne 0 ]; then
            echo -e "${YELLOW}Unable to query '$table'. Run ${GREEN}$SEED_SCRIPT${NC} before running the tests.${NC}"
            exit 1
        fi
        if [ -z "$count" ] || [ "$count" -eq 0 ]; then
            echo -e "${YELLOW}Table '$table' exists but has no rows.${NC}"
            echo -e "Run ${GREEN}$SEED_SCRIPT${NC} to ensure necessary data is available for the tests."
            exit 1
        fi
    done
}

ensure_db_ready

# Brief header removed to avoid duplicating the summary at the end

# Find all test scripts (exclude this script)
TEST_SCRIPTS=($(find "$SCRIPT_DIR" -maxdepth 1 -name "test_*.sh" -type f | sort))

if [ ${#TEST_SCRIPTS[@]} -eq 0 ]; then
    echo -e "${RED}✗ No test scripts found!${NC}"
    exit 1
fi

# Spinner function
spinner() {
    local pid=$1
    local script_name=$2
    local spin='⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏'
    local i=0
    while kill -0 $pid 2>/dev/null; do
        printf "\r${BLUE}${spin:$i:1}${NC} Running %s..." "$script_name"
        sleep 0.08
        i=$(( (i+1) % ${#spin} ))
    done
    printf "\r\033[K"  # Clear line
}

# Run each test
for script in "${TEST_SCRIPTS[@]}"; do
    script_name=$(basename "$script")
    chmod +x "$script"
    
    if [ "$QUIET" = true ]; then
        # Quiet mode: show spinner, capture output
        temp_out=$(mktemp)
        bash "$script" > "$temp_out" 2>&1 &
        pid=$!
        spinner $pid "$script_name"
        wait $pid
        exit_code=$?
        rm -f "$temp_out"
    else
        # Verbose mode: show full output
        echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
        echo -e "${BLUE}Running: $script_name${NC}"
        echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
        echo ""
        bash "$script"
        exit_code=$?
        echo ""
    fi
    
    # Track result (records only; skip per-test immediate printing to avoid duplicate summaries)
    if [ $exit_code -eq 0 ]; then
        RESULTS["$script_name"]="PASS"
        PASSED=$((PASSED + 1))
    else
        RESULTS["$script_name"]="FAIL"
        FAILED=$((FAILED + 1))
    fi
done

# Summary
echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}                    SUMMARY                                 ${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo ""

for script_name in $(printf '%s\n' "${!RESULTS[@]}" | sort); do
    if [ "${RESULTS[$script_name]}" = "PASS" ]; then
        echo -e "  ${GREEN}✓${NC} $script_name"
    else
        echo -e "  ${RED}✗${NC} $script_name"
    fi
done

echo ""
echo -e "${YELLOW}───────────────────────────────────────────────────────────${NC}"
echo -e "  Passed: ${GREEN}$PASSED${NC} | Failed: ${RED}$FAILED${NC} | Total: $((PASSED + FAILED))"
echo -e "${YELLOW}───────────────────────────────────────────────────────────${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ ALL TESTS PASSED!${NC}"
    echo ""
    exit 0
else
    echo -e "${RED}✗ SOME TESTS FAILED${NC}"
    echo ""
    exit 1
fi
