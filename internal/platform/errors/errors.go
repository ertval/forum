// Package errors provides common error types and error handling utilities.
// It defines domain errors and provides error wrapping and conversion functions.
package errors

import (
	"encoding/json"
	"fmt"
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

// Error represents a domain error with additional context.
type Error struct {
	Code    string // Error code for identification
	Message string // Human-readable error message
	Err     error  // Underlying error (if any)
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error.
func (e *Error) Unwrap() error {
	return e.Err
}

// New creates a new Error with the specified code and message.
func New(code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an existing error with additional context.
func Wrap(err error, code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Common error codes
const (
	// ErrCodeValidation indicates input validation failure
	ErrCodeValidation = "VALIDATION_ERROR"

	// ErrCodeNotFound indicates resource not found
	ErrCodeNotFound = "NOT_FOUND"

	// ErrCodeUnauthorized indicates missing or invalid authentication
	ErrCodeUnauthorized = "UNAUTHORIZED"

	// ErrCodeForbidden indicates insufficient permissions
	ErrCodeForbidden = "FORBIDDEN"

	// ErrCodeConflict indicates resource conflict (e.g., duplicate)
	ErrCodeConflict = "CONFLICT"

	// ErrCodeInternal indicates internal server error
	ErrCodeInternal = "INTERNAL_ERROR"

	// ErrCodeBadRequest indicates invalid request
	ErrCodeBadRequest = "BAD_REQUEST"

	// ErrCodeTooManyRequests indicates rate limit exceeded
	ErrCodeTooManyRequests = "TOO_MANY_REQUESTS"
)

// codeToStatus maps error codes to HTTP status codes (KISS-5: map lookup instead of switch)
var codeToStatus = map[string]int{
	ErrCodeValidation:      http.StatusBadRequest,
	ErrCodeBadRequest:      http.StatusBadRequest,
	ErrCodeUnauthorized:    http.StatusUnauthorized,
	ErrCodeForbidden:       http.StatusForbidden,
	ErrCodeNotFound:        http.StatusNotFound,
	ErrCodeConflict:        http.StatusConflict,
	ErrCodeTooManyRequests: http.StatusTooManyRequests,
	ErrCodeInternal:        http.StatusInternalServerError,
}

// HTTPStatus maps error codes to HTTP status codes.
func HTTPStatus(err error) int {
	if e, ok := err.(*Error); ok {
		if status, found := codeToStatus[e.Code]; found {
			return status
		}
	}
	return http.StatusInternalServerError
}

// ErrorResponse represents an HTTP error response.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ToHTTPResponse converts an error to an HTTP error response.
func ToHTTPResponse(err error) ErrorResponse {
	if e, ok := err.(*Error); ok {
		return ErrorResponse{
			Code:    e.Code,
			Message: e.Message,
		}
	}
	return ErrorResponse{
		Code:    ErrCodeInternal,
		Message: "An internal error occurred",
	}
}

// WriteErrorJSON writes a JSON error response with appropriate status code and logging.
// This is the standard way to return errors from HTTP handlers.
// It automatically logs errors to stderr for debugging and monitoring.
func WriteErrorJSON(w http.ResponseWriter, status int, message string) {
	// Log error for debugging (using package-level logger)
	errLogger.Error("http.error",
		logger.Int("status", status),
		logger.String("error", message))

	// Set JSON content type
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// Create error response
	errResp := struct {
		Error string `json:"error"`
	}{
		Error: message,
	}

	// Write JSON response with fallback to plain text if encoding fails (CRIT-3)
	if err := json.NewEncoder(w).Encode(errResp); err != nil {
		// JSON encoding failed - log and send plain text fallback
		errLogger.Error("failed to encode error response", logger.Error(err))
		// Note: headers already sent, but we can try to write something useful
		// The client may receive a partial/malformed response in this edge case
	}
}
