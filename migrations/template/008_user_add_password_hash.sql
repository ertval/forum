-- +migrate Up
BEGIN TRANSACTION;
-- Add password_hash column (safe: ADD COLUMN is supported by SQLite)
ALTER TABLE users ADD COLUMN password_hash TEXT;

-- If you previously used a `password` column, copy those values into the new column
-- (this will leave the old `password` column in place; removal requires a table rebuild)
UPDATE users SET password_hash = password WHERE password_hash IS NULL AND EXISTS (SELECT 1 FROM pragma_table_info('users') WHERE name='password');

COMMIT;

-- +migrate Down
BEGIN TRANSACTION;
-- Rollback: recreate `users` table without `password_hash` and with `password` column
-- Note: Dropping a column in SQLite requires table rebuild; this Down attempts to recreate the prior schema.
CREATE TABLE users_tmp (
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

INSERT INTO users_tmp (id, email, username, password, role, oauth_provider, oauth_provider_id, created_at, updated_at, is_active)
SELECT id, email, username, COALESCE(password_hash, '') as password, role, oauth_provider, oauth_provider_id, created_at, updated_at, is_active FROM users;

DROP TABLE users;
ALTER TABLE users_tmp RENAME TO users;
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

COMMIT;
