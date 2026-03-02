-- Migration: Add missing indexes for query optimization
-- Module: cross-cutting
-- Version: 008
-- Finding 6: Add targeted indexes for columns used in queries without proper indexing

-- +migrate Up

-- Notifications: actor_id is used in JOIN (LEFT JOIN users u ON n.actor_id = u.id)
-- but has no index, causing full table scans as data grows.
CREATE INDEX IF NOT EXISTS idx_notifications_actor_id ON notifications(actor_id);

-- Categories: name is queried case-insensitively (WHERE LOWER(name) = LOWER(?))
-- The existing UNIQUE constraint index doesn't optimize case-insensitive lookups.
CREATE INDEX IF NOT EXISTS idx_categories_name_nocase ON categories(name COLLATE NOCASE);

-- +migrate Down
DROP INDEX IF EXISTS idx_categories_name_nocase;
DROP INDEX IF EXISTS idx_notifications_actor_id;
