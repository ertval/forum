#!/bin/bash

# Seed script to add default categories
# Run this after the server is running

BASE_URL="http://localhost:8080"

echo "Adding default categories to database..."

# We need to insert directly into the database since there's no public API for creating categories
# This would normally be done via migrations or admin interface

sqlite3 data/forum.db << EOF
INSERT OR IGNORE INTO categories (id, name, description, created_at) VALUES 
('general-id', 'general', 'General discussions', datetime('now')),
('tech-id', 'tech', 'Technology topics', datetime('now')),
('news-id', 'news', 'News and current events', datetime('now')),
('gaming-id', 'gaming', 'Gaming discussions', datetime('now')),
('music-id', 'music', 'Music and entertainment', datetime('now'));
EOF

echo "Categories added successfully!"
echo ""
echo "Listing categories:"
sqlite3 data/forum.db "SELECT * FROM categories;"
