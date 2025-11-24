package httpserver

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"forum/internal/platform/config"
	"forum/internal/platform/logger"
)

// TestServerWithMiddleware tests that server properly registers and applies middleware
func TestServerWithMiddleware(t *testing.T) {
	// Create config
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
	}

	// Create logger that writes to buffer
	buf := &bytes.Buffer{}
	lgr := logger.New(logger.InfoLevel, buf)

	// Create server
	srv := New(cfg)

	// Register middleware
	srv.RegisterMiddleware(Recovery(lgr))
	srv.RegisterMiddleware(Logger(lgr))

	// Register test handler
	srv.Router().HandleFunc("GET /test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	// Serve request through the server's handler (which includes middleware)
	srv.handler.ServeHTTP(rec, req)

	// Verify response
	if rec.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusOK)
	}

	// Verify logging middleware captured the request
	logOutput := buf.String()
	if !strings.Contains(logOutput, "http.request") {
		t.Errorf("log output missing http.request\nLog: %s", logOutput)
	}

	if !strings.Contains(logOutput, "GET") {
		t.Errorf("log output missing GET method\nLog: %s", logOutput)
	}

	if !strings.Contains(logOutput, "/test") {
		t.Errorf("log output missing /test path\nLog: %s", logOutput)
	}
}

// TestServerRecoveryMiddleware tests that panic recovery works in the server
func TestServerRecoveryMiddleware(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
	}

	buf := &bytes.Buffer{}
	lgr := logger.New(logger.InfoLevel, buf)

	srv := New(cfg)
	srv.RegisterMiddleware(Recovery(lgr))
	srv.RegisterMiddleware(Logger(lgr))

	// Register handler that panics
	srv.Router().HandleFunc("GET /panic", func(w http.ResponseWriter, r *http.Request) {
		panic("intentional panic")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	rec := httptest.NewRecorder()

	srv.handler.ServeHTTP(rec, req)

	// Should return 500
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	// Should log panic
	logOutput := buf.String()
	if !strings.Contains(logOutput, "panic.recovered") {
		t.Errorf("log output missing panic.recovered\nLog: %s", logOutput)
	}

	if !strings.Contains(logOutput, "intentional panic") {
		t.Errorf("log output missing panic message\nLog: %s", logOutput)
	}
}

// TestServerMiddlewareOrder tests that middleware executes in registration order
func TestServerMiddlewareOrder(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
	}

	var order []string

	m1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "m1-before")
			next.ServeHTTP(w, r)
			order = append(order, "m1-after")
		})
	}

	m2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "m2-before")
			next.ServeHTTP(w, r)
			order = append(order, "m2-after")
		})
	}

	srv := New(cfg)
	srv.RegisterMiddleware(m1)
	srv.RegisterMiddleware(m2)

	srv.Router().HandleFunc("GET /test", func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	srv.handler.ServeHTTP(rec, req)

	expectedOrder := []string{
		"m1-before", "m2-before",
		"handler",
		"m2-after", "m1-after",
	}

	if len(order) != len(expectedOrder) {
		t.Fatalf("order length = %d, want %d\nGot: %v", len(order), len(expectedOrder), order)
	}

	for i, expected := range expectedOrder {
		if order[i] != expected {
			t.Errorf("order[%d] = %q, want %q", i, order[i], expected)
		}
	}
}
