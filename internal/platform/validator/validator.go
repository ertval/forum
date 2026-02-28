// Package validator provides input validation and sanitization utilities.
// It validates user input according to business rules and security requirements.
package validator

import (
	"html"
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
	// namePartRegex validates name parts (must start with capital)
	namePartRegex = regexp.MustCompile(`^[A-Z][a-zA-Z]*$`)
	// Sanitization regexes
	reScript = regexp.MustCompile(`(?i)<script[^>]*>[\s\S]*?</script>`)
	reStyle  = regexp.MustCompile(`(?i)<style[^>]*>[\s\S]*?</style>`)
	reTags   = regexp.MustCompile(`<[^>]+>`)
	reSpace  = regexp.MustCompile(`\s+`)
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
	if strings.TrimSpace(Sanitize(value)) == "" {
		v.AddError(field, "This field is required")
	}
}

// MinLength checks if a string has minimum length.
func (v *Validator) MinLength(field, value string, min int) {
	value = Sanitize(value)
	if utf8.RuneCountInString(value) < min {
		v.AddError(field, "Must be at least "+strconv.Itoa(min)+" characters")
	}
}

// MaxLength checks if a string has maximum length.
func (v *Validator) MaxLength(field, value string, max int) {
	value = Sanitize(value)
	if utf8.RuneCountInString(value) > max {
		v.AddError(field, "Must be at most "+strconv.Itoa(max)+" characters")
	}
}

// Email validates an email address format.
func (v *Validator) Email(field, value string) {
	// Normalize and sanitize before validation
	value = strings.ToLower(strings.TrimSpace(Sanitize(value)))
	if !emailRegex.MatchString(value) {
		v.AddError(field, "Must be a valid email address")
	}
}

// Username validates a username format.
// Username must be a proper name (e.g., "John" or "John Smith").
// Each part must start with a capital letter and contain only letters.
// Single names (e.g., "Alice") or full names (e.g., "Alice Smith") are accepted.
// No numbers or special symbols allowed.
func (v *Validator) Username(field, value string) {
	// Sanitize and trim first
	value = strings.TrimSpace(Sanitize(value))
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

	// Each part must contain only letters and start with capital letter
	// Allows both "Alice" and "alice" patterns, but must start with capital
	for _, part := range parts {
		if !namePartRegex.MatchString(part) {
			v.AddError(field, "Name must start with a capital letter and contain only letters (e.g., Alice or Alice Smith)")
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
	value = Sanitize(value)
	for _, a := range allowed {
		if value == a {
			return
		}
	}
	v.AddError(field, "Invalid value")
}

// Matches checks if a value matches a regular expression.
func (v *Validator) Matches(field, value string, pattern *regexp.Regexp) {
	value = Sanitize(value)
	if !pattern.MatchString(value) {
		v.AddError(field, "Invalid format")
	}
}

// Sanitize removes potentially dangerous characters from input.
// Sanitize performs lightweight, safe sanitization of user-supplied text.
// It is intentionally conservative: it trims, collapses whitespace, removes
// control characters (including NUL), strips HTML tags and script/style
// blocks, and unescapes HTML entities. This keeps the implementation
// dependency free and easy to audit.
func Sanitize(input string) string {
	if input == "" {
		return ""
	}

	// Unescape any HTML entities first (e.g., &lt;script&gt;)
	s := html.UnescapeString(input)

	// Remove script blocks and style blocks (case-insensitive)
	s = reScript.ReplaceAllString(s, "")
	s = reStyle.ReplaceAllString(s, "")

	// Strip remaining tags
	s = reTags.ReplaceAllString(s, "")

	// Remove control characters (except common whitespace)
	var b strings.Builder
	for _, r := range s {
		if unicode.IsControl(r) {
			// allow tab, newline and carriage return
			if r == '\t' || r == '\n' || r == '\r' {
				b.WriteRune(r)
			}
			continue
		}
		b.WriteRune(r)
	}
	s = b.String()

	// Collapse all whitespace sequences to a single space
	s = reSpace.ReplaceAllString(s, " ")

	// Trim edges
	s = strings.TrimSpace(s)

	return s
}

// SanitizeHTML sanitizes HTML input to prevent XSS attacks.
// SanitizeHTML removes dangerous elements from HTML while leaving
// plain text and minimal markup. This implementation strips script/style
// blocks and removes all tags — it intentionally does not attempt to
// preserve safe tags to keep behavior predictable and dependency-free.
func SanitizeHTML(input string) string {
	// Reuse Sanitize which already strips tags and script/style blocks.
	return Sanitize(input)
}
