#!/bin/bash

# Database path (same as in config)
DB_PATH="./data/forum.db"

# Seed data file path
SEED_FILE="./scripts/seed/seed_data.sql"

# Check if database exists
if [ ! -f "$DB_PATH" ]; then
    echo "Database file does not exist: $DB_PATH. Please run the application first to create the database."
    exit 1
fi

# Check if tables exist
tables=("users" "categories" "posts" "post_categories" "reactions" "comments" "sessions" "reports" "notifications")
for table in "${tables[@]}"; do
    count=$(sqlite3 "$DB_PATH" "SELECT COUNT(name) FROM sqlite_master WHERE type='table' AND name='$table';")
    if [ "$count" -eq 0 ]; then
        echo "Table $table does not exist. Please run migrations first."
        exit 1
    fi
done

# Execute the seed data
echo "Seeding database with mock data..."
sqlite3 "$DB_PATH" < "$SEED_FILE"

if [ $? -eq 0 ]; then
    echo "✅ Database seeded successfully!"
    echo "Mock data includes:"
    echo "- 20 test users"
    echo "- 20 categories"
    echo "- 20 sample posts"
    echo "- 20 sample comments"
    echo "- 18 sample reactions (likes/dislikes)"
    echo "- 20 active sessions"
    echo "- 20 moderation reports"
    echo "- 20 notifications"
else
    echo "Failed to seed database."
    exit 1
fi