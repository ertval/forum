package main

// main.go is the entry point of the forum application.
// It initializes the database, sets up HTTP routes and middleware,
// and starts the web server.

import (
	"log"
	"net/http"
)

func main() {
	// Initialize database connection
	// Setup routes and middleware
	// Start HTTP server on port 8080

	log.Println("Starting forum server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
