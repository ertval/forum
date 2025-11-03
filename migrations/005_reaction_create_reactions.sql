-- Migration: Create reactions table
-- Module: reaction
-- Version: 005

-- +migrate Up
CREATE TABLE IF NOT EXISTS reactions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    target_id TEXT NOT NULL,
    target_type TEXT NOT NULL,
    type TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    UNIQUE(user_id, target_id, target_type),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_reactions_target ON reactions(target_id, target_type);
CREATE INDEX idx_reactions_user ON reactions(user_id);

-- +migrate Down
DROP INDEX IF EXISTS idx_reactions_user;
DROP INDEX IF EXISTS idx_reactions_target;
DROP TABLE IF EXISTS reactions;
