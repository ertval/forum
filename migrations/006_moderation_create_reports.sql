-- Migration: Create reports table
-- Module: moderation
-- Version: 006

-- +migrate Up
CREATE TABLE IF NOT EXISTS reports (
    id TEXT PRIMARY KEY,
    reporter_id TEXT NOT NULL,
    moderator_id TEXT,
    target_id TEXT NOT NULL,
    target_type TEXT NOT NULL,
    reason TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    response TEXT,
    created_at DATETIME NOT NULL,
    reviewed_at DATETIME,
    FOREIGN KEY (reporter_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (moderator_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX idx_reports_status ON reports(status);
CREATE INDEX idx_reports_target ON reports(target_id, target_type);
CREATE INDEX idx_reports_moderator ON reports(moderator_id);

-- +migrate Down
DROP INDEX IF EXISTS idx_reports_moderator;
DROP INDEX IF EXISTS idx_reports_target;
DROP INDEX IF EXISTS idx_reports_status;
DROP TABLE IF EXISTS reports;
