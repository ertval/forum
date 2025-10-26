package main

// main.go is the entry point of the forum application.
// It orchestrates server initialization and startup.

import (
	"log"

	"forum/internal/server"
)

func main() {
	// Load configuration from environment or use defaults
	cfg := server.DefaultConfig()

	// Create new server instance
	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Run server with graceful shutdown handling
	if err := srv.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
