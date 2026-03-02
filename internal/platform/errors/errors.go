// Package errors provides error handling utilities for HTTP handlers.
// It provides JSON error responses and HTML error page rendering.
package errors

import (
	"encoding/json"
	"net/http"
	"os"

	"forum/internal/platform/logger"
)

// Package-level error logger (created once, not on every error)
var errLogger = logger.NewWithConfig(logger.ErrorLevel, os.Stderr, &logger.Config{
	TimePrecision: logger.TimePrecisionSeconds,
	AllowedFields: []string{"status", "error"},
	MaxLineWidth:  200,
	Colorize:      true,
})

// WriteErrorJSON writes a JSON error response with appropriate status code and logging.
// This is the standard way to return errors from HTTP handlers.
// It automatically logs errors to stderr for debugging and monitoring.
func WriteErrorJSON(w http.ResponseWriter, status int, message string) {
	// Log error for debugging (using package-level logger)
	errLogger.Error("http.error",
		logger.Int("status", status),
		logger.String("error", message))

	// Create error response
	errResp := struct {
		Error string `json:"error"`
	}{
		Error: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(errResp); err != nil {
		// JSON encoding failed - log and send plain text fallback
		errLogger.Error("failed to encode error response", logger.Error(err))
		// Best-effort fallback write for clients when JSON encoding fails.
		// Status/header may already be written by the server at this point.
		_, _ = w.Write([]byte(message + "\n"))
	}
}
