-- Template migration: replace NNN and description
-- Use a numeric prefix for ordering: NNN_description.sql e.g. 009_add_column_example.sql

-- +migrate Up
-- Example 1: Add a new column (safe, fast)
BEGIN TRANSACTION;
ALTER TABLE users ADD COLUMN new_column TEXT;
-- Backfill example: copy from an existing column if present
UPDATE users
SET new_column = old_column
WHERE new_column IS NULL
  AND EXISTS (SELECT 1 FROM pragma_table_info('users') WHERE name='old_column');
COMMIT;

-- +migrate Down
-- Rolling back ADD COLUMN requires table rebuild in SQLite. The following is a template for rebuilding the table
BEGIN TRANSACTION;
PRAGMA foreign_keys=off;

CREATE TABLE users_tmp (
    id INTEGER PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    username TEXT UNIQUE NOT NULL,
    -- omit new_column here to remove it in rollback
    password_hash TEXT,
    role TEXT NOT NULL DEFAULT 'user',
    oauth_provider TEXT,
    oauth_provider_id TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    is_active INTEGER NOT NULL DEFAULT 1
);

INSERT INTO users_tmp (id, email, username, password_hash, role, oauth_provider, oauth_provider_id, created_at, updated_at, is_active)
SELECT id, email, username, password_hash, role, oauth_provider, oauth_provider_id, created_at, updated_at, is_active FROM users;

DROP TABLE users;
ALTER TABLE users_tmp RENAME TO users;

-- recreate indexes/triggers
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

PRAGMA foreign_keys=on;
COMMIT;

-- ---------------------------
-- Example 2: Rename/change type (table rebuild)
-- +migrate Up
-- This demonstrates renaming `password` -> `password_hash` and changing column order/attributes
BEGIN TRANSACTION;
PRAGMA foreign_keys=off;

CREATE TABLE users_new (
    id INTEGER PRIMARY KEY,
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

INSERT INTO users_new (id, email, username, password_hash, role, oauth_provider, oauth_provider_id, created_at, updated_at, is_active)
SELECT id, email, username, password AS password_hash, role, oauth_provider, oauth_provider_id, created_at, updated_at, is_active FROM users;

DROP TABLE users;
ALTER TABLE users_new RENAME TO users;

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
PRAGMA foreign_keys=on;
COMMIT;

-- +migrate Down
-- To rollback the rename/type-change, reverse the process: create old table shape and copy data back.
BEGIN TRANSACTION;
PRAGMA foreign_keys=off;

CREATE TABLE users_old (
    id INTEGER PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user',
    oauth_provider TEXT,
    oauth_provider_id TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    is_active INTEGER NOT NULL DEFAULT 1
);

INSERT INTO users_old (id, email, username, password, role, oauth_provider, oauth_provider_id, created_at, updated_at, is_active)
SELECT id, email, username, password_hash AS password, role, oauth_provider, oauth_provider_id, created_at, updated_at, is_active FROM users;

DROP TABLE users;
ALTER TABLE users_old RENAME TO users;
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
PRAGMA foreign_keys=on;
COMMIT;
