-- Migration: Create moderator requests table
-- Module: moderation
-- Version: 008

-- +migrate Up
CREATE TABLE IF NOT EXISTS moderator_requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    public_id TEXT UNIQUE NOT NULL,
    requester_id INTEGER NOT NULL,
    reviewer_id INTEGER,
    status TEXT NOT NULL DEFAULT 'pending',
    message TEXT,
    response TEXT,
    created_at DATETIME NOT NULL,
    reviewed_at DATETIME,
    FOREIGN KEY (requester_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (reviewer_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX idx_moderator_requests_public_id ON moderator_requests(public_id);
CREATE INDEX idx_moderator_requests_requester ON moderator_requests(requester_id);
CREATE INDEX idx_moderator_requests_status ON moderator_requests(status);

-- +migrate Down
DROP INDEX IF EXISTS idx_moderator_requests_status;
DROP INDEX IF EXISTS idx_moderator_requests_requester;
DROP INDEX IF EXISTS idx_moderator_requests_public_id;
DROP TABLE IF EXISTS moderator_requests;