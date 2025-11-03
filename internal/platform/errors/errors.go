// Package errors provides common error types and error handling utilities.
// It defines domain errors and provides error wrapping and conversion functions.
package errors

import (
	"fmt"
	"net/http"
)

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

// HTTPStatus maps error codes to HTTP status codes.
func HTTPStatus(err error) int {
	if e, ok := err.(*Error); ok {
		switch e.Code {
		case ErrCodeValidation, ErrCodeBadRequest:
			return http.StatusBadRequest
		case ErrCodeUnauthorized:
			return http.StatusUnauthorized
		case ErrCodeForbidden:
			return http.StatusForbidden
		case ErrCodeNotFound:
			return http.StatusNotFound
		case ErrCodeConflict:
			return http.StatusConflict
		case ErrCodeTooManyRequests:
			return http.StatusTooManyRequests
		case ErrCodeInternal:
			return http.StatusInternalServerError
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
