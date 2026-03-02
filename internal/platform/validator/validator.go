// Package validator provides input validation and sanitization utilities.
// It validates user input according to business rules and security requirements.
package validator

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Pre-compiled regexes for performance (compiled once at package init)
var (
	// emailRegex validates email format
	emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	// namePartRegex validates username handles (letters, digits, underscores, hyphens)
	namePartRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)
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
// Callers should sanitize the value once before calling validation rules;
// individual rule methods do not re-sanitize.
func (v *Validator) Required(field, value string) {
	if strings.TrimSpace(value) == "" {
		v.AddError(field, "This field is required")
	}
}

// MinLength checks if a string has minimum length.
func (v *Validator) MinLength(field, value string, min int) {
	if utf8.RuneCountInString(value) < min {
		v.AddError(field, "Must be at least "+strconv.Itoa(min)+" characters")
	}
}

// MaxLength checks if a string has maximum length.
func (v *Validator) MaxLength(field, value string, max int) {
	if utf8.RuneCountInString(value) > max {
		v.AddError(field, "Must be at most "+strconv.Itoa(max)+" characters")
	}
}

// Email validates an email address format.
func (v *Validator) Email(field, value string) {
	// Normalize before validation
	value = strings.ToLower(strings.TrimSpace(value))
	if !emailRegex.MatchString(value) {
		v.AddError(field, "Must be a valid email address")
	}
}

// Username validates a username format.
// Username must start with a letter and may contain letters, digits,
// underscores, or hyphens (handle-style). Single-word or multi-word
// names are accepted (e.g., "alice", "John", "alice_smith").
func (v *Validator) Username(field, value string) {
	// Trim first
	value = strings.TrimSpace(value)
	if utf8.RuneCountInString(value) < 2 || utf8.RuneCountInString(value) > 50 {
		v.AddError(field, "Name must be between 2 and 50 characters")
		return
	}

	// Split by space to validate each part
	parts := strings.Fields(value)
	if len(parts) == 0 {
		v.AddError(field, "Please enter your name (e.g., Alice or Alice Smith)")
		return
	}

	// Each part must start with a letter and contain only letters, digits, underscores, hyphens
	for _, part := range parts {
		if !namePartRegex.MatchString(part) {
			v.AddError(field, "Name must start with a letter and contain only letters, digits, underscores, or hyphens")
			return
		}
	}
}

// Password validates password strength.
// It checks minimum length, and requires at least one uppercase letter,
// one lowercase letter, and one digit.
func (v *Validator) Password(field, value string, minLength int) {
	if utf8.RuneCountInString(value) < minLength {
		v.AddError(field, "Password must be at least "+strconv.Itoa(minLength)+" characters long")
		return // If too short, just report length first
	}

	var missing []string
	hasUpper, hasLower, hasDigit := false, false, false
	for _, r := range value {
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		if unicode.IsLower(r) {
			hasLower = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	if !hasUpper {
		missing = append(missing, "an uppercase letter")
	}
	if !hasLower {
		missing = append(missing, "a lowercase letter")
	}
	if !hasDigit {
		missing = append(missing, "a digit")
	}

	if len(missing) > 0 {
		v.AddError(field, "Password must contain at least "+strings.Join(missing, ", "))
	}
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

