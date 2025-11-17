-- Migration: Create users table
-- Module: user
-- Version: 002

-- +migrate Up
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    public_id TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user',
    oauth_provider TEXT,
    oauth_provider_id TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    is_active INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_users_public_id ON users(public_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_oauth ON users(oauth_provider, oauth_provider_id);

-- +migrate Down
DROP INDEX IF EXISTS idx_users_oauth;
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_public_id;
DROP TABLE IF EXISTS users;
