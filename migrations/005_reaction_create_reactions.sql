-- Migration: Create reactions table
-- Module: reaction
-- Version: 005

-- +migrate Up
CREATE TABLE IF NOT EXISTS reactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    public_id TEXT UNIQUE NOT NULL,
    user_id INTEGER NOT NULL,
    target_id INTEGER NOT NULL,
    target_type TEXT NOT NULL,
    type TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    UNIQUE(user_id, target_id, target_type),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_reactions_public_id ON reactions(public_id);
CREATE INDEX idx_reactions_target ON reactions(target_id, target_type);
CREATE INDEX idx_reactions_user ON reactions(user_id);

-- +migrate Down
DROP INDEX IF EXISTS idx_reactions_user;
DROP INDEX IF EXISTS idx_reactions_target;
DROP INDEX IF EXISTS idx_reactions_public_id;
DROP TABLE IF EXISTS reactions;
