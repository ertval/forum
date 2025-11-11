// Package domain contains the core business entities for the auth module.
// Domain entities represent the fundamental concepts of authentication and sessions.
package domain

import (
	"errors"
	"time"
	"forum/internal/platform/validator"
)

// Session represents an authenticated user session.
// Sessions are created when a user logs in and expire after a certain duration.
type Session struct {
	ID        string    // Unique session identifier (UUID)
	UserID    int       // ID of the authenticated user
	Token     string    // Session token stored in cookie
	ExpiresAt time.Time // Session expiration time
	CreatedAt time.Time // Session creation time
	IPAddress string    // IP address of the client
	UserAgent string    // User agent string of the client
}

// IsExpired checks if the session has expired.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsValid checks if the session is valid (not expired and has required fields).
func (s *Session) IsValid() bool {
	return s.ID != "" && s.UserID > 0 && !s.IsExpired()
}

// Credentials represents user credentials for authentication.
type Credentials struct {
	Email    string // User's email address
	Password string // User's password (plaintext, will be hashed)
}

// Validate validates the credentials.
// Returns an error if the credentials are invalid.
func (c *Credentials) Validate() error {
	validator := validator.New()
	
	validator.Required("email", c.Email)
	if c.Email != "" {
		validator.Email("email", c.Email)
	}
	
	validator.Required("password", c.Password)
	if c.Password != "" {
		// Using minimum 6 characters for basic validation, can be increased for production
		validator.Password("password", c.Password, 6)
	}
	
	if !validator.Valid() {
		// Convert validator errors to error string for now
		// In a real implementation, you might want to return a more structured error
		for _, msg := range validator.Errors() {
			return errors.New(msg)
		}
	}
	
	return nil
}
