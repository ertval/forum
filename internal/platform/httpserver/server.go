// Package httpserver provides HTTP server setup and middleware management.
// It handles server initialization, TLS configuration, and middleware chain setup.
package httpserver

import (
	"net/http"
	"time"
)

// Server represents an HTTP server with middleware support.
type Server struct {
	httpServer *http.Server
	tlsServer  *http.Server
	router     *http.ServeMux
}

// Config contains HTTP and HTTPS server configuration.
type Config struct {
	Host         string
	Port         int
	TLSPort      int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	TLSCertFile  string
	TLSKeyFile   string
}

// New creates a new HTTP server with the specified configuration.
// TODO: Implement server initialization.
func New(cfg Config) *Server {
	return &Server{
		router: http.NewServeMux(),
	}
}

// RegisterHandler registers a handler for a specific path and HTTP method.
// The handler will be wrapped with the middleware chain.
func (s *Server) RegisterHandler(method, path string, handler http.HandlerFunc) {
	// Implementation placeholder
}

// RegisterMiddleware registers global middleware.
// Middleware is executed in the order it is registered.
func (s *Server) RegisterMiddleware(middleware Middleware) {
	// Implementation placeholder
}

// Start starts the HTTP and HTTPS servers.
// Returns an error if the server fails to start.
// TODO: Implement server startup logic.
func (s *Server) Start() error {
	// Implementation placeholder
	// 1. Start HTTP server on Port
	// 2. Start HTTPS server on TLSPort if TLS is configured
	return nil
}

// Shutdown gracefully shuts down the server.
// It waits for existing connections to finish before shutting down.
// TODO: Implement graceful shutdown.
func (s *Server) Shutdown() error {
	// Implementation placeholder
	return nil
}

// Router returns the underlying router for advanced route registration.
func (s *Server) Router() *http.ServeMux {
	return s.router
}
