// Package domain contains the core business entities for the auth module.
// Domain entities represent the fundamental concepts of authentication and sessions.
package domain

import (
	"time"
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

// Credentials represents user credentials for authentication.
type Credentials struct {
	Email    string // User's email address
	Password string // User's password (plaintext, will be hashed)
}
