#!/bin/bash
# =============================================================================
# Database Seed Script
# Seeds the database with test data required for all test scripts
# =============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Database path
DB_PATH="${PROJECT_ROOT}/data/forum.db"
SEED_FILE="${SCRIPT_DIR}/seed_data.sql"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}=== Forum Database Seeder ===${NC}"
echo ""

# Check if database file exists
if [ ! -f "$DB_PATH" ]; then
    echo -e "${RED}Database file does not exist: $DB_PATH${NC}"
    echo "Please run the application first to create the database and apply migrations."
    exit 1
fi

# Check if seed file exists
if [ ! -f "$SEED_FILE" ]; then
    echo -e "${RED}Seed file not found: $SEED_FILE${NC}"
    exit 1
fi

# Check required tables exist
REQUIRED_TABLES=("users" "categories" "posts" "post_categories" "reactions" "comments" "sessions" "reports" "notifications")
echo "Checking database tables..."

for table in "${REQUIRED_TABLES[@]}"; do
    count=$(sqlite3 "$DB_PATH" "SELECT COUNT(name) FROM sqlite_master WHERE type='table' AND name='$table';")
    if [ "$count" -eq 0 ]; then
        echo -e "${RED}Table '$table' does not exist. Please run migrations first.${NC}"
        exit 1
    fi
done
echo -e "${GREEN}✓ All required tables exist${NC}"
echo ""

# Execute the seed data
echo "Seeding database with test data..."
sqlite3 "$DB_PATH" < "$SEED_FILE"

if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✓ Database seeded successfully!${NC}"
    echo ""
    echo "Test data includes:"
    echo "  • 10 test users (including testuser@example.com)"
    echo "  • 9 categories"
    echo "  • 10 sample posts"
    echo "  • 5 sample comments"
    echo "  • 7 sample reactions"
    echo "  • 2 notifications"
    echo "  • 2 moderation reports"
    echo ""
    echo -e "${YELLOW}Test Credentials:${NC}"
    echo "  Primary:   testuser@example.com / password123"
    echo "  Secondary: testuser2@example.com / password123"
    echo ""
else
    echo -e "${RED}Failed to seed database.${NC}"
    exit 1
fi
