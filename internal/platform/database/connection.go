package database
// Package database provides database connection management and migrations.
// This package handles SQLite database initialization, connection pooling,
// and migration execution for the forum application.
package database

import (
	"database/sql"
)

// Connection represents a database connection manager.
// It handles connection lifecycle and provides access to the database.
type Connection struct {
	db *sql.DB
}

// NewConnection creates a new database connection manager.
// TODO: Implement connection initialization with proper configuration.
func NewConnection(dsn string) (*Connection, error) {
	// Implementation placeholder
	return nil, nil
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

// Ping checks if the database connection is alive.
func (c *Connection) Ping() error {
	if c.db != nil {
		return c.db.Ping()
	}
	return nil
}
