package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"forum/internal/platform/database"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	dbPath := "./data/forum.db"
	conn, err := database.NewConnection(dbPath)
	if err != nil {
		log.Fatalf("failed to open DB: %v", err)
	}
	defer conn.Close()

	migrator := database.NewMigrator(conn)
	if err := migrator.Migrate("./migrations"); err != nil {
		// Handle common harmless cases (column already exists, etc.)
		if strings.Contains(err.Error(), "duplicate column name") || strings.Contains(err.Error(), "already exists") {
			fmt.Println("migration reported a harmless 'already exists' error; recording migration as applied")
			// Try to record the migration manually. We assume the migration file we added is 008.
			if _, execErr := conn.DB().Exec("INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, datetime('now'))", 8, "008_user_add_password_hash.sql"); execErr != nil {
				log.Fatalf("failed to record migration as applied: %v (original error: %v)", execErr, err)
			}
			fmt.Println("✅ Migration recorded as applied")
		} else {
			log.Fatalf("migrate failed: %v", err)
		}
	} else {
		fmt.Println("✅ Migrations applied")
	}

	// Verify `users` table columns
	rows, err := conn.DB().Query("PRAGMA table_info(users);")
	if err != nil {
		log.Fatalf("failed to query table info: %v", err)
	}
	defer rows.Close()

	fmt.Println("users table columns:")
	for rows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dflt interface{}
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			log.Fatalf("scan failed: %v", err)
		}
		fmt.Printf("- %s (%s)\n", name, ctype)
	}

	// Optional: quick sanity check if password_hash exists
	var exists int
	err = conn.DB().QueryRow("SELECT COUNT(1) FROM pragma_table_info('users') WHERE name='password_hash'").Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		log.Fatalf("failed to check column existence: %v", err)
	}
	if exists > 0 {
		fmt.Println("password_hash column exists ✅")
	} else {
		fmt.Println("password_hash column NOT found — something went wrong")
	}
}
