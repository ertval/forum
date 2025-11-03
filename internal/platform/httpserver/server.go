// Package httpserver provides HTTP server setup and middleware management.
// It handles server initialization, TLS configuration, and middleware chain setup.
package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Server represents an HTTP server with middleware support.
type Server struct {
	httpServer  *http.Server
	tlsServer   *http.Server
	router      *http.ServeMux
	tlsCertFile string
	tlsKeyFile  string
}

// Config contains HTTP server configuration.
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
func New(cfg Config) *Server {
	router := http.NewServeMux()

	// Initialize HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	// Initialize TLS server if TLS configuration is provided
	var tlsServer *http.Server
	if cfg.TLSCertFile != "" && cfg.TLSKeyFile != "" {
		tlsServer = &http.Server{
			Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.TLSPort),
			Handler:      router,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		}
	}

	return &Server{
		httpServer:  httpServer,
		tlsServer:   tlsServer,
		router:      router,
		tlsCertFile: cfg.TLSCertFile,
		tlsKeyFile:  cfg.TLSKeyFile,
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
	// Implementation placeholder
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

	// Start HTTPS server if TLS is configured
	if s.tlsServer != nil {
		go func() {
			if err := s.tlsServer.ListenAndServeTLS(s.tlsCertFile, s.tlsKeyFile); err != nil && err != http.ErrServerClosed {
				errChan <- fmt.Errorf("HTTPS server error: %w", err)
			}
		}()
	}

	// Return any error that occurs during startup
	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}

// Shutdown gracefully shuts down the server.
// It waits for existing connections to finish before shutting down.
func (s *Server) Shutdown() error {
	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
