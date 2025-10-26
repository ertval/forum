package server

// server.go handles HTTP server initialization and configuration.
// It provides a clean API for setting up and running the forum server.

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"forum/internal/database"
)

// Server represents the HTTP server and its dependencies
type Server struct {
	httpServer *http.Server
	router     http.Handler
}

// Config holds server configuration
type Config struct {
	Port   string
	DBPath string
}

// New creates a new Server instance with the given configuration
func New(cfg Config) (*Server, error) {
	// Initialize database
	if err := database.InitDB(cfg.DBPath); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Setup router with all routes and middleware
	router := setupRouter()

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		httpServer: srv,
		router:     router,
	}, nil
}

// Start begins listening for HTTP requests
func (s *Server) Start() error {
	log.Printf("🚀 Forum server is running at http://localhost%s 🌐", s.httpServer.Addr)

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed to start: %w", err)
	}
	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down server gracefully...")

	// Shutdown HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	// Close database connection
	if err := database.CloseDB(); err != nil {
		return fmt.Errorf("database close failed: %w", err)
	}

	log.Println("Server stopped")
	return nil
}

// Run starts the server and handles graceful shutdown
func (s *Server) Run() error {
	// Channel to listen for interrupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- s.Start()
	}()

	// Block until we receive a signal or an error
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case <-stop:
		// Create context with timeout for shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		return s.Shutdown(ctx)
	}
}

// DefaultConfig returns the default server configuration
func DefaultConfig() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "forum.db"
	}

	return Config{
		Port:   port,
		DBPath: dbPath,
	}
}
