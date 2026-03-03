-- Migration: Create notifications table
-- Module: notification
-- Version: 007

-- +migrate Up
CREATE TABLE IF NOT EXISTS notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    public_id TEXT UNIQUE NOT NULL,
    user_id INTEGER NOT NULL,
    actor_id INTEGER NOT NULL,
    target_id INTEGER NOT NULL,
    type TEXT NOT NULL,
    message TEXT NOT NULL,
    read BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (actor_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_notifications_public_id ON notifications(public_id);
CREATE INDEX idx_notifications_user ON notifications(user_id, read);
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);
-- Notifications: actor_id is used in JOIN (LEFT JOIN users u ON n.actor_id = u.id)
-- but has no index, causing full table scans as data grows.
CREATE INDEX IF NOT EXISTS idx_notifications_actor_id ON notifications(actor_id);

-- +migrate Down
DROP INDEX IF EXISTS idx_notifications_created_at;
DROP INDEX IF EXISTS idx_notifications_user;
DROP INDEX IF EXISTS idx_notifications_public_id;
DROP TABLE IF EXISTS notifications;
DROP INDEX IF EXISTS idx_notifications_actor_id;
