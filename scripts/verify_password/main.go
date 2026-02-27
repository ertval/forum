package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

const defaultDBPath = "./data/forum.db"

func main() {
	dbPath := flag.String("dbpath", defaultDBPath, "Path to SQLite database")
	useDB := flag.Bool("db", false, "Verify password against user in database")
	generate := flag.Bool("generate", false, "Generate a bcrypt hash")
	cost := flag.Int("cost", bcrypt.DefaultCost, "Cost factor for bcrypt")
	flag.Parse()

	args := flag.Args()

	if *generate {
		if len(args) != 1 {
			fmt.Fprintf(os.Stderr, "Usage: %s -generate <password>\n", os.Args[0])
			os.Exit(1)
		}
		generateHash(args[0], *cost)
		return
	}

	if *useDB {
		if len(args) != 2 {
			fmt.Fprintf(os.Stderr, "Usage: %s -db <email> <password>\n", os.Args[0])
			os.Exit(1)
		}
		verifyFromDB(*dbPath, args[0], args[1])
		return
	}

	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <password> <hash>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "       %s -db <email> <password>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "       %s -generate <password>\n", os.Args[0])
		os.Exit(1)
	}

	verifyPassword(args[0], args[1])
}

func generateHash(password string, cost int) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating hash: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Password: %s\nHash: %s\nCost: %d\n", password, string(hash), cost)
}

func verifyPassword(password, hash string) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			fmt.Println("Password does NOT match")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}
	fmt.Println("Password matches!")
	if c, err := bcrypt.Cost([]byte(hash)); err == nil {
		fmt.Printf("Cost factor: %d\n", c)
	}
}

func verifyFromDB(dbPath, email, password string) {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Database not found: %s\n", dbPath)
		os.Exit(1)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	var username, passwordHash string
	var userID int
	query := `SELECT id, username, password_hash FROM users WHERE email = ?`
	err = db.QueryRow(query, email).Scan(&userID, &username, &passwordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Fprintf(os.Stderr, "User not found: %s\n", email)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("User: %s (ID: %d, Email: %s)\n", username, userID, email)

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			fmt.Println("Password does NOT match")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}
	fmt.Println("Password matches!")
	if c, err := bcrypt.Cost([]byte(passwordHash)); err == nil {
		fmt.Printf("Cost factor: %d\n", c)
	}
}
