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

# TLS certificate files (for local dev)
CERT_DIR="${PROJECT_ROOT}/certs"
CERT_FILE="${CERT_DIR}/cert.pem"
KEY_FILE="${CERT_DIR}/key.pem"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}=== Forum Database Seeder ===${NC}"
echo ""

# Ensure TLS certs exist for local development (generate if missing)
echo "Checking TLS certificates in: $CERT_DIR"
if [ -f "$CERT_FILE" ] && [ -f "$KEY_FILE" ]; then
    echo -e "${GREEN}✓ TLS certificate and key found${NC}"
else
    echo -e "${YELLOW}TLS certificate or key not found. Generating...${NC}"
    # Call the bundled generator script. Use bash to avoid exec permission issues.
    if ! bash "$SCRIPT_DIR/generate_certs.sh" "$CERT_DIR"; then
        echo -e "${RED}Failed to generate TLS certificates.${NC}"
        exit 1
    fi

    # Verify generation succeeded
    if [ -f "$CERT_FILE" ] && [ -f "$KEY_FILE" ]; then
        echo -e "${GREEN}✓ TLS certificates generated: $CERT_FILE, $KEY_FILE${NC}"
    else
        echo -e "${RED}Certificate generation did not produce expected files.${NC}"
        exit 1
    fi
fi


# Check if database file exists - if not, run migrations first
if [ ! -f "$DB_PATH" ]; then
    echo -e "${YELLOW}Database file does not exist. Running migrations...${NC}"
    # Run migrations from project root, not from seed directory
    if ! bash "${PROJECT_ROOT}/scripts/seed/run_migrations.sh"; then
        echo -e "${RED}Failed to run migrations.${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ Migrations completed${NC}"
    echo ""
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
    # Fetch actual counts from the database
    users_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM users;")
    categories_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM categories;")
    posts_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM posts;")
    post_categories_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM post_categories;")
    comments_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM comments;")
    reactions_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM reactions;")
    notifications_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM notifications;")
    reports_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM reports;")

    echo "Test data includes:"
    echo "  • ${users_count:-0} test users (including testuser@example.com)"
    echo "  • ${categories_count:-0} categories"
    echo "  • ${posts_count:-0} sample posts"
    echo "  • ${post_categories_count:-0} post-categories links"
    echo "  • ${comments_count:-0} sample comments"
    echo "  • ${reactions_count:-0} sample reactions"
    echo "  • ${notifications_count:-0} notifications"
    echo "  • ${reports_count:-0} moderation reports"
    echo ""
    echo -e "${YELLOW}Test Credentials:${NC}"
    echo "  Primary:   testuser@example.com / password123"
    echo "  Secondary: testuser2@example.com / password123"
    echo ""
else
    echo -e "${RED}Failed to seed database.${NC}"
    exit 1
fi
