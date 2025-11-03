package errors
// Package errors provides common error types and utilities.
// It defines domain-agnostic errors that can be used across all modules.
package errors

import (
	"errors"
	"fmt"
)

// Error represents an application error with additional context.
type Error struct {
	Code    string // Error code for programmatic handling
	Message string // Human-readable error message
	Err     error  // Underlying error
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

// Common error codes
const (
	CodeNotFound          = "NOT_FOUND"
	CodeAlreadyExists     = "ALREADY_EXISTS"
	CodeInvalidInput      = "INVALID_INPUT"
	CodeUnauthorized      = "UNAUTHORIZED"
	CodeForbidden         = "FORBIDDEN"
	CodeInternal          = "INTERNAL"
	CodeConflict          = "CONFLICT"
	CodeTooManyRequests   = "TOO_MANY_REQUESTS"
	CodeBadRequest        = "BAD_REQUEST"
	CodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)

// New creates a new error with the given code and message.
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

// Is checks if the error matches the given code.
func Is(err error, code string) bool {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}

// Common error constructors
func NotFound(message string) *Error {
	return New(CodeNotFound, message)
}

func AlreadyExists(message string) *Error {
	return New(CodeAlreadyExists, message)
}

func InvalidInput(message string) *Error {
	return New(CodeInvalidInput, message)
}

func Unauthorized(message string) *Error {
	return New(CodeUnauthorized, message)
}

func Forbidden(message string) *Error {
	return New(CodeForbidden, message)
}

func Internal(message string) *Error {
	return New(CodeInternal, message)
}

func Conflict(message string) *Error {
	return New(CodeConflict, message)
}
