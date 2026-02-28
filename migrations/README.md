# Database Migrations

This directory contains SQL migration files organized by module.

## Migration Naming Convention

Migrations follow the pattern: `NNN_module_description.sql`

Example: `001_auth_create_sessions.sql`

## Migration Structure

Each migration file should contain both UP and DOWN migrations:

```sql
-- +migrate Up
CREATE TABLE IF NOT EXISTS users (...);

-- +migrate Down
DROP TABLE IF EXISTS users;
```

## Applying Migrations

Migrations are automatically applied on application startup by the migration runner in `internal/platform/database/migrator.go`.

## Module Organization

Migrations are logically organized by module:
- `auth/`: Authentication and session tables
- `user/`: User management tables
- `post/`: Post and category tables
- `comment/`: Comment tables
- `reaction/`: Reaction (likes/dislikes) tables
- `moderation/`: Moderation and report tables
- `notification/`: Notification tables

All migrations are applied in order based on their numeric prefix (001, 002, ...).
