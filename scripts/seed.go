package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"

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

	// Read seed data file
	seedFile := "./scripts/seed_data.sql"
	if _, err := os.Stat(seedFile); os.IsNotExist(err) {
		log.Fatalf("Seed file does not exist: %s", seedFile)
	}

	content, err := ioutil.ReadFile(seedFile)
	if err != nil {
		log.Fatalf("Failed to read seed file: %v", err)
	}

	// Execute seed data
	fmt.Println("Seeding database with mock data...")
	_, err = db.Exec(string(content))
	if err != nil {
		log.Fatalf("Failed to execute seed data: %v", err)
	}

	fmt.Println("✅ Database seeded successfully!")
	fmt.Println("Mock data includes:")
	fmt.Println("- 3 test users (alice, bob, charlie)")
	fmt.Println("- 5 categories (General, Technology, Gaming, Science, Entertainment)")
	fmt.Println("- 6 sample posts")
	fmt.Println("- 5 sample comments")
	fmt.Println("- Various reactions (likes/dislikes)")
}
