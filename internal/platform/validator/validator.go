package validator
// Package validator provides input validation utilities.
// It offers common validation functions for emails, passwords, usernames, etc.
package validator

import (
	"regexp"
	"strings"
	"unicode"
)

// ValidationError represents a validation error with field-specific messages.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (v ValidationError) Error() string {
	return v.Field + ": " + v.Message
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []ValidationError

// Error implements the error interface.
func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, err := range v {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(err.Error())
	}
	return sb.String()
}

// Add adds a validation error to the collection.
func (v *ValidationErrors) Add(field, message string) {
	*v = append(*v, ValidationError{Field: field, Message: message})
}

// HasErrors returns true if there are any validation errors.
func (v ValidationErrors) HasErrors() bool {
	return len(v) > 0
}

// Email validation
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// IsValidEmail checks if the email is valid.
func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// IsValidPassword checks if the password meets security requirements.
// Password must be at least 8 characters long and contain at least one uppercase,
// one lowercase, one digit, and one special character.
func IsValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}

// IsValidUsername checks if the username meets requirements.
// Username must be 3-30 characters and contain only alphanumeric characters and underscores.
func IsValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 30 {
		return false
	}
	return regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(username)
}

// Required checks if a value is not empty.
func Required(value string) bool {
	return strings.TrimSpace(value) != ""
}

// MaxLength checks if a string does not exceed the maximum length.
func MaxLength(value string, max int) bool {
	return len(value) <= max
}

// MinLength checks if a string meets the minimum length.
func MinLength(value string, min int) bool {
	return len(value) >= min
}
