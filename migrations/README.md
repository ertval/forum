# Database Migrations

This directory contains SQL migration files organized logically by module context.

## Migration Naming Convention

Migrations MUST follow the pattern: `NNN_module.sql`

Example: `001_auth.sql`

All migrations are applied in sequential order based on their numeric prefix (`001`, `002`, `003`, etc). Keep filenames short, containing just the numeric prefix and the module name.

Core rule for this repository: keep one core schema migration per module in this directory (plus cross-cutting schema files such as indexes when needed). Do not place schema migrations under `scripts/` or module folders.

## Migration Structure

Each migration must contain an `Up` and `Down` section with specific markers. The migrator extracts the Up section when applying, and the Down section when rolling back.

```sql
-- +migrate Up
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT
);

-- +migrate Down
DROP TABLE IF EXISTS users;
```

## Running & Applying Migrations

- **Automatic**: Migrations are automatically applied on application startup by the migration runner in `internal/platform/database/migrator.go`.
- **Manual**: Use the Makefile command: `make migrate`
- **Script**: Or run directly: `bash scripts/seed/run_migrations.sh`

The migrator creates a `schema_migrations` table to track applied migrations. Already-applied migrations are automatically skipped.

## Module Organization

Migrations map directly to the Bounded Contexts:
- `auth/`: Authentication and session tables
- `user/`: User management tables
- `post/`: Post and category tables
- `comment/`: Comment tables
- `reaction/`: Reaction (likes/dislikes) tables
- `moderation/`: Moderation and report tables
- `notification/`: Notification tables

---

## Practical Rules & Guidelines for SQLite

### Core Principles
- **Immutability**: Always add a new migration file for every schema change. **Do not edit migrations that have already been applied** in any environment.
- **Testing**: Test migrations on a copy of the database before applying to staging/production.
- **Safety First**: Back up the DB file before running migrations in production manually: `cp data/forum.db data/forum.db.bak`.

### SQLite Specific Operations
- **ADD COLUMN**: `ALTER TABLE <table> ADD COLUMN <col> <type>` is fully supported and the recommended approach for adding columns safely. It is fast and does not require a table rebuild.
- **DROP / RENAME / CHANGE TYPE**: Usually requires a full table rebuild. 
  - Pattern: create a new table with desired schema, copy data over, drop old table, rename the new table securely, then **recreate indexes/triggers**.
- **Indexes and Triggers**: Must be explicitly recreated when rebuilding tables.

### Idempotence and "already exists" errors
SQLite will error with `duplicate column name` if you try to `ADD COLUMN` when the column already exists. Options to handle this safely:
1. Wrap checks using `PRAGMA table_info` first via a tiny wrapper script (plain SQL conditionals in SQLite are limited).
2. Let the migrator detect the error, fail, and record the migration as applied manually after manual verification.
3. Always run migrations on a copy first and mark the migration as applied manually if you verified it beforehand.

### Backfill and Data Migrations
- You can perform data backfills in the same migration transaction after adding a column (e.g., copy data from an old column or compute values).
- For very large tables, consider batching or running a separate background job to avoid long locks on the DB file.

### Rollback Caveats
- `Down` migrations for schema changes that require a table rebuild are potentially destructive and should always be tested on a copy.
- In production runtime environments, prefer **forward-only** migrations (by writing compensating `Up` migrations) for maximum safety instead of rolling back.

## Manually Marking a Migration Applied

If you hand-inspected the SQLite DB and the schema change is already present (perhaps from manual tweaking), record it as applied so the runner skips it:

```sql
INSERT INTO schema_migrations (version, name, applied_at) VALUES (2, '002_user.sql', datetime('now'));
```
