#!/bin/bash
# =============================================================================
# RUN ALL TESTS SCRIPT
# Discovers and runs all test scripts in the scripts/tests directory
# Usage: ./run_all_tests.sh [--quiet|-q]
#   --quiet/-q: Show only summary, hide individual test output
# =============================================================================

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Parse arguments
QUIET_MODE=false
for arg in "$@"; do
    case $arg in
        --quiet|-q)
            QUIET_MODE=true
            shift
            ;;
    esac
done

# Colors
if [ -t 1 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    NC='\033[0m'
else
    RED='' GREEN='' YELLOW='' BLUE='' NC=''
fi

# Track results
declare -A RESULTS
TOTAL_PASSED=0
TOTAL_FAILED=0

echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║             FORUM TEST SUITE - ALL TESTS                   ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Get script name to exclude itself
THIS_SCRIPT=$(basename "${BASH_SOURCE[0]}")

# Find all test scripts and sort them alphabetically by filename
if [ "$QUIET_MODE" = false ]; then
    echo -e "${YELLOW}Discovering test scripts...${NC}"
fi
TEST_SCRIPTS=()
while IFS= read -r script; do
    TEST_SCRIPTS+=("$script")
    script_name=$(basename "$script")
    if [ "$QUIET_MODE" = false ]; then
        echo -e "  ${GREEN}✓${NC} Found: $script_name"
    fi
done < <(find "$SCRIPT_DIR" -maxdepth 1 -name "test_*.sh" -type f -exec basename {} \; | sort | while read name; do echo "$SCRIPT_DIR/$name"; done)

if [ "$QUIET_MODE" = false ]; then
    echo ""
    echo -e "${YELLOW}Found ${#TEST_SCRIPTS[@]} test script(s)${NC}"
    echo ""
fi

if [ ${#TEST_SCRIPTS[@]} -eq 0 ]; then
    echo -e "${RED}No test scripts found!${NC}"
    exit 1
fi

# Verify database exists and has data (DO NOT seed - this is a test runner)
if [ "$QUIET_MODE" = false ]; then
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}STEP 1: Verifying database${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo ""
fi

DB_PATH="${PROJECT_ROOT}/data/forum.db"

# Check database file exists
if [ ! -f "$DB_PATH" ]; then
    echo -e "${RED}✗ Database file does not exist: $DB_PATH${NC}"
    echo "Please run the application first to create the database and apply migrations."
    echo "Then run scripts/seed/seed.sh to populate test data."
    exit 1
fi
if [ "$QUIET_MODE" = false ]; then
    echo -e "${GREEN}✓${NC} Database file exists"
fi

# Required tables
REQUIRED_TABLES=("users" "categories" "posts" "post_categories" "reactions" "comments" "sessions")
if [ "$QUIET_MODE" = false ]; then
    echo "Checking database tables..."
fi

for table in "${REQUIRED_TABLES[@]}"; do
    count=$(sqlite3 "$DB_PATH" "SELECT COUNT(name) FROM sqlite_master WHERE type='table' AND name='$table';")
    if [ "$count" -eq 0 ]; then
        echo -e "${RED}✗ Table '$table' does not exist. Please run migrations first.${NC}"
        exit 1
    fi
done
if [ "$QUIET_MODE" = false ]; then
    echo -e "${GREEN}✓${NC} All required tables exist"
fi

# Verify each table has data
if [ "$QUIET_MODE" = false ]; then
    echo "Verifying test data..."
fi
MISSING_DATA=0

USER_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM users;")
if [ "$USER_COUNT" -eq 0 ]; then
    echo -e "${RED}✗ No users in database${NC}"
    MISSING_DATA=1
elif [ "$QUIET_MODE" = false ]; then
    echo -e "${GREEN}✓${NC} Users: $USER_COUNT"
fi

CATEGORY_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM categories;")
if [ "$CATEGORY_COUNT" -eq 0 ]; then
    echo -e "${RED}✗ No categories in database${NC}"
    MISSING_DATA=1
elif [ "$QUIET_MODE" = false ]; then
    echo -e "${GREEN}✓${NC} Categories: $CATEGORY_COUNT"
fi

POST_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM posts;")
if [ "$POST_COUNT" -eq 0 ]; then
    echo -e "${RED}✗ No posts in database${NC}"
    MISSING_DATA=1
elif [ "$QUIET_MODE" = false ]; then
    echo -e "${GREEN}✓${NC} Posts: $POST_COUNT"
fi

COMMENT_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM comments;")
if [ "$COMMENT_COUNT" -eq 0 ]; then
    if [ "$QUIET_MODE" = false ]; then
        echo -e "${YELLOW}⚠${NC} No comments in database (optional)"
    fi
elif [ "$QUIET_MODE" = false ]; then
    echo -e "${GREEN}✓${NC} Comments: $COMMENT_COUNT"
fi

REACTION_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM reactions;")
if [ "$REACTION_COUNT" -eq 0 ]; then
    if [ "$QUIET_MODE" = false ]; then
        echo -e "${YELLOW}⚠${NC} No reactions in database (optional)"
    fi
elif [ "$QUIET_MODE" = false ]; then
    echo -e "${GREEN}✓${NC} Reactions: $REACTION_COUNT"
fi

if [ $MISSING_DATA -eq 1 ]; then
    echo ""
    echo -e "${RED}✗ Required test data missing!${NC}"
    echo "Please run: scripts/seed/seed.sh"
    exit 1
fi

if [ "$QUIET_MODE" = false ]; then
    echo ""
    echo -e "${GREEN}✓ Database verification passed${NC}"
    echo ""
fi

# Run each test script
SCRIPT_NUM=2
for script in "${TEST_SCRIPTS[@]}"; do
    script_name=$(basename "$script")
    
    if [ "$QUIET_MODE" = false ]; then
        echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
        echo -e "${BLUE}STEP $SCRIPT_NUM: Running $script_name${NC}"
        echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
        echo ""
    fi
    
    # Make sure script is executable
    chmod +x "$script"
    
    # Run the script (suppress output in quiet mode)
    if [ "$QUIET_MODE" = true ]; then
        if bash "$script" > /dev/null 2>&1; then
            RESULTS["$script_name"]="PASS"
            TOTAL_PASSED=$((TOTAL_PASSED + 1))
        else
            RESULTS["$script_name"]="FAIL"
            TOTAL_FAILED=$((TOTAL_FAILED + 1))
        fi
    else
        if bash "$script"; then
            RESULTS["$script_name"]="PASS"
            TOTAL_PASSED=$((TOTAL_PASSED + 1))
            echo ""
            echo -e "${GREEN}✓ $script_name PASSED${NC}"
        else
            RESULTS["$script_name"]="FAIL"
            TOTAL_FAILED=$((TOTAL_FAILED + 1))
            echo ""
            echo -e "${RED}✗ $script_name FAILED${NC}"
        fi
        echo ""
    fi
    
    SCRIPT_NUM=$((SCRIPT_NUM + 1))
done

# Final summary
echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                    FINAL SUMMARY                           ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Display results in alphabetical order
for script_name in $(echo "${!RESULTS[@]}" | tr ' ' '\n' | sort); do
    result="${RESULTS[$script_name]}"
    if [ "$result" = "PASS" ]; then
        echo -e "  ${GREEN}✓${NC} $script_name - ${GREEN}PASSED${NC}"
    else
        echo -e "  ${RED}✗${NC} $script_name - ${RED}FAILED${NC}"
    fi
done

echo ""
echo -e "${YELLOW}───────────────────────────────────────────────────────────────${NC}"
echo -e "  Scripts Passed: ${GREEN}$TOTAL_PASSED${NC}"
echo -e "  Scripts Failed: ${RED}$TOTAL_FAILED${NC}"
echo -e "  Total Scripts:  $((TOTAL_PASSED + TOTAL_FAILED))"
echo -e "${YELLOW}───────────────────────────────────────────────────────────────${NC}"
echo ""

if [ $TOTAL_FAILED -eq 0 ]; then
    echo -e "${GREEN}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║              ALL TESTS PASSED SUCCESSFULLY!                ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════════╝${NC}"
    exit 0
else
    echo -e "${RED}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${RED}║              SOME TESTS FAILED - SEE ABOVE                 ║${NC}"
    echo -e "${RED}╚════════════════════════════════════════════════════════════╝${NC}"
    exit 1
fi
