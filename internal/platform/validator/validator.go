// Package validator provides input validation and sanitization utilities.
// It validates user input according to business rules and security requirements.
package validator

import (
	"regexp"
	"strings"
)

// Validator provides validation methods for common data types.
type Validator struct {
	errors map[string]string
}

// New creates a new Validator instance.
func New() *Validator {
	return &Validator{
		errors: make(map[string]string),
	}
}

// Valid returns true if there are no validation errors.
func (v *Validator) Valid() bool {
	return len(v.errors) == 0
}

// AddError adds a validation error for a field.
func (v *Validator) AddError(field, message string) {
	if _, exists := v.errors[field]; !exists {
		v.errors[field] = message
	}
}

// Errors returns all validation errors.
func (v *Validator) Errors() map[string]string {
	return v.errors
}

// Required checks if a value is not empty.
func (v *Validator) Required(field, value string) {
	if strings.TrimSpace(value) == "" {
		v.AddError(field, "This field is required")
	}
}

// MinLength checks if a string has minimum length.
func (v *Validator) MinLength(field, value string, min int) {
	if len(value) < min {
		v.AddError(field, "Must be at least "+string(rune(min))+" characters")
	}
}

// MaxLength checks if a string has maximum length.
func (v *Validator) MaxLength(field, value string, max int) {
	if len(value) > max {
		v.AddError(field, "Must be at most "+string(rune(max))+" characters")
	}
}

// Email validates an email address format.
// TODO: Implement proper email validation.
func (v *Validator) Email(field, value string) {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(value) {
		v.AddError(field, "Must be a valid email address")
	}
}

// Username validates a username format.
// Usernames should contain only letters, numbers, and underscores.
func (v *Validator) Username(field, value string) {
	if len(value) < 3 || len(value) > 20 {
		v.AddError(field, "Username must be between 3 and 20 characters")
		return
	}

	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !usernameRegex.MatchString(value) {
		v.AddError(field, "Username can only contain letters, numbers, and underscores")
	}
}

// Password validates password strength.
// TODO: Implement password strength requirements.
func (v *Validator) Password(field, value string, minLength int) {
	if len(value) < minLength {
		v.AddError(field, "Password must be at least "+string(rune(minLength))+" characters")
	}
	// Additional password requirements can be added here
	// e.g., must contain uppercase, lowercase, numbers, special characters
}

// In checks if a value is in a list of allowed values.
func (v *Validator) In(field, value string, allowed []string) {
	for _, a := range allowed {
		if value == a {
			return
		}
	}
	v.AddError(field, "Invalid value")
}

// Matches checks if a value matches a regular expression.
func (v *Validator) Matches(field, value string, pattern *regexp.Regexp) {
	if !pattern.MatchString(value) {
		v.AddError(field, "Invalid format")
	}
}

// Sanitize removes potentially dangerous characters from input.
// TODO: Implement sanitization logic.
func Sanitize(input string) string {
	// Implementation placeholder
	// Remove HTML tags, scripts, etc.
	return strings.TrimSpace(input)
}

// SanitizeHTML sanitizes HTML input to prevent XSS attacks.
// TODO: Implement HTML sanitization.
func SanitizeHTML(input string) string {
	// Implementation placeholder
	// Use an HTML sanitization library or implement custom logic
	return input
}
