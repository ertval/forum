// Package domain contains the core business entities for the auth module.
// Domain entities represent the fundamental concepts of authentication and sessions.
package domain

import (
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// Pre-compiled validation regexes (stdlib only, no external dependencies).
var (
	credEmailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
)

// Session represents an authenticated user session.
// Sessions are created when a user logs in and expire after a certain duration.
type Session struct {
	ID        int       `json:"-"`          // Internal unique identifier (INT PRIMARY KEY)
	PublicID  string    `json:"id"`         // Public UUID identifier (exposed in API)
	UserID    int       `json:"-"`          // ID of the authenticated user (internal INT)
	Token     string    `json:"token"`      // Session token stored in cookie
	ExpiresAt time.Time `json:"expires_at"` // Session expiration time
	CreatedAt time.Time `json:"created_at"` // Session creation time
	IPAddress string    `json:"-"`          // IP address of the client (internal only)
	UserAgent string    `json:"-"`          // User agent string of the client (internal only)
}

// IsExpired checks if the session has expired.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsValid checks if the session is valid (not expired and has required fields).
func (s *Session) IsValid() bool {
	return s.ID > 0 && s.UserID > 0 && !s.IsExpired()
}

// Validate checks that the session has the required fields set.
func (s *Session) Validate() error {
	if s.Token == "" {
		return ErrInvalidSession
	}
	if s.UserID <= 0 {
		return ErrInvalidSession
	}
	if s.ExpiresAt.IsZero() {
		return ErrInvalidSession
	}
	return nil
}

// Credentials represents user credentials for authentication.
type Credentials struct {
	Email    string // User's email address
	Password string // User's password (plaintext, will be hashed)
}

// Validate checks that the credentials meet format and strength requirements.
// This is used by the application layer instead of importing platform/validator.
func (c *Credentials) Validate() error {
	// Validate email
	email := strings.ToLower(strings.TrimSpace(c.Email))
	if email == "" {
		return ErrInvalidEmail
	}
	if !credEmailRegex.MatchString(email) {
		return ErrInvalidEmail
	}

	// Validate password presence
	if strings.TrimSpace(c.Password) == "" {
		return &PasswordValidationError{Message: "Password is required"}
	}

	// Validate password strength (minimum 8 chars, upper, lower, digit)
	n := utf8.RuneCountInString(c.Password)
	if n < 8 {
		return &PasswordValidationError{Message: "Password must be at least 8 characters long"}
	}

	var hasUpper, hasLower, hasDigit bool
	for _, r := range c.Password {
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

	missing := make([]string, 0, 3)
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
		return &PasswordValidationError{Message: "Password must contain at least " + strings.Join(missing, ", ")}
	}

	return nil
}
