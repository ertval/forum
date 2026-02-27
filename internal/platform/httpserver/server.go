// Package httpserver provides HTTP server setup and middleware management.
// It handles server initialization, TLS configuration, and middleware chain setup.
package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"forum/internal/platform/config"
)

// Server represents an HTTP server with middleware support.
type Server struct {
	httpServer *http.Server
	tlsServer  *http.Server
	router     *http.ServeMux
	handler    http.Handler
	middleware []Middleware
	config     *config.Config
}

// New creates a new HTTP server with the specified config.
// It uses the ServerConfig substruct directly as the single source of truth.
func New(cfg *config.Config) *Server {
	router := http.NewServeMux()

	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	var tlsServer *http.Server
	if cfg.Security.TLSCertFile != "" && cfg.Security.TLSKeyFile != "" {
		// Check if certificate files exist before configuring TLS
		if _, err := os.Stat(cfg.Security.TLSCertFile); err == nil {
			if _, err := os.Stat(cfg.Security.TLSKeyFile); err == nil {
				tlsServer = &http.Server{
					Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.TLSPort),
					Handler:      router,
					ReadTimeout:  cfg.Server.ReadTimeout,
					WriteTimeout: cfg.Server.WriteTimeout,
					IdleTimeout:  cfg.Server.IdleTimeout,
					TLSConfig:    TLSConfig(), // Use secure TLS configuration with cipher suites
				}
			}
		}
	}

	return &Server{
		httpServer: httpServer,
		tlsServer:  tlsServer,
		router:     router,
		handler:    router,
		middleware: []Middleware{},
		config:     cfg,
	}
}

// RegisterHandler registers a handler for a specific path and HTTP method.
// The handler will be wrapped with the middleware chain.
func (s *Server) RegisterHandler(method, path string, handler http.HandlerFunc) {
	wrappedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request method matches
		if r.Method != method {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler(w, r)
	})
	s.router.Handle(path, wrappedHandler)
}

// RegisterMiddleware registers global middleware.
// Middleware is executed in the order it is registered.
func (s *Server) RegisterMiddleware(middleware Middleware) {
	s.middleware = append(s.middleware, middleware)
	// Rebuild the handler chain with all registered middleware
	s.handler = Chain(s.middleware...)(s.router)
	s.httpServer.Handler = s.handler
	if s.tlsServer != nil {
		s.tlsServer.Handler = s.handler
	}
}

// Start starts the HTTP and HTTPS servers.
// Returns an error if the server fails to start.
func (s *Server) Start() error {
	errChan := make(chan error, 2)

	// Start HTTP server
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	// Start HTTPS server if TLS is configured and certificates exist
	if s.tlsServer != nil && s.config.Security.TLSCertFile != "" && s.config.Security.TLSKeyFile != "" {
		go func() {
			if err := s.tlsServer.ListenAndServeTLS(s.config.Security.TLSCertFile, s.config.Security.TLSKeyFile); err != nil && err != http.ErrServerClosed {
				errChan <- fmt.Errorf("HTTPS server error: %w", err)
			}
		}()
	}

	// Wait for any error or return nil if both servers start
	select {
	case err := <-errChan:
		return err
	case <-time.After(100 * time.Millisecond):
		return nil
	}
}

// defaultShutdownTimeout is the duration to wait for graceful shutdown (KISS-6)
const defaultShutdownTimeout = 30 * time.Second

// Shutdown gracefully shuts down the server.
// It waits for existing connections to finish before shutting down.
func (s *Server) Shutdown() error {
	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer cancel()

	// Shutdown HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("HTTP server shutdown error: %w", err)
	}

	// Shutdown HTTPS server if it exists
	if s.tlsServer != nil {
		if err := s.tlsServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("HTTPS server shutdown error: %w", err)
		}
	}

	return nil
}

// Router returns the underlying router for advanced route registration.
func (s *Server) Router() *http.ServeMux {
	return s.router
}
