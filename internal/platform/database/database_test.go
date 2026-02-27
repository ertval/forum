package database

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewConnection(t *testing.T) {
	tests := []struct {
		name    string
		dsn     string
		wantErr bool
		cleanup func()
	}{
		{
			name:    "in-memory database",
			dsn:     ":memory:",
			wantErr: false,
			cleanup: func() {},
		},
		{
			name:    "file database with simple path",
			dsn:     "./test_forum.db",
			wantErr: false,
			cleanup: func() {
				os.Remove("./test_forum.db")
			},
		},
		{
			name:    "file database with nested directory",
			dsn:     "./testdata/nested/test.db",
			wantErr: false,
			cleanup: func() {
				os.RemoveAll("./testdata")
			},
		},
		{
			name:    "file database with URI and params",
			dsn:     "file:./test_uri.db?cache=shared&mode=rwc",
			wantErr: false,
			cleanup: func() {
				os.Remove("./test_uri.db")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer tt.cleanup()

			conn, err := NewConnection(tt.dsn)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConnection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if conn == nil {
					t.Error("NewConnection() returned nil connection")
					return
				}
				if conn.DB() == nil {
					t.Error("NewConnection().DB() returned nil database")
				}

				// Test Ping
				if err := conn.Ping(); err != nil {
					t.Errorf("Connection.Ping() failed: %v", err)
				}

				// Test Close
				if err := conn.Close(); err != nil {
					t.Errorf("Connection.Close() failed: %v", err)
				}

				// Ping should fail after close
				if err := conn.Ping(); err == nil {
					t.Error("Connection.Ping() should fail after Close()")
				}
			}
		})
	}
}

func TestConnection_DirectoryCreation(t *testing.T) {
	testDir := "./testdata/auto_created"
	dbPath := filepath.Join(testDir, "test.db")
	defer os.RemoveAll("./testdata")

	// Ensure directory doesn't exist
	os.RemoveAll("./testdata")

	conn, err := NewConnection(dbPath)
	if err != nil {
		t.Fatalf("NewConnection() failed: %v", err)
	}
	defer conn.Close()

	// Check if directory was created
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Error("Directory was not created")
	}

	// Check if database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}
}

func TestConnection_DB(t *testing.T) {
	conn, err := NewConnection(":memory:")
	if err != nil {
		t.Fatalf("NewConnection() failed: %v", err)
	}
	defer conn.Close()

	db := conn.DB()
	if db == nil {
		t.Fatal("DB() returned nil")
	}

	// Test that we can execute a query
	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Errorf("Failed to execute query on DB(): %v", err)
	}
}

func TestConnection_Close(t *testing.T) {
	conn, err := NewConnection(":memory:")
	if err != nil {
		t.Fatalf("NewConnection() failed: %v", err)
	}

	// First close should succeed
	if err := conn.Close(); err != nil {
		t.Errorf("First Close() failed: %v", err)
	}

	// Second close should not panic and should succeed (no-op)
	if err := conn.Close(); err != nil {
		t.Errorf("Second Close() failed: %v", err)
	}
}

func TestConnection_Ping(t *testing.T) {
	t.Run("successful ping", func(t *testing.T) {
		conn, err := NewConnection(":memory:")
		if err != nil {
			t.Fatalf("NewConnection() failed: %v", err)
		}
		defer conn.Close()

		if err := conn.Ping(); err != nil {
			t.Errorf("Ping() failed: %v", err)
		}
	})

	t.Run("ping after close", func(t *testing.T) {
		conn, err := NewConnection(":memory:")
		if err != nil {
			t.Fatalf("NewConnection() failed: %v", err)
		}

		conn.Close()

		if err := conn.Ping(); err == nil {
			t.Error("Ping() should fail after Close()")
		}
	})
}

// TestStringsIndexByte tests the standard library function we now use
// instead of the custom indexOf function
func TestStringsIndexByte(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		ch       byte
		expected int
	}{
		{
			name:     "character found at start",
			s:        "?param=value",
			ch:       '?',
			expected: 0,
		},
		{
			name:     "character found in middle",
			s:        "file:path?param",
			ch:       '?',
			expected: 9,
		},
		{
			name:     "character not found",
			s:        "nospecialchar",
			ch:       '?',
			expected: -1,
		},
		{
			name:     "empty string",
			s:        "",
			ch:       '?',
			expected: -1,
		},
		{
			name:     "multiple occurrences returns first",
			s:        "a?b?c",
			ch:       '?',
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strings.IndexByte(tt.s, tt.ch)
			if result != tt.expected {
				t.Errorf("strings.IndexByte(%q, %q) = %d, want %d", tt.s, tt.ch, result, tt.expected)
			}
		})
	}
}

// TestConnection_ConcurrentAccess tests thread-safety of Connection
func TestConnection_ConcurrentAccess(t *testing.T) {
	// Use a file-based database with WAL mode for better concurrent support
	dbPath := "./test_concurrent.db"
	defer os.Remove(dbPath)
	defer os.Remove("./test_concurrent.db-wal")
	defer os.Remove("./test_concurrent.db-shm")

	conn, err := NewConnection(dbPath + "?_journal_mode=WAL")
	if err != nil {
		t.Fatalf("NewConnection() failed: %v", err)
	}
	defer conn.Close()

	// Create a table
	_, err = conn.DB().Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, value INTEGER)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Run concurrent operations (reduced for SQLite limitations)
	const numGoroutines = 5
	const numOpsPerGoroutine = 50

	done := make(chan int, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(workerID int) {
			successCount := 0
			for j := 0; j < numOpsPerGoroutine; j++ {
				_, err := conn.DB().Exec("INSERT INTO test (value) VALUES (?)", workerID*1000+j)
				if err != nil {
					// SQLite may lock under heavy concurrent writes, that's expected
					continue
				}
				successCount++
			}
			done <- successCount
		}(i)
	}

	// Wait for all goroutines and count successes
	totalSuccess := 0
	for i := 0; i < numGoroutines; i++ {
		totalSuccess += <-done
	}

	// Verify some rows were inserted (may not be all due to locking)
	var count int
	err = conn.DB().QueryRow("SELECT COUNT(*) FROM test").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count rows: %v", err)
	}

	if count < 1 {
		t.Errorf("Expected at least some rows to be inserted, got %d", count)
	}

	if count != totalSuccess {
		t.Errorf("Mismatch between success count %d and actual rows %d", totalSuccess, count)
	}

	t.Logf("Successfully inserted %d/%d rows with concurrent access", count, numGoroutines*numOpsPerGoroutine)
}

func TestNewMigrator(t *testing.T) {
	conn, err := NewConnection(":memory:")
	if err != nil {
		t.Fatalf("NewConnection() failed: %v", err)
	}
	defer conn.Close()

	migrator := NewMigrator(conn)
	if migrator == nil {
		t.Fatal("NewMigrator() returned nil")
	}

	if migrator.conn == nil {
		t.Error("Migrator.conn is nil")
	}
}

func TestMigrator_Migrate(t *testing.T) {
	// Create a temporary directory for test migrations
	tmpDir := t.TempDir()

	t.Run("successful migration", func(t *testing.T) {
		conn, err := NewConnection(":memory:")
		if err != nil {
			t.Fatalf("NewConnection() failed: %v", err)
		}
		defer conn.Close()

		// Create test migration files
		migration1 := `-- +migrate Up
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    email TEXT NOT NULL
);

-- +migrate Down
DROP TABLE users;
`
		migration2 := `-- +migrate Up
CREATE TABLE posts (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    title TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- +migrate Down
DROP TABLE posts;
`

		if err := os.WriteFile(filepath.Join(tmpDir, "001_create_users.sql"), []byte(migration1), 0644); err != nil {
			t.Fatalf("Failed to write migration file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "002_create_posts.sql"), []byte(migration2), 0644); err != nil {
			t.Fatalf("Failed to write migration file: %v", err)
		}

		migrator := NewMigrator(conn)
		if err := migrator.Migrate(tmpDir); err != nil {
			t.Fatalf("Migrate() failed: %v", err)
		}

		// Verify schema_migrations table exists
		var count int
		err = conn.DB().QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query schema_migrations: %v", err)
		}
		if count != 2 {
			t.Errorf("Expected 2 migrations applied, got %d", count)
		}

		// Verify users table exists
		_, err = conn.DB().Exec("INSERT INTO users (id, username, email) VALUES ('1', 'test', 'test@example.com')")
		if err != nil {
			t.Errorf("Users table not created properly: %v", err)
		}

		// Verify posts table exists
		_, err = conn.DB().Exec("INSERT INTO posts (id, user_id, title) VALUES ('1', '1', 'Test Post')")
		if err != nil {
			t.Errorf("Posts table not created properly: %v", err)
		}
	})

	t.Run("idempotent migrations", func(t *testing.T) {
		conn, err := NewConnection(":memory:")
		if err != nil {
			t.Fatalf("NewConnection() failed: %v", err)
		}
		defer conn.Close()

		migration := `-- +migrate Up
CREATE TABLE test (id INTEGER PRIMARY KEY);

-- +migrate Down
DROP TABLE test;
`
		testMigDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(testMigDir, "001_test.sql"), []byte(migration), 0644); err != nil {
			t.Fatalf("Failed to write migration file: %v", err)
		}

		migrator := NewMigrator(conn)

		// Run migrations first time
		if err := migrator.Migrate(testMigDir); err != nil {
			t.Fatalf("First Migrate() failed: %v", err)
		}

		// Run migrations second time (should be idempotent)
		if err := migrator.Migrate(testMigDir); err != nil {
			t.Fatalf("Second Migrate() failed: %v", err)
		}

		// Verify only one migration record
		var count int
		err = conn.DB().QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query schema_migrations: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected 1 migration record, got %d", count)
		}
	})

	t.Run("migrations applied in order", func(t *testing.T) {
		conn, err := NewConnection(":memory:")
		if err != nil {
			t.Fatalf("NewConnection() failed: %v", err)
		}
		defer conn.Close()

		testMigDir := t.TempDir()

		// Create migrations in reverse order
		for i := 5; i >= 1; i-- {
			content := `-- +migrate Up
CREATE TABLE test` + string(rune('0'+i)) + ` (id INTEGER PRIMARY KEY);

-- +migrate Down
DROP TABLE test` + string(rune('0'+i)) + `;
`
			filename := filepath.Join(testMigDir, "00"+string(rune('0'+i))+"_test.sql")
			if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
				t.Fatalf("Failed to write migration file: %v", err)
			}
		}

		migrator := NewMigrator(conn)
		if err := migrator.Migrate(testMigDir); err != nil {
			t.Fatalf("Migrate() failed: %v", err)
		}

		// Verify migrations were applied in order (1, 2, 3, 4, 5)
		rows, err := conn.DB().Query("SELECT version FROM schema_migrations ORDER BY version")
		if err != nil {
			t.Fatalf("Failed to query migrations: %v", err)
		}
		defer rows.Close()

		expectedVersions := []int{1, 2, 3, 4, 5}
		i := 0
		for rows.Next() {
			var version int
			if err := rows.Scan(&version); err != nil {
				t.Fatalf("Failed to scan version: %v", err)
			}
			if i >= len(expectedVersions) {
				t.Fatal("More migrations than expected")
			}
			if version != expectedVersions[i] {
				t.Errorf("Expected version %d, got %d", expectedVersions[i], version)
			}
			i++
		}
		if i != len(expectedVersions) {
			t.Errorf("Expected %d migrations, got %d", len(expectedVersions), i)
		}
	})

	t.Run("invalid migration directory", func(t *testing.T) {
		conn, err := NewConnection(":memory:")
		if err != nil {
			t.Fatalf("NewConnection() failed: %v", err)
		}
		defer conn.Close()

		migrator := NewMigrator(conn)
		err = migrator.Migrate("/nonexistent/path")
		if err == nil {
			t.Error("Expected error for nonexistent directory")
		}
	})

	t.Run("migration with SQL error", func(t *testing.T) {
		conn, err := NewConnection(":memory:")
		if err != nil {
			t.Fatalf("NewConnection() failed: %v", err)
		}
		defer conn.Close()

		testMigDir := t.TempDir()

		badMigration := `-- +migrate Up
CREATE TABLE users (
    id TEXT PRIMARY KEY
    invalid syntax here
);

-- +migrate Down
DROP TABLE users;
`
		if err := os.WriteFile(filepath.Join(testMigDir, "001_bad.sql"), []byte(badMigration), 0644); err != nil {
			t.Fatalf("Failed to write migration file: %v", err)
		}

		migrator := NewMigrator(conn)
		err = migrator.Migrate(testMigDir)
		if err == nil {
			t.Error("Expected error for invalid SQL")
		}
	})

	t.Run("skip non-migration files", func(t *testing.T) {
		conn, err := NewConnection(":memory:")
		if err != nil {
			t.Fatalf("NewConnection() failed: %v", err)
		}
		defer conn.Close()

		testMigDir := t.TempDir()

		// Create a valid migration
		validMigration := `-- +migrate Up
CREATE TABLE test (id INTEGER PRIMARY KEY);

-- +migrate Down
DROP TABLE test;
`
		if err := os.WriteFile(filepath.Join(testMigDir, "001_test.sql"), []byte(validMigration), 0644); err != nil {
			t.Fatalf("Failed to write README: %v", err)
		}

		// Create files that should be skipped
		if err := os.WriteFile(filepath.Join(testMigDir, "README.md"), []byte("# Migrations"), 0644); err != nil {
			t.Fatalf("Failed to write README: %v", err)
		}
		if err := os.WriteFile(filepath.Join(testMigDir, "invalid_name.sql"), []byte("-- Some SQL"), 0644); err != nil {
			t.Fatalf("Failed to write invalid file: %v", err)
		}
		if err := os.Mkdir(filepath.Join(testMigDir, "subdir"), 0755); err != nil {
			t.Fatalf("Failed to create subdirectory: %v", err)
		}

		migrator := NewMigrator(conn)
		if err := migrator.Migrate(testMigDir); err != nil {
			t.Fatalf("Migrate() failed: %v", err)
		}

		// Verify only one migration was applied
		var count int
		err = conn.DB().QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query schema_migrations: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected 1 migration applied, got %d", count)
		}
	})
}

func TestExtractUpSQL(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name: "standard migration",
			content: `-- +migrate Up
CREATE TABLE test (id INTEGER);

-- +migrate Down
DROP TABLE test;`,
			expected: "CREATE TABLE test (id INTEGER);",
		},
		{
			name: "multi-line SQL",
			content: `-- +migrate Up
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    email TEXT NOT NULL
);

CREATE INDEX idx_users_email ON users(email);

-- +migrate Down
DROP TABLE users;`,
			expected: `CREATE TABLE users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    email TEXT NOT NULL
);

CREATE INDEX idx_users_email ON users(email);`,
		},
		{
			name: "no up section",
			content: `-- Some comments
CREATE TABLE test (id INTEGER);`,
			expected: "",
		},
		{
			name: "up only, no down",
			content: `-- +migrate Up
CREATE TABLE test (id INTEGER);`,
			expected: "CREATE TABLE test (id INTEGER);",
		},
		{
			name: "empty up section",
			content: `-- +migrate Up

-- +migrate Down
DROP TABLE test;`,
			expected: "",
		},
		{
			name: "up section with comments",
			content: `-- +migrate Up
-- Create the users table
CREATE TABLE users (id INTEGER);
-- Add index
CREATE INDEX idx_users ON users(id);

-- +migrate Down
DROP TABLE users;`,
			expected: `-- Create the users table
CREATE TABLE users (id INTEGER);
-- Add index
CREATE INDEX idx_users ON users(id);`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractUpSQL(tt.content)
			if result != tt.expected {
				t.Errorf("extractUpSQL() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestMigrator_Rollback(t *testing.T) {
	conn, err := NewConnection(":memory:")
	if err != nil {
		t.Fatalf("NewConnection() failed: %v", err)
	}
	defer conn.Close()

	migrator := NewMigrator(conn)

	// Rollback now returns an error indicating it's not implemented
	err = migrator.Rollback()
	if err == nil {
		t.Error("Rollback() should return an error (not yet implemented)")
	}
}

func TestMigrator_Version(t *testing.T) {
	conn, err := NewConnection(":memory:")
	if err != nil {
		t.Fatalf("NewConnection() failed: %v", err)
	}
	defer conn.Close()

	migrator := NewMigrator(conn)

	// First run Migrate to create the schema_migrations table
	tmpDir := t.TempDir()
	if err := migrator.Migrate(tmpDir); err != nil {
		t.Fatalf("Migrate() failed: %v", err)
	}

	// Version should return 0 when no migrations are applied
	version, err := migrator.Version()
	if err != nil {
		t.Errorf("Version() error = %v, expected nil", err)
	}
	if version != 0 {
		t.Errorf("Version() = %d, expected 0 (no migrations applied)", version)
	}
}

// TestMigrator_WithRealMigrations tests the migrator with actual project migration files
func TestMigrator_WithRealMigrations(t *testing.T) {
	// Skip if migrations directory doesn't exist
	migrationsPath := "../../../../migrations"
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		t.Skip("Migrations directory not found")
	}

	conn, err := NewConnection(":memory:")
	if err != nil {
		t.Fatalf("NewConnection() failed: %v", err)
	}
	defer conn.Close()

	migrator := NewMigrator(conn)
	if err := migrator.Migrate(migrationsPath); err != nil {
		t.Fatalf("Migrate() with real migrations failed: %v", err)
	}

	// Verify at least one migration was applied
	var count int
	err = conn.DB().QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query schema_migrations: %v", err)
	}
	if count == 0 {
		t.Error("Expected at least one migration to be applied")
	}

	t.Logf("Successfully applied %d migrations", count)
}

// Note: transaction-related tests were moved to `transaction_test.go` to avoid
// duplicate declarations across test files. The transaction test implementations
// live in that file; keep this placeholder comment here to explain the split.
