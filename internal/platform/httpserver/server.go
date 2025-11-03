package httpserver
// Package httpserver provides HTTP server setup and configuration.
// It handles server lifecycle, graceful shutdown, TLS configuration,
// and common server settings.
package httpserver

import (
	"net/http"
)

// Server wraps http.Server with additional functionality.
type Server struct {
	httpServer  *http.Server
	httpsServer *http.Server
	// TODO: Add server configuration fields
}

// Config holds HTTP server configuration.
type Config struct {
	Port         string
	TLSPort      string
	ReadTimeout  int
	WriteTimeout int
	IdleTimeout  int
	CertFile     string
	KeyFile      string
	// TODO: Add more configuration fields
}

// New creates a new HTTP server instance.
func New(config Config, handler http.Handler) *Server {
	// TODO: Implement server creation
	return &Server{}
}

// Start starts the HTTP and HTTPS servers.
func (s *Server) Start() error {
	// TODO: Implement server start
	return nil
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown() error {
	// TODO: Implement graceful shutdown
	return nil
}
