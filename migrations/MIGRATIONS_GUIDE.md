Migrations Guide — Practical rules and templates

This file is a practical companion for creating and running migrations in this repository (SQLite-specific guidance).

Summary
- Always add a new migration file for every schema change. Do not edit migrations that have already been applied in any environment.
- Test migrations on a copy of the database before applying to staging/production.
- Back up the DB file before running migrations in production: `cp data/forum.db data/forum.db.bak`.

Filename and ordering
- Use a numeric prefix that increases with time to order migrations (this repo uses `NNN_description.sql` for clarity, e.g. `008_user_add_password_hash.sql`).
- Keep filenames short and descriptive.

Migration structure
- Each migration must contain an Up and Down section with the markers below. The migrator extracts the Up section and runs it.

  -- +migrate Up
  -- SQL to apply the change

  -- +migrate Down
  -- SQL to rollback (if feasible)

General rules for SQLite
- ADD COLUMN: `ALTER TABLE <table> ADD COLUMN <col> <type>` is supported and the recommended approach for adding columns. It is fast and does not require table rebuild.
- DROP / RENAME / CHANGE TYPE: Usually requires a table rebuild. Pattern: create new table with desired schema, copy data, drop old table, rename new table, recreate indexes/triggers.
- Indexes and triggers must be recreated when rebuilding tables.

Idempotence and "already exists" errors
- SQLite will error with `duplicate column name` if you try to `ADD COLUMN` when the column already exists. Options to handle this:
  1. Make the migration query check PRAGMA table_info first (possible with a tiny wrapper script). Plain SQL-only conditionals are limited.
  2. Let the migrator detect the error and record the migration as applied after manual verification (this repo includes a helper runner script that can do that).
  3. Always run migrations on a copy first and mark the migration as applied manually if you verified it.

Backfill and data migrations
- You can perform backfills in the same migration transaction after adding a column (e.g., copy data from an old column or compute values).
- For very large tables, consider batching or running a separate background job to avoid long locks.

Rollback caveats
- Down migrations for schema changes that require a table rebuild are potentially destructive and should be tested on a copy.
- In production, prefer forward-only migrations and compensating migrations for safety.

Manually mark a migration applied
- If you inspected the DB and the schema change is already present, record it as applied:

  INSERT INTO schema_migrations (version, name, applied_at) VALUES (8, '008_user_add_password_hash.sql', datetime('now'));

Examples and templates
- See `000_template_migration.sql` in this folder for a ready-to-copy template with examples for both ADD COLUMN and table-rebuild rename/type-change.


