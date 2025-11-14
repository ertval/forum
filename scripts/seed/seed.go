package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Database path (same as in config)
	dbPath := "./data/forum.db"

	// Check if database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Fatalf("Database file does not exist: %s. Please run the application first to create the database.", dbPath)
	}

	// Open database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Check if tables exist
	tables := []string{"users", "categories", "posts", "post_categories", "reactions", "comments"}
	for _, table := range tables {
		if !tableExists(db, table) {
			log.Fatalf("Table %s does not exist. Please run migrations first.", table)
		}
	}

	// Read seed data file
	seedFile := "./scripts/seed_data.sql"
	if _, err := os.Stat(seedFile); os.IsNotExist(err) {
		log.Fatalf("Seed file does not exist: %s", seedFile)
	}

	content, err := os.ReadFile(seedFile)
	if err != nil {
		log.Fatalf("Failed to read seed file: %v", err)
	}

	// Execute the seed data
	fmt.Println("Seeding database with mock data...")

	// Split the content by semicolon to execute multiple statements
	commands := strings.Split(string(content), ";")

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback() // Rollback if not committed

	for _, cmd := range commands {
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			continue
		}

		// Skip comment lines that start with SELECT (these are for verification)
		if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(cmd)), "SELECT ") {
			continue
		}

		_, err := tx.Exec(cmd)
		if err != nil {
			log.Printf("Error executing command: %s\nError: %v", cmd, err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	fmt.Println("✅ Database seeded successfully!")
	fmt.Println("Mock data includes:")
	fmt.Println("- 3 test users (alice, bob, charlie)")
	fmt.Println("- 5 categories (General, Technology, Gaming, Science, Entertainment)")
	fmt.Println("- 6 sample posts")
	fmt.Println("- 5 sample comments")
	fmt.Println("- Various reactions (likes/dislikes)")
}

// tableExists checks if a table exists in the database
func tableExists(db *sql.DB, tableName string) bool {
	query := `SELECT name FROM sqlite_master WHERE type='table' AND name=?;`
	var name string
	err := db.QueryRow(query, tableName).Scan(&name)
	return err == nil
}
