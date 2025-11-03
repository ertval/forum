// Package database provides database connection management and migrations.
package database

import (
	"context"
	"database/sql"
)

// Transaction represents a database transaction.
// It wraps sql.Tx to provide a consistent interface for transactional operations.
type Transaction struct {
	tx *sql.Tx
}

// Begin starts a new database transaction.
// Transactions should be committed or rolled back when done.
func (c *Connection) Begin(ctx context.Context) (*Transaction, error) {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &Transaction{tx: tx}, nil
}

// Commit commits the transaction.
// All changes made within the transaction are persisted to the database.
func (t *Transaction) Commit() error {
	if t.tx != nil {
		return t.tx.Commit()
	}
	return nil
}

// Rollback rolls back the transaction.
// All changes made within the transaction are discarded.
func (t *Transaction) Rollback() error {
	if t.tx != nil {
		return t.tx.Rollback()
	}
	return nil
}

// Tx returns the underlying sql.Tx for query execution.
func (t *Transaction) Tx() *sql.Tx {
	return t.tx
}
