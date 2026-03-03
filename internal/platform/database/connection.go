package database

// Package database provides database connection management and migrations.
// This package handles SQLite database initialization, connection pooling,
// and migration execution for the forum application.

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Connection represents a database connection manager.
// It handles connection lifecycle and provides access to the database.
type Connection struct {
	db *sql.DB
}

// ConnectionConfig holds connection pool settings for the database.
type ConnectionConfig struct {
	MaxOpenConns    int           // Maximum number of open connections
	MaxIdleConns    int           // Maximum number of idle connections
	ConnMaxLifetime time.Duration // Maximum connection lifetime
}

// NewConnection creates a new SQLite database connection manager with default
// pool settings. It ensures the parent directory for the database file exists,
// opens the connection and returns a Connection wrapper.
func NewConnection(dsn string) (*Connection, error) {
	return NewConnectionWithConfig(dsn, ConnectionConfig{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 30 * time.Minute,
	})
}

// NewConnectionWithConfig creates a new SQLite database connection manager
// with the provided pool configuration. It ensures the parent directory for
// the database file exists, opens the connection and returns a Connection wrapper.
func NewConnectionWithConfig(dsn string, cfg ConnectionConfig) (*Connection, error) {
	// If the DSN is a simple file path (e.g. './data/forum.db'), ensure its
	// parent directory exists. For more complex DSNs (file:... with params)
	// try to extract the file portion up to the first '?' char.
	dbPath := dsn
	// If DSN looks like URI with params, strip params for directory creation.
	if idx := strings.IndexByte(dsn, '?'); idx != -1 {
		dbPath = dsn[:idx]
	}

	// If DSN starts with file:, remove the scheme for filesystem ops.
	if len(dbPath) > 5 && dbPath[:5] == "file:" {
		dbPath = dbPath[5:]
	}

	dir := filepath.Dir(dbPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create database directory %s: %w", dir, err)
		}
	}

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// It's okay to return the connection even if ping fails here; callers may
	// want to handle migration or retries. We'll still attempt a Ping.
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping sqlite database: %w", err)
	}

	// Use WAL (Write-Ahead Logging) journal mode for better concurrency and
	// durability compared to MEMORY mode. WAL allows readers and writers to
	// operate concurrently and provides crash recovery.
	if _, err := db.Exec("PRAGMA journal_mode = WAL;"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to set journal_mode=WAL: %w", err)
	}

	// Apply connection pool settings
	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	return &Connection{db: db}, nil
}

// DB returns the underlying database connection.
// This is used by repositories to execute queries.
func (c *Connection) DB() *sql.DB {
	return c.db
}

// Close closes the database connection.
// Should be called when the application shuts down.
func (c *Connection) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}
