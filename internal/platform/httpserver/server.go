// Package httpserver provides HTTP server setup and middleware management.
// It handles server initialization, TLS configuration, and middleware chain setup.
package httpserver

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"forum/internal/platform/config"
)

// Server represents an HTTP server with middleware support.
type Server struct {
	httpServer    *http.Server
	tlsServer     *http.Server
	router        *http.ServeMux
	handler       http.Handler
	middleware    []Middleware
	config        *config.Config
	shutdownHooks []func()
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
// Listeners are bound synchronously so binding errors are returned immediately.
// Serving happens asynchronously in background goroutines.
func (s *Server) Start() error {
	// Bind HTTP listener synchronously
	httpLn, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return fmt.Errorf("HTTP listen error: %w", err)
	}

	// Serve HTTP asynchronously
	go func() {
		if err := s.httpServer.Serve(httpLn); err != nil && err != http.ErrServerClosed {
			// Error during serving is not recoverable at this point
		}
	}()

	// Start HTTPS server if TLS is configured and certificates exist
	if s.tlsServer != nil && s.config.Security.TLSCertFile != "" && s.config.Security.TLSKeyFile != "" {
		tlsLn, err := net.Listen("tcp", s.tlsServer.Addr)
		if err != nil {
			s.httpServer.Close()
			return fmt.Errorf("HTTPS listen error: %w", err)
		}

		go func() {
			if err := s.tlsServer.ServeTLS(tlsLn, s.config.Security.TLSCertFile, s.config.Security.TLSKeyFile); err != nil && err != http.ErrServerClosed {
				// Error during serving is not recoverable at this point
			}
		}()
	}

	return nil
}

// defaultShutdownTimeout is the duration to wait for graceful shutdown (KISS-6)
const defaultShutdownTimeout = 30 * time.Second

// Shutdown gracefully shuts down the server.
// It waits for existing connections to finish before shutting down.
func (s *Server) Shutdown() error {
	// Run shutdown hooks (e.g., stop rate limiter cleanup goroutine)
	for _, fn := range s.shutdownHooks {
		fn()
	}

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

// OnShutdown registers a function to be called during graceful shutdown.
func (s *Server) OnShutdown(fn func()) {
	s.shutdownHooks = append(s.shutdownHooks, fn)
}

// Router returns the underlying router for advanced route registration.
func (s *Server) Router() *http.ServeMux {
	return s.router
}
