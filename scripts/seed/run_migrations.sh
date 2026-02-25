#!/bin/bash
# =============================================================================
# RUN MIGRATIONS SCRIPT
# Applies all pending SQL migrations from the migrations directory
# Usage: ./run_migrations.sh
# =============================================================================

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
MIGRATIONS_DIR="${PROJECT_ROOT}/migrations"
DB_PATH="${PROJECT_ROOT}/data/forum.db"

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

echo -e "${BLUE}════════════════════════════════════════${NC}"
echo -e "${BLUE}  Forum Database Migrations Runner${NC}"
echo -e "${BLUE}════════════════════════════════════════${NC}"
echo ""

# Check if database directory exists
DB_DIR=$(dirname "$DB_PATH")
if [ ! -d "$DB_DIR" ]; then
    echo -e "${YELLOW}Creating database directory: $DB_DIR${NC}"
    mkdir -p "$DB_DIR"
fi

# Check if migrations directory exists
if [ ! -d "$MIGRATIONS_DIR" ]; then
    echo -e "${RED}Error: Migrations directory not found: $MIGRATIONS_DIR${NC}"
    exit 1
fi

# Create schema_migrations table if it doesn't exist
echo -e "${BLUE}Initializing migrations tracking table...${NC}"
sqlite3 "$DB_PATH" <<EOF
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    applied_at DATETIME NOT NULL DEFAULT (datetime('now'))
);
EOF
echo -e "${GREEN}✓${NC} Migrations table ready"
echo ""

# Get list of all migration files
MIGRATION_FILES=()
while IFS= read -r file; do
    MIGRATION_FILES+=("$file")
done < <(find "$MIGRATIONS_DIR" -maxdepth 1 -name "[0-9][0-9][0-9]_*.sql" -type f | sort)

if [ ${#MIGRATION_FILES[@]} -eq 0 ]; then
    echo -e "${YELLOW}No migration files found in $MIGRATIONS_DIR${NC}"
    exit 0
fi

echo -e "${BLUE}Found ${#MIGRATION_FILES[@]} migration file(s)${NC}"
echo ""

# Apply each migration
APPLIED=0
SKIPPED=0

for migration_file in "${MIGRATION_FILES[@]}"; do
    filename=$(basename "$migration_file")
    
    # Extract version number (first 3 digits)
    version=$(echo "$filename" | grep -oE '^[0-9]{3}' | sed 's/^0*//')
    
    # Check if migration already applied
    count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM schema_migrations WHERE version = $version;")
    
    if [ "$count" -gt 0 ]; then
        echo -e "${YELLOW}⊙${NC} Skipping $filename (already applied)"
        SKIPPED=$((SKIPPED + 1))
        continue
    fi
    
    echo -e "${BLUE}→${NC} Applying $filename..."
    
    # Extract the "Up" section from migration file
    sql=$(sed -n '/-- +migrate Up/,/-- +migrate Down/p' "$migration_file" | sed '1d;$d')
    
    if [ -z "$sql" ]; then
        echo -e "${YELLOW}  Warning: No SQL found in Up section${NC}"
        continue
    fi
    
    # Apply migration in a transaction
    if sqlite3 "$DB_PATH" <<EOF
BEGIN TRANSACTION;
$sql
INSERT INTO schema_migrations (version, name) VALUES ($version, '$filename');
COMMIT;
EOF
    then
        echo -e "${GREEN}✓${NC} Applied $filename"
        APPLIED=$((APPLIED + 1))
    else
        echo -e "${RED}✗${NC} Failed to apply $filename"
        echo -e "${RED}Migration failed! Database may be in an inconsistent state.${NC}"
        exit 1
    fi
done

echo ""
echo -e "${BLUE}════════════════════════════════════════${NC}"
echo -e "${GREEN}Migrations complete!${NC}"
echo -e "  Applied: ${GREEN}$APPLIED${NC}"
echo -e "  Skipped: ${YELLOW}$SKIPPED${NC}"
echo -e "  Total:   $((APPLIED + SKIPPED))"
echo -e "${BLUE}════════════════════════════════════════${NC}"
