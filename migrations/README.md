# Database Migrations

This directory contains SQL migration files organized by module.

## Migration Naming Convention

Migrations follow the pattern: `YYYYMMDDHHMMSS_module_description.sql`

Example: `20231101120000_auth_create_sessions_table.sql`

## Migration Structure

Each migration file should contain both UP and DOWN migrations:

```sql
-- +migrate Up
CREATE TABLE IF NOT EXISTS users (...);

-- +migrate Down
DROP TABLE IF EXISTS users;
```

## Applying Migrations

Migrations are automatically applied on application startup by the migration runner in `internal/platform/database/migrations.go`.

## Module Organization

Migrations are logically organized by module:
- `auth/`: Authentication and session tables
- `user/`: User management tables
- `post/`: Post and category tables
- `comment/`: Comment tables
- `reaction/`: Reaction (likes/dislikes) tables
- `moderation/`: Moderation and report tables
- `notification/`: Notification tables

All migrations are applied in order based on their version number (timestamp).
