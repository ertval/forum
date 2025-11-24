// Package database provides database connection management and migrations.
package database

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Migrator handles database schema migrations.
// It executes SQL migration files in order to initialize and update the database schema.
type Migrator struct {
	conn *Connection
}

// NewMigrator creates a new database migrator.
// TODO: Implement migrator initialization.
func NewMigrator(conn *Connection) *Migrator {
	return &Migrator{
		conn: conn,
	}
}

// Migrate runs all pending migrations.
// Migrations are executed in order based on their version numbers.
// TODO: Implement migration execution logic.
func (m *Migrator) Migrate(migrationsPath string) error {
	// Ensure schema_migrations table exists
	_, err := m.conn.DB().Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at DATETIME NOT NULL
		);
	`)
	if err != nil {
		return err
	}

	// Read migration files
	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		return err
	}

	// Get applied migration versions
	rows, err := m.conn.DB().Query("SELECT version FROM schema_migrations")
	if err != nil {
		return err
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return err
		}
		applied[v] = true
	}

	// Sort migration files by version
	type migrationFile struct {
		version int
		name    string
		path    string
	}
	migrations := []migrationFile{}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name := f.Name()
		var version int
		_, err := fmt.Sscanf(name, "%d_", &version)
		if err != nil {
			continue // skip files not matching pattern
		}
		migrations = append(migrations, migrationFile{
			version: version,
			name:    name,
			path:    filepath.Join(migrationsPath, name),
		})
	}
	slices.SortFunc(migrations, func(a, b migrationFile) int {
		return a.version - b.version
	})

	// Apply pending migrations
	for _, mig := range migrations {
		if applied[mig.version] {
			continue // already applied
		}
		content, err := os.ReadFile(mig.path)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", mig.name, err)
		}
		// Only apply the Up section
		upSQL := extractUpSQL(string(content))
		if upSQL == "" {
			continue // skip if no Up section
		}
		_, err = m.conn.DB().Exec(upSQL)
		if err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", mig.name, err)
		}
		_, err = m.conn.DB().Exec(
			"INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, datetime('now'))",
			mig.version, mig.name,
		)
		if err != nil {
			return fmt.Errorf("failed to record migration %s: %w", mig.name, err)
		}
	}
	return nil
}

// extractUpSQL extracts the Up migration SQL from a migration file
func extractUpSQL(content string) string {
	upMarker := "-- +migrate Up"
	downMarker := "-- +migrate Down"
	upIdx := strings.Index(content, upMarker)
	if upIdx == -1 {
		return ""
	}
	upIdx += len(upMarker)
	downIdx := strings.Index(content, downMarker)
	if downIdx == -1 {
		downIdx = len(content)
	}
	return strings.TrimSpace(content[upIdx:downIdx])
}

// Rollback rolls back the last migration.
// TODO: Implement rollback logic.
func (m *Migrator) Rollback() error {
	// Implementation placeholder
	return nil
}

// Version returns the current database schema version.
// TODO: Implement version tracking.
func (m *Migrator) Version() (int, error) {
	// Implementation placeholder
	return 0, nil
}
