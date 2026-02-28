-- Migration: Add avatar_path to users table
-- Module: user
-- Version: 008

-- +migrate Up
ALTER TABLE users ADD COLUMN avatar_path TEXT;

-- +migrate Down
-- SQLite does not support DROP COLUMN directly.
-- No-op downgrade for this additive migration.
SELECT 1;
