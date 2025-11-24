# Database Agnostic Refactor Guide

## Overview

This document outlines the refactoring plan to make all module repositories database-agnostic by abstracting SQL dialect differences into the platform layer. This approach maintains hexagonal architecture principles while enabling support for multiple databases (SQLite, PostgreSQL, etc.) without changing module code.

## Goals

1. **Single Source of Truth**: All SQL dialect logic lives in `internal/platform/database`
2. **Module Independence**: Repository adapters don't know about specific database implementations
3. **Easy Testing**: Mock `database.DB` interface in unit tests
4. **Maintainability**: Add new database support by only changing platform layer
5. **Architecture Compliance**: Platform handles infrastructure, modules stay pure

## Current State

```
Module Repositories (Currently):
├── internal/modules/auth/adapters/
│   └── sqlite_session_repository.go   ← SQLite-specific (uses ?, CURRENT_TIMESTAMP, etc.)
├── internal/modules/user/adapters/
│   └── sqlite_user_repository.go      ← SQLite-specific
└── ... (other modules)

Platform (Currently):
├── internal/platform/database/
│   ├── connection.go                  ← Generic sql.DB wrapper
│   ├── migrator.go                    ← Migration runner
│   └── transaction.go                 ← Transaction helpers
```

## Target State

```
Module Repositories (After Refactor):
├── internal/modules/auth/adapters/
│   └── session_repository.go          ← Database-agnostic (uses DB interface)
├── internal/modules/user/adapters/
│   └── user_repository.go             ← Database-agnostic
└── ... (other modules)

Platform (After Refactor):
├── internal/platform/database/
│   ├── connection.go                  ← Generic connection management
│   ├── migrator.go                    ← Migration runner
│   ├── transaction.go                 ← Transaction helpers
│   ├── database.go                    ← DB interface definition
│   ├── sqlite.go                      ← SQLite implementation
│   ├── postgres.go                    ← PostgreSQL implementation
│   └── query_builder.go               ← Optional: Query building helpers
```

## Architecture Principles

### Hexagonal Architecture Compliance

```
┌─────────────────────────────────────────────────────────────┐
│                    Module Layer (auth, user, etc.)          │
│                                                              │
│  Adapters (OUTPUT ADAPTERS)                                 │
│  ├── session_repository.go  ───┐                           │
│  └── user_repository.go     ───┤ Depend on database.DB     │
│                                 │ (platform interface)      │
└─────────────────────────────────┼──────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────┐
│                    Platform Layer                            │
│                                                              │
│  internal/platform/database/                                │
│  ├── database.go (DB interface)  ← Abstract contract       │
│  ├── sqlite.go                   ← SQLite implementation    │
│  └── postgres.go                 ← PostgreSQL implementation│
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Key Points**:

- Modules depend on platform **interfaces**, not implementations
- Platform provides multiple implementations of the same interface
- Dependency injection happens at startup (wire package)
- No circular dependencies

---

## Implementation Steps

### Phase 1: Create Platform Database Abstraction

**Objective**: Define the `DB` interface that abstracts SQL dialect differences.

**Files to Create**:

1. `internal/platform/database/database.go` - Core interface definition
2. `internal/platform/database/sqlite.go` - SQLite implementation
3. `internal/platform/database/postgres.go` - PostgreSQL implementation (optional, can be added later)

**Implementation Order**:

1. Start with `database.go` to define the interface
2. Implement SQLite first (since it's currently used)
3. Test with existing code before refactoring modules
4. Add PostgreSQL support once SQLite is working

### Phase 2: Refactor Module Repositories

**Objective**: Update all module repositories to use `database.DB` interface instead of `*sql.DB`.

**Modules to Update** (in order of priority):

1. **auth module** - `session_repository.go`
2. **user module** - `user_repository.go`
3. **post module** - `post_repository.go`
4. **comment module** - `comment_repository.go`
5. **reaction module** - `reaction_repository.go`
6. **moderation module** - `report_repository.go` (optional)
7. **notification module** - `notification_repository.go` (optional)

**Pattern for Each Repository**:

1. Change constructor parameter from `*sql.DB` to `database.DB`
2. Replace hardcoded `?` placeholders with `r.db.Placeholder(n)`
3. Replace `CURRENT_TIMESTAMP` with `r.db.Now()`
4. Add helper method for generating multiple placeholders
5. Update tests to use mock `database.DB`

### Phase 3: Update Dependency Injection

**Objective**: Wire the correct database implementation at startup.

**Files to Update**:

1. `cmd/forum/wire/app.go` - Initialize database wrapper
2. `cmd/forum/wire/repositories.go` - Pass `database.DB` to repositories
3. `internal/platform/config/config.go` - Add database driver configuration

**Key Changes**:

- Replace `dbConn.DB()` calls with `database.NewSQLiteDB(dbConn.DB())`
- Add configuration option to choose database driver
- Support environment variable `DATABASE_DRIVER=sqlite|postgres`

### Phase 4: Migration Strategy

**Objective**: Support migrations for multiple database types.

**Approach**:

1. Keep existing migrations in `migrations/` for SQLite
2. Create `migrations/postgres/` for PostgreSQL-specific migrations
3. Update `database.Migrator` to select migration path based on driver
4. Document migration differences between databases

**Key Differences to Handle**:

- **Auto-increment**: `AUTOINCREMENT` (SQLite) vs `SERIAL` (PostgreSQL)
- **Timestamps**: `DATETIME` (SQLite) vs `TIMESTAMP` (PostgreSQL)
- **Boolean**: `INTEGER` (SQLite) vs `BOOLEAN` (PostgreSQL)
- **JSON**: `TEXT` (SQLite) vs `JSONB` (PostgreSQL)

### Phase 5: Testing

**Objective**: Ensure all repositories work with both database implementations.

**Test Strategy**:

1. **Unit Tests**: Mock `database.DB` interface
2. **Integration Tests**: Test with real SQLite database
3. **Integration Tests**: Test with real PostgreSQL database (Docker)
4. **Migration Tests**: Verify migrations work for both databases

**Test Files to Update**:

- `tests/unit/` - Add mock DB interface
- `tests/integration/` - Test with both SQLite and PostgreSQL
- `internal/platform/database/database_test.go` - Test DB implementations

---

## Technical Specification

### 1. Database Interface (database.go)

The core abstraction that all database implementations must satisfy.

```go
package database

import (
	"context"
	"database/sql"
)

// DB provides a database-agnostic interface for SQL operations.
// This interface wraps database/sql.DB and adds dialect-specific helpers.
type DB interface {
	// Core database/sql methods
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Close() error
	Ping(ctx context.Context) error
	
	// Dialect-specific helpers
	Placeholder(n int) string           // Returns $1 (Postgres) or ? (SQLite) for parameter n
	Now() string                        // Returns NOW() or CURRENT_TIMESTAMP
	AutoIncrement() string              // Returns SERIAL or AUTOINCREMENT
	Returning(cols ...string) string   // Returns RETURNING clause or empty string
	
	// Utility methods
	DriverName() string                 // Returns "sqlite3", "postgres", etc.
	Placeholders(count int) string      // Returns comma-separated placeholders: ?, ?, ?
}
```

**Design Decisions**:

- **Extends `database/sql`**: All standard methods are available
- **Placeholder system**: Handles different parameter markers (`?` vs `$1`)
- **Dialect helpers**: Encapsulates database-specific SQL syntax
- **No query builder**: Keeps interface minimal; repositories write SQL
- **Context-aware**: All methods accept `context.Context` for cancellation

### 2. SQLite Implementation (sqlite.go)

```go
package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteDB wraps sql.DB for SQLite-specific operations.
type SQLiteDB struct {
	db *sql.DB
}

// NewSQLiteDB creates a new SQLite database wrapper.
func NewSQLiteDB(db *sql.DB) DB {
	return &SQLiteDB{db: db}
}

// ExecContext executes a query without returning any rows.
func (s *SQLiteDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return s.db.ExecContext(ctx, query, args...)
}

// QueryContext executes a query that returns rows.
func (s *SQLiteDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return s.db.QueryContext(ctx, query, args...)
}

// QueryRowContext executes a query that is expected to return at most one row.
func (s *SQLiteDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return s.db.QueryRowContext(ctx, query, args...)
}

// BeginTx starts a transaction.
func (s *SQLiteDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return s.db.BeginTx(ctx, opts)
}

// Close closes the database connection.
func (s *SQLiteDB) Close() error {
	return s.db.Close()
}

// Ping verifies a connection to the database is still alive.
func (s *SQLiteDB) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// Placeholder returns ? for SQLite (position-agnostic).
// The parameter n is ignored for SQLite since it uses ? for all positions.
func (s *SQLiteDB) Placeholder(n int) string {
	return "?"
}

// Now returns SQLite's CURRENT_TIMESTAMP function.
func (s *SQLiteDB) Now() string {
	return "CURRENT_TIMESTAMP"
}

// AutoIncrement returns SQLite's AUTOINCREMENT keyword.
func (s *SQLiteDB) AutoIncrement() string {
	return "AUTOINCREMENT"
}

// Returning returns empty string (SQLite doesn't support RETURNING in older versions).
// For SQLite 3.35+, this could return a RETURNING clause, but we keep it empty for compatibility.
func (s *SQLiteDB) Returning(cols ...string) string {
	return ""
}

// DriverName returns the driver name.
func (s *SQLiteDB) DriverName() string {
	return "sqlite3"
}

// Placeholders generates a comma-separated list of placeholders.
// Example: Placeholders(3) returns "?, ?, ?"
func (s *SQLiteDB) Placeholders(count int) string {
	if count <= 0 {
		return ""
	}
	placeholders := make([]string, count)
	for i := 0; i < count; i++ {
		placeholders[i] = s.Placeholder(i + 1)
	}
	return strings.Join(placeholders, ", ")
}
```

**Key Features**:

- **Position-agnostic placeholders**: Always returns `?`
- **Compatibility**: Works with SQLite 3.x (most common version)
- **No RETURNING**: Older SQLite doesn't support it (use `LastInsertId()` instead)
- **Simple wrapper**: Delegates most work to `sql.DB`

### 3. PostgreSQL Implementation (postgres.go)

```go
package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	
	_ "github.com/lib/pq"
)

// PostgresDB wraps sql.DB for PostgreSQL-specific operations.
type PostgresDB struct {
	db *sql.DB
}

// NewPostgresDB creates a new PostgreSQL database wrapper.
func NewPostgresDB(db *sql.DB) DB {
	return &PostgresDB{db: db}
}

// ExecContext executes a query without returning any rows.
func (p *PostgresDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return p.db.ExecContext(ctx, query, args...)
}

// QueryContext executes a query that returns rows.
func (p *PostgresDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return p.db.QueryContext(ctx, query, args...)
}

// QueryRowContext executes a query that is expected to return at most one row.
func (p *PostgresDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return p.db.QueryRowContext(ctx, query, args...)
}

// BeginTx starts a transaction.
func (p *PostgresDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return p.db.BeginTx(ctx, opts)
}

// Close closes the database connection.
func (p *PostgresDB) Close() error {
	return p.db.Close()
}

// Ping verifies a connection to the database is still alive.
func (p *PostgresDB) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

// Placeholder returns $n for PostgreSQL (position-aware).
// Example: Placeholder(1) returns "$1", Placeholder(2) returns "$2"
func (p *PostgresDB) Placeholder(n int) string {
	return fmt.Sprintf("$%d", n)
}

// Now returns PostgreSQL's NOW() function.
func (p *PostgresDB) Now() string {
	return "NOW()"
}

// AutoIncrement returns SERIAL for PostgreSQL.
// Note: In CREATE TABLE, use SERIAL for auto-incrementing integer columns.
func (p *PostgresDB) AutoIncrement() string {
	return "SERIAL"
}

// Returning returns RETURNING clause for PostgreSQL.
// Example: Returning("id", "created_at") returns "RETURNING id, created_at"
func (p *PostgresDB) Returning(cols ...string) string {
	if len(cols) == 0 {
		return ""
	}
	return "RETURNING " + strings.Join(cols, ", ")
}

// DriverName returns the driver name.
func (p *PostgresDB) DriverName() string {
	return "postgres"
}

// Placeholders generates a comma-separated list of placeholders.
// Example: Placeholders(3) returns "$1, $2, $3"
func (p *PostgresDB) Placeholders(count int) string {
	if count <= 0 {
		return ""
	}
	placeholders := make([]string, count)
	for i := 0; i < count; i++ {
		placeholders[i] = p.Placeholder(i + 1)
	}
	return strings.Join(placeholders, ", ")
}
```

**Key Features**:

- **Position-aware placeholders**: Returns `$1`, `$2`, etc.
- **RETURNING support**: PostgreSQL supports returning values from INSERT/UPDATE/DELETE
- **Standard library compatible**: Works with `github.com/lib/pq` driver
- **Production-ready**: PostgreSQL is enterprise-grade database

**Placeholder Comparison**:

```sql
-- SQLite
INSERT INTO users (name, email) VALUES (?, ?)

-- PostgreSQL
INSERT INTO users (name, email) VALUES ($1, $2)
```

---

## Repository Refactoring Guide

### Before and After Example: Session Repository

#### Before (SQLite-specific)

**File**: `internal/modules/auth/adapters/sqlite_session_repository.go`

```go
package adapters

import (
	"context"
	"database/sql"
	"forum/internal/modules/auth/domain"
	"forum/internal/modules/auth/ports"
)

type SQLiteSessionRepository struct {
	db *sql.DB  // ← Direct sql.DB dependency
}

func NewSQLiteSessionRepository(db *sql.DB) ports.SessionRepository {
	return &SQLiteSessionRepository{db: db}
}

func (r *SQLiteSessionRepository) Create(ctx context.Context, session *domain.Session) error {
	// Hardcoded SQLite syntax
	query := `INSERT INTO sessions (id, user_id, token, expires_at, created_at, ip_address, user_agent)
	          VALUES (?, ?, ?, ?, ?, ?, ?)`  // ← Hardcoded ?
	
	_, err := r.db.ExecContext(ctx, query,
		session.ID, session.UserID, session.Token,
		session.ExpiresAt, session.CreatedAt,
		session.IPAddress, session.UserAgent,
	)
	return err
}

func (r *SQLiteSessionRepository) GetByUserID(ctx context.Context, userID int) ([]*domain.Session, error) {
	// Hardcoded CURRENT_TIMESTAMP
	query := `SELECT id, user_id, token, expires_at, created_at, ip_address, user_agent
	          FROM sessions WHERE user_id = ? AND expires_at > CURRENT_TIMESTAMP`  // ← Hardcoded
	
	rows, err := r.db.QueryContext(ctx, query, userID)
	// ... rest of implementation
}
```

#### After (Database-agnostic)

**File**: `internal/modules/auth/adapters/session_repository.go`

```go
package adapters

import (
	"context"
	"database/sql"
	"forum/internal/modules/auth/domain"
	"forum/internal/modules/auth/ports"
	"forum/internal/platform/database"  // ← Import platform abstraction
)

type SessionRepository struct {
	db database.DB  // ← Use platform DB interface
}

func NewSessionRepository(db database.DB) ports.SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, session *domain.Session) error {
	// Use Placeholders() helper
	query := `INSERT INTO sessions (id, user_id, token, expires_at, created_at, ip_address, user_agent)
	          VALUES (` + r.db.Placeholders(7) + `)`  // ← Dynamic placeholders
	
	_, err := r.db.ExecContext(ctx, query,
		session.ID, session.UserID, session.Token,
		session.ExpiresAt, session.CreatedAt,
		session.IPAddress, session.UserAgent,
	)
	return err
}

func (r *SessionRepository) GetByUserID(ctx context.Context, userID int) ([]*domain.Session, error) {
	// Use Now() and Placeholder(n)
	query := `SELECT id, user_id, token, expires_at, created_at, ip_address, user_agent
	          FROM sessions 
	          WHERE user_id = ` + r.db.Placeholder(1) + ` 
	          AND expires_at > ` + r.db.Now()  // ← Dynamic SQL functions
	
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var sessions []*domain.Session
	for rows.Next() {
		session := &domain.Session{}
		if err := rows.Scan(
			&session.ID, &session.UserID, &session.Token,
			&session.ExpiresAt, &session.CreatedAt,
			&session.IPAddress, &session.UserAgent,
		); err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	
	return sessions, rows.Err()
}

func (r *SessionRepository) Delete(ctx context.Context, token string) error {
	query := `DELETE FROM sessions WHERE token = ` + r.db.Placeholder(1)
	_, err := r.db.ExecContext(ctx, query, token)
	return err
}

func (r *SessionRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at < ` + r.db.Now()
	_, err := r.db.ExecContext(ctx, query)
	return err
}
```

**Changes Made**:

1. ✅ Renamed file from `sqlite_session_repository.go` to `session_repository.go`
2. ✅ Renamed struct from `SQLiteSessionRepository` to `SessionRepository`
3. ✅ Changed parameter from `*sql.DB` to `database.DB`
4. ✅ Replaced `?` with `r.db.Placeholder(n)` or `r.db.Placeholders(count)`
5. ✅ Replaced `CURRENT_TIMESTAMP` with `r.db.Now()`
6. ✅ Updated imports to include `forum/internal/platform/database`

### Refactoring Checklist for Each Repository

Use this checklist when refactoring each module repository:

- [ ] Rename file to remove database-specific prefix (e.g., `sqlite_` → generic name)
- [ ] Rename struct to remove database-specific prefix
- [ ] Change constructor parameter from `*sql.DB` to `database.DB`
- [ ] Update all imports
- [ ] Replace hardcoded `?` with `r.db.Placeholder(n)`
- [ ] For multiple placeholders, use `r.db.Placeholders(count)` helper
- [ ] Replace `CURRENT_TIMESTAMP` with `r.db.Now()`
- [ ] Replace `NOW()` with `r.db.Now()`
- [ ] If using `RETURNING`, wrap with `r.db.Returning(...)`
- [ ] Update tests to mock `database.DB` interface
- [ ] Test with both SQLite and PostgreSQL (if available)

---

## Wiring and Configuration

### 1. Update Configuration (config.go)

Add database driver selection to configuration:

```go
// File: internal/platform/config/config.go

package config

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Session  SessionConfig
    Security SecurityConfig
    Upload   UploadConfig
    Logger   LoggerConfig
}

type DatabaseConfig struct {
    Driver   string // "sqlite" or "postgres"
    
    // SQLite configuration
    SQLite SQLiteConfig
    
    // PostgreSQL configuration
    Postgres PostgresConfig
    
    // Connection pooling (applies to both)
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime int // seconds
}

type SQLiteConfig struct {
    Path string // Path to SQLite database file (e.g., "./forum.db")
}

type PostgresConfig struct {
    Host     string
    Port     int
    User     string
    Password string
    DBName   string
    SSLMode  string // "disable", "require", "verify-ca", "verify-full"
}

// LoadConfig loads configuration from environment variables with defaults.
func LoadConfig() (*Config, error) {
    cfg := &Config{
        Database: DatabaseConfig{
            Driver: getEnv("DATABASE_DRIVER", "sqlite"),
            SQLite: SQLiteConfig{
                Path: getEnv("SQLITE_PATH", "./forum.db"),
            },
            Postgres: PostgresConfig{
                Host:     getEnv("POSTGRES_HOST", "localhost"),
                Port:     getEnvAsInt("POSTGRES_PORT", 5432),
                User:     getEnv("POSTGRES_USER", "forum"),
                Password: getEnv("POSTGRES_PASSWORD", ""),
                DBName:   getEnv("POSTGRES_DB", "forum"),
                SSLMode:  getEnv("POSTGRES_SSL_MODE", "disable"),
            },
            MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
            MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
            ConnMaxLifetime: getEnvAsInt("DB_CONN_MAX_LIFETIME", 300),
        },
        // ... other config fields
    }
    
    return cfg, nil
}
```

**Environment Variables**:

```bash
# Database driver selection
DATABASE_DRIVER=sqlite        # or "postgres"

# SQLite configuration
SQLITE_PATH=./forum.db

# PostgreSQL configuration
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=forum
POSTGRES_PASSWORD=secret
POSTGRES_DB=forum
POSTGRES_SSL_MODE=disable

# Connection pooling
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=300
```

### 2. Update Connection Logic (wire/app.go)

**File**: `cmd/forum/wire/app.go`

```go
package wire

import (
    "context"
    "database/sql"
    "fmt"
    "time"
    
    "forum/internal/platform/config"
    "forum/internal/platform/database"
    "forum/internal/platform/logger"
    
    _ "github.com/mattn/go-sqlite3"
    _ "github.com/lib/pq"
)

// InitializeDatabase creates and configures the database connection.
func InitializeDatabase(cfg *config.Config, lgr *logger.Logger) (database.DB, error) {
    var sqlDB *sql.DB
    var err error
    
    switch cfg.Database.Driver {
    case "sqlite":
        lgr.Info("Connecting to SQLite database", logger.String("path", cfg.Database.SQLite.Path))
        sqlDB, err = sql.Open("sqlite3", cfg.Database.SQLite.Path+"?_foreign_keys=on")
        if err != nil {
            return nil, fmt.Errorf("failed to open SQLite database: %w", err)
        }
        
    case "postgres":
        dsn := fmt.Sprintf(
            "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
            cfg.Database.Postgres.Host,
            cfg.Database.Postgres.Port,
            cfg.Database.Postgres.User,
            cfg.Database.Postgres.Password,
            cfg.Database.Postgres.DBName,
            cfg.Database.Postgres.SSLMode,
        )
        lgr.Info("Connecting to PostgreSQL database", 
            logger.String("host", cfg.Database.Postgres.Host),
            logger.String("database", cfg.Database.Postgres.DBName))
        sqlDB, err = sql.Open("postgres", dsn)
        if err != nil {
            return nil, fmt.Errorf("failed to open PostgreSQL database: %w", err)
        }
        
    default:
        return nil, fmt.Errorf("unsupported database driver: %s", cfg.Database.Driver)
    }
    
    // Configure connection pool
    sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
    sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
    sqlDB.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Second)
    
    // Verify connection
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := sqlDB.PingContext(ctx); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }
    
    lgr.Info("Database connection established")
    
    // Wrap with appropriate database implementation
    var db database.DB
    switch cfg.Database.Driver {
    case "sqlite":
        db = database.NewSQLiteDB(sqlDB)
    case "postgres":
        db = database.NewPostgresDB(sqlDB)
    }
    
    // Run migrations
    if err := runMigrations(db, cfg.Database.Driver, lgr); err != nil {
        return nil, fmt.Errorf("failed to run migrations: %w", err)
    }
    
    return db, nil
}

// runMigrations executes database migrations.
func runMigrations(db database.DB, driver string, lgr *logger.Logger) error {
    lgr.Info("Running database migrations", logger.String("driver", driver))
    
    // Select migration path based on driver
    migrationPath := "./migrations"
    if driver == "postgres" {
        migrationPath = "./migrations/postgres"
    }
    
    migrator := database.NewMigrator(db, migrationPath)
    if err := migrator.Up(); err != nil {
        return fmt.Errorf("migration failed: %w", err)
    }
    
    lgr.Info("Database migrations completed successfully")
    return nil
}
```

### 3. Update Repository Wiring (wire/repositories.go)

**File**: `cmd/forum/wire/repositories.go`

```go
package wire

import (
    authAdapters "forum/internal/modules/auth/adapters"
    userAdapters "forum/internal/modules/user/adapters"
    postAdapters "forum/internal/modules/post/adapters"
    commentAdapters "forum/internal/modules/comment/adapters"
    reactionAdapters "forum/internal/modules/reaction/adapters"
    
    "forum/internal/platform/database"
)

// Repositories holds all repository instances.
type Repositories struct {
    Session  authAdapters.SessionRepository
    User     userAdapters.UserRepository
    Post     postAdapters.PostRepository
    Comment  commentAdapters.CommentRepository
    Reaction reactionAdapters.ReactionRepository
}

// NewRepositories creates all repository instances.
// Changed: Now accepts database.DB instead of *sql.DB
func NewRepositories(db database.DB) *Repositories {
    return &Repositories{
        Session:  authAdapters.NewSessionRepository(db),
        User:     userAdapters.NewUserRepository(db),
        Post:     postAdapters.NewPostRepository(db),
        Comment:  commentAdapters.NewCommentRepository(db),
        Reaction: reactionAdapters.NewReactionRepository(db),
    }
}
```

**Key Change**: Pass `database.DB` interface instead of `*sql.DB`

### 4. Update Main Entry Point (main.go)

**File**: `cmd/forum/main.go`

```go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"
    
    "forum/cmd/forum/wire"
    "forum/internal/platform/config"
    "forum/internal/platform/logger"
)

func main() {
    // 1. Load configuration
    cfg, err := config.LoadConfig()
    if err != nil {
        logger.Fatal("Failed to load config", logger.Error(err))
    }
    
    // 2. Initialize logger
    lgr := logger.New(cfg.Logger.Level)
    lgr.Info("Starting Forum Application")
    
    // 3. Initialize database (returns database.DB interface)
    db, err := wire.InitializeDatabase(cfg, lgr)
    if err != nil {
        lgr.Fatal("Failed to initialize database", logger.Error(err))
    }
    defer db.Close()
    
    // 4. Wire dependencies
    repos := wire.NewRepositories(db)
    services := wire.NewServices(repos, lgr)
    handlers := wire.NewHandlers(services, lgr)
    
    // 5. Start HTTP server
    server := wire.NewHTTPServer(cfg, handlers, lgr)
    
    // 6. Graceful shutdown
    shutdown := make(chan os.Signal, 1)
    signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        if err := server.Start(); err != nil {
            lgr.Fatal("Server failed", logger.Error(err))
        }
    }()
    
    <-shutdown
    lgr.Info("Shutting down server...")
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    if err := server.Shutdown(ctx); err != nil {
        lgr.Error("Server shutdown error", logger.Error(err))
    }
    
    lgr.Info("Server stopped")
}
```

---

## Testing Strategy

### 1. Mock Database Interface

Create a mock implementation for unit tests:

**File**: `tests/unit/mock_database.go`

```go
package unit

import (
    "context"
    "database/sql"
    "forum/internal/platform/database"
)

// MockDB is a mock implementation of database.DB for testing.
type MockDB struct {
    ExecContextFunc      func(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
    QueryContextFunc     func(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
    QueryRowContextFunc  func(ctx context.Context, query string, args ...interface{}) *sql.Row
    BeginTxFunc          func(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
    CloseFunc            func() error
    PingFunc             func(ctx context.Context) error
    PlaceholderFunc      func(n int) string
    NowFunc              func() string
    AutoIncrementFunc    func() string
    ReturningFunc        func(cols ...string) string
    DriverNameFunc       func() string
    PlaceholdersFunc     func(count int) string
}

func (m *MockDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
    if m.ExecContextFunc != nil {
        return m.ExecContextFunc(ctx, query, args...)
    }
    return nil, nil
}

func (m *MockDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
    if m.QueryContextFunc != nil {
        return m.QueryContextFunc(ctx, query, args...)
    }
    return nil, nil
}

func (m *MockDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
    if m.QueryRowContextFunc != nil {
        return m.QueryRowContextFunc(ctx, query, args...)
    }
    return nil
}

func (m *MockDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
    if m.BeginTxFunc != nil {
        return m.BeginTxFunc(ctx, opts)
    }
    return nil, nil
}

func (m *MockDB) Close() error {
    if m.CloseFunc != nil {
        return m.CloseFunc()
    }
    return nil
}

func (m *MockDB) Ping(ctx context.Context) error {
    if m.PingFunc != nil {
        return m.PingFunc(ctx)
    }
    return nil
}

func (m *MockDB) Placeholder(n int) string {
    if m.PlaceholderFunc != nil {
        return m.PlaceholderFunc(n)
    }
    return "?"
}

func (m *MockDB) Now() string {
    if m.NowFunc != nil {
        return m.NowFunc()
    }
    return "CURRENT_TIMESTAMP"
}

func (m *MockDB) AutoIncrement() string {
    if m.AutoIncrementFunc != nil {
        return m.AutoIncrementFunc()
    }
    return "AUTOINCREMENT"
}

func (m *MockDB) Returning(cols ...string) string {
    if m.ReturningFunc != nil {
        return m.ReturningFunc(cols...)
    }
    return ""
}

func (m *MockDB) DriverName() string {
    if m.DriverNameFunc != nil {
        return m.DriverNameFunc()
    }
    return "mock"
}

func (m *MockDB) Placeholders(count int) string {
    if m.PlaceholdersFunc != nil {
        return m.PlaceholdersFunc(count)
    }
    placeholders := make([]string, count)
    for i := 0; i < count; i++ {
        placeholders[i] = m.Placeholder(i + 1)
    }
    return strings.Join(placeholders, ", ")
}
```

### 2. Unit Test Example

**File**: `internal/modules/auth/adapters/session_repository_test.go`

```go
package adapters

import (
    "context"
    "database/sql"
    "errors"
    "testing"
    "time"
    
    "forum/internal/modules/auth/domain"
    "forum/tests/unit"
)

func TestSessionRepository_Create(t *testing.T) {
    tests := []struct {
        name        string
        session     *domain.Session
        mockSetup   func(*unit.MockDB)
        expectError bool
    }{
        {
            name: "successful creation",
            session: &domain.Session{
                ID:        "session-123",
                UserID:    456,
                Token:     "token-abc",
                ExpiresAt: time.Now().Add(24 * time.Hour),
                CreatedAt: time.Now(),
            },
            mockSetup: func(m *unit.MockDB) {
                m.PlaceholdersFunc = func(count int) string {
                    return "?, ?, ?, ?, ?, ?, ?"
                }
                m.ExecContextFunc = func(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
                    // Verify query contains placeholders
                    if !strings.Contains(query, "INSERT INTO sessions") {
                        t.Errorf("expected INSERT query, got: %s", query)
                    }
                    return &mockResult{lastInsertID: 1, rowsAffected: 1}, nil
                }
            },
            expectError: false,
        },
        {
            name: "database error",
            session: &domain.Session{
                ID:     "session-456",
                UserID: 789,
            },
            mockSetup: func(m *unit.MockDB) {
                m.PlaceholdersFunc = func(count int) string { return "?, ?, ?, ?, ?, ?, ?" }
                m.ExecContextFunc = func(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
                    return nil, errors.New("database error")
                }
            },
            expectError: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockDB := &unit.MockDB{}
            tt.mockSetup(mockDB)
            
            repo := NewSessionRepository(mockDB)
            
            err := repo.Create(context.Background(), tt.session)
            
            if tt.expectError && err == nil {
                t.Error("expected error but got nil")
            }
            if !tt.expectError && err != nil {
                t.Errorf("unexpected error: %v", err)
            }
        })
    }
}

// mockResult implements sql.Result for testing
type mockResult struct {
    lastInsertID int64
    rowsAffected int64
}

func (m *mockResult) LastInsertId() (int64, error) {
    return m.lastInsertID, nil
}

func (m *mockResult) RowsAffected() (int64, error) {
    return m.rowsAffected, nil
}
```

### 3. Integration Test with Both Databases

**File**: `tests/integration/database_test.go`

```go
package integration

import (
    "context"
    "database/sql"
    "testing"
    
    "forum/internal/platform/database"
    
    _ "github.com/mattn/go-sqlite3"
    _ "github.com/lib/pq"
)

func TestDatabaseAbstraction_SQLite(t *testing.T) {
    sqlDB, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
    if err != nil {
        t.Fatalf("failed to open SQLite: %v", err)
    }
    defer sqlDB.Close()
    
    db := database.NewSQLiteDB(sqlDB)
    testDatabaseInterface(t, db, "sqlite")
}

func TestDatabaseAbstraction_Postgres(t *testing.T) {
    // Skip if no PostgreSQL test database available
    dsn := getTestPostgresDSN()
    if dsn == "" {
        t.Skip("PostgreSQL test database not configured")
    }
    
    sqlDB, err := sql.Open("postgres", dsn)
    if err != nil {
        t.Fatalf("failed to open PostgreSQL: %v", err)
    }
    defer sqlDB.Close()
    
    db := database.NewPostgresDB(sqlDB)
    testDatabaseInterface(t, db, "postgres")
}

// testDatabaseInterface runs common tests for any DB implementation
func testDatabaseInterface(t *testing.T, db database.DB, driver string) {
    t.Run("Ping", func(t *testing.T) {
        err := db.Ping(context.Background())
        if err != nil {
            t.Errorf("Ping failed: %v", err)
        }
    })
    
    t.Run("Placeholder", func(t *testing.T) {
        p1 := db.Placeholder(1)
        if p1 == "" {
            t.Error("Placeholder(1) returned empty string")
        }
        
        if driver == "sqlite" && p1 != "?" {
            t.Errorf("SQLite placeholder should be '?', got: %s", p1)
        }
        if driver == "postgres" && p1 != "$1" {
            t.Errorf("Postgres placeholder should be '$1', got: %s", p1)
        }
    })
    
    t.Run("Placeholders", func(t *testing.T) {
        p3 := db.Placeholders(3)
        if driver == "sqlite" && p3 != "?, ?, ?" {
            t.Errorf("SQLite placeholders incorrect: %s", p3)
        }
        if driver == "postgres" && p3 != "$1, $2, $3" {
            t.Errorf("Postgres placeholders incorrect: %s", p3)
        }
    })
    
    t.Run("Now", func(t *testing.T) {
        now := db.Now()
        if now == "" {
            t.Error("Now() returned empty string")
        }
    })
    
    t.Run("DriverName", func(t *testing.T) {
        name := db.DriverName()
        if name != driver {
            t.Errorf("expected driver name %s, got: %s", driver, name)
        }
    })
}

func getTestPostgresDSN() string {
    // Read from environment or return empty to skip
    return os.Getenv("TEST_POSTGRES_DSN")
}
```

### 4. Run Tests with Docker Compose

**File**: `docker-compose.test.yml`

```yaml
version: '3.8'

services:
  postgres-test:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: test
      POSTGRES_PASSWORD: test
      POSTGRES_DB: forum_test
    ports:
      - "5433:5432"
    tmpfs:
      - /var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U test"]
      interval: 5s
      timeout: 5s
      retries: 5
```

**Test Command**:

```bash
# Start test database
docker-compose -f docker-compose.test.yml up -d

# Wait for database to be ready
sleep 3

# Run tests with PostgreSQL
TEST_POSTGRES_DSN="host=localhost port=5433 user=test password=test dbname=forum_test sslmode=disable" \
go test ./tests/integration/... -v

# Run all tests
go test ./... -v

# Stop test database
docker-compose -f docker-compose.test.yml down
```

---

## Migration Guide

### Migration Directory Structure

```text
migrations/
├── sqlite/                          ← SQLite-specific migrations
│   ├── 001_auth_create_sessions.sql
│   ├── 002_user_create_users.sql
│   ├── 003_post_create_tables.sql
│   └── ...
├── postgres/                        ← PostgreSQL-specific migrations
│   ├── 001_auth_create_sessions.sql
│   ├── 002_user_create_users.sql
│   ├── 003_post_create_tables.sql
│   └── ...
└── README.md
```

### Database-Specific Differences

#### 1. Auto-Increment Fields

**SQLite**:

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE
);
```

**PostgreSQL**:

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE
);
```

#### 2. Timestamps

**SQLite**:

```sql
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

**PostgreSQL**:

```sql
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```

#### 3. Boolean Fields

**SQLite**:

```sql
CREATE TABLE posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    is_deleted INTEGER DEFAULT 0 CHECK (is_deleted IN (0, 1))
);
```

**PostgreSQL**:

```sql
CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    is_deleted BOOLEAN DEFAULT FALSE
);
```

#### 4. Text Fields

**SQLite**:

```sql
CREATE TABLE posts (
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    metadata TEXT  -- JSON stored as text
);
```

**PostgreSQL**:

```sql
CREATE TABLE posts (
    title VARCHAR(500) NOT NULL,
    content TEXT NOT NULL,
    metadata JSONB  -- Native JSON support
);
```

### Example Migration Pair

**SQLite**: `migrations/sqlite/001_auth_create_sessions.sql`

```sql
-- +migrate Up
CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL,
    token TEXT UNIQUE NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    ip_address TEXT,
    user_agent TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_sessions_token ON sessions(token);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- +migrate Down
DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP INDEX IF EXISTS idx_sessions_token;
DROP TABLE IF EXISTS sessions;
```

**PostgreSQL**: `migrations/postgres/001_auth_create_sessions.sql`

```sql
-- +migrate Up
CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL,
    token TEXT UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    ip_address VARCHAR(45),  -- IPv6 support
    user_agent TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_sessions_token ON sessions(token);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- +migrate Down
DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP INDEX IF EXISTS idx_sessions_token;
DROP TABLE IF EXISTS sessions;
```

---

## Implementation Roadmap

### Phase 1: Platform Layer (Week 1)

**Priority**: HIGH - Foundation for all other work

- [ ] Create `internal/platform/database/database.go` - Define DB interface
- [ ] Create `internal/platform/database/sqlite.go` - Implement SQLite wrapper
- [ ] Create `internal/platform/database/postgres.go` - Implement PostgreSQL wrapper
- [ ] Update `internal/platform/config/config.go` - Add database driver config
- [ ] Write tests for `sqlite.go` and `postgres.go`
- [ ] Update `cmd/forum/wire/app.go` - Add database initialization logic

**Success Criteria**:
- ✅ DB interface compiles and is well-documented
- ✅ SQLite implementation passes all tests
- ✅ PostgreSQL implementation passes all tests
- ✅ Configuration loads database driver from environment
- ✅ Connection logic works for both databases

### Phase 2: Auth Module (Week 1-2)

**Priority**: HIGH - Most critical for user sessions

- [ ] Refactor `internal/modules/auth/adapters/sqlite_session_repository.go`
  - [ ] Rename to `session_repository.go`
  - [ ] Change constructor to accept `database.DB`
  - [ ] Replace `?` with `r.db.Placeholder(n)`
  - [ ] Replace `CURRENT_TIMESTAMP` with `r.db.Now()`
  - [ ] Add `Placeholders()` helper usage
- [ ] Write unit tests with mock DB
- [ ] Write integration tests with both SQLite and PostgreSQL
- [ ] Update `cmd/forum/wire/repositories.go` to use new repository

**Success Criteria**:
- ✅ Session repository works with both databases
- ✅ All tests pass
- ✅ No hardcoded SQL dialect

### Phase 3: User Module (Week 2)

**Priority**: HIGH - Core functionality

- [ ] Refactor `internal/modules/user/adapters/sqlite_user_repository.go`
  - [ ] Rename to `user_repository.go`
  - [ ] Update constructor and all methods
  - [ ] Replace dialect-specific SQL
- [ ] Write tests
- [ ] Update wiring

**Success Criteria**:
- ✅ User repository is database-agnostic
- ✅ Tests pass for both databases

### Phase 4: Post Module (Week 2-3)

**Priority**: MEDIUM - Core content

- [ ] Refactor `internal/modules/post/adapters/sqlite_post_repository.go`
  - [ ] Rename to `post_repository.go`
  - [ ] Update for database abstraction
- [ ] Refactor category repository (if separate)
- [ ] Write tests
- [ ] Update wiring

**Success Criteria**:
- ✅ Post operations work with both databases
- ✅ Category associations work correctly

### Phase 5: Comment Module (Week 3)

**Priority**: MEDIUM - Social features

- [ ] Refactor `internal/modules/comment/adapters/sqlite_comment_repository.go`
  - [ ] Rename to `comment_repository.go`
  - [ ] Update for database abstraction
- [ ] Write tests
- [ ] Update wiring

**Success Criteria**:
- ✅ Comment CRUD works with both databases

### Phase 6: Reaction Module (Week 3)

**Priority**: MEDIUM - Social features

- [ ] Refactor `internal/modules/reaction/adapters/sqlite_reaction_repository.go`
  - [ ] Rename to `reaction_repository.go`
  - [ ] Update for database abstraction
- [ ] Write tests
- [ ] Update wiring

**Success Criteria**:
- ✅ Like/dislike works with both databases
- ✅ Reaction counts are accurate

### Phase 7: Optional Modules (Week 4)

**Priority**: LOW - Can be deferred

- [ ] Moderation module repositories
- [ ] Notification module repositories
- [ ] Write tests for optional modules

### Phase 8: Migrations (Week 4)

**Priority**: MEDIUM - Required for PostgreSQL support

- [ ] Create `migrations/postgres/` directory
- [ ] Convert all SQLite migrations to PostgreSQL syntax
- [ ] Update `internal/platform/database/migrator.go` to support driver-specific paths
- [ ] Test migrations on both databases
- [ ] Document migration differences

**Success Criteria**:
- ✅ Migrations run successfully on SQLite
- ✅ Migrations run successfully on PostgreSQL
- ✅ Database schemas are functionally equivalent

### Phase 9: Integration Testing (Week 4-5)

**Priority**: HIGH - Verification

- [ ] Write end-to-end tests with SQLite
- [ ] Write end-to-end tests with PostgreSQL
- [ ] Test all audit scenarios with both databases
- [ ] Performance benchmarking
- [ ] Load testing with both databases

**Success Criteria**:
- ✅ All audit requirements pass with SQLite
- ✅ All audit requirements pass with PostgreSQL
- ✅ No performance regressions
- ✅ Database can be switched with config change only

### Phase 10: Documentation and Deployment (Week 5)

**Priority**: MEDIUM - Operations

- [ ] Update README.md with database configuration
- [ ] Update Dockerfile to support both databases
- [ ] Update docker-compose.yml with PostgreSQL option
- [ ] Write deployment guide for both databases
- [ ] Update environment variable documentation

**Success Criteria**:
- ✅ Clear documentation for database selection
- ✅ Docker setup works for both databases
- ✅ Production deployment guide is complete

---

## Benefits of This Approach

### 1. Architectural Benefits

- **Hexagonal Compliance**: Platform handles infrastructure, modules stay pure
- **Single Responsibility**: Database dialect logic isolated in platform layer
- **Open/Closed Principle**: Add new databases without modifying modules
- **Dependency Inversion**: Modules depend on abstractions, not concrete implementations

### 2. Development Benefits

- **Easy Testing**: Mock `database.DB` interface in unit tests
- **Flexibility**: Switch databases with configuration change
- **Maintainability**: One place to update for new database support
- **Clarity**: Clear separation between business logic and database specifics

### 3. Operational Benefits

- **Database Choice**: Start with SQLite, scale to PostgreSQL
- **Migration Path**: Easy migration from SQLite to PostgreSQL
- **Development Environment**: Use SQLite locally, PostgreSQL in production
- **Testing**: Test with both databases in CI/CD pipeline

### 4. Future-Proofing

- **New Databases**: Add MySQL, CockroachDB, etc. by implementing `DB` interface
- **Query Optimization**: Database-specific optimizations in platform layer
- **Feature Support**: Leverage database-specific features when available
- **Cloud Databases**: Easy integration with managed database services

---

## Common Pitfalls and Solutions

### Pitfall 1: Complex Queries with Placeholders

**Problem**: Building dynamic queries with variable numbers of placeholders.

**Solution**: Use `Placeholders(count)` helper or query builder pattern.

```go
// Bad: Manual placeholder counting
query := "INSERT INTO posts (title, content, user_id) VALUES (?, ?, ?)"

// Good: Use helper
query := "INSERT INTO posts (title, content, user_id) VALUES (" + r.db.Placeholders(3) + ")"
```

### Pitfall 2: Database-Specific Features

**Problem**: Using features only available in one database (e.g., RETURNING).

**Solution**: Check feature availability or use conditional logic.

```go
// Use RETURNING if supported
returningClause := r.db.Returning("id", "created_at")
query := "INSERT INTO posts (title, content) VALUES (" + r.db.Placeholders(2) + ") " + returningClause

if returningClause != "" {
    // PostgreSQL: Get values from RETURNING
    err := r.db.QueryRowContext(ctx, query, title, content).Scan(&id, &createdAt)
} else {
    // SQLite: Use LastInsertId
    result, err := r.db.ExecContext(ctx, query, title, content)
    id, _ = result.LastInsertId()
}
```

### Pitfall 3: Transaction Handling

**Problem**: Transactions need to use same placeholder style.

**Solution**: Pass `database.DB` to transaction helpers, not raw `*sql.Tx`.

```go
// Good: Transaction helper that maintains abstraction
func (r *Repository) WithTransaction(ctx context.Context, fn func(database.DB) error) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    
    txDB := database.NewTxWrapper(tx, r.db)  // Wrapper maintains placeholder logic
    
    if err := fn(txDB); err != nil {
        tx.Rollback()
        return err
    }
    
    return tx.Commit()
}
```

### Pitfall 4: Migration Differences

**Problem**: Schema differences cause subtle bugs.

**Solution**: Maintain parallel migration files and test thoroughly.

- Use INTEGER for booleans in SQLite, BOOLEAN in PostgreSQL
- Always test migrations on both databases
- Document schema differences in README

---

## Conclusion

This refactoring plan provides a **database-agnostic architecture** while maintaining **hexagonal principles**. The platform layer handles all database-specific concerns, allowing modules to remain pure and focused on business logic.

**Next Steps**:

1. Review this proposal with the team
2. Start with Phase 1 (Platform Layer)
3. Test thoroughly before moving to next phase
4. Update documentation as you progress
5. Celebrate when all modules are database-agnostic! 🎉

**Questions or Issues?**

- Check existing implementations in `internal/platform/database/`
- Review hexagonal architecture docs in `docs/ARCHITECTURE.md`
- Test each change with both SQLite and PostgreSQL
- Ask for code review before merging major changes

