package domain
// Package domain contains the core authentication business logic and entities.
// This is the heart of the auth module and has no external dependencies.
package domain

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

// Session represents a user session in the domain.
type Session struct {
	ID        uuid.UUID
	UserID    string
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
	IPAddress string
	UserAgent string
}

// IsExpired checks if the session has expired.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsValid checks if the session is valid (not expired and has valid token).
func (s *Session) IsValid() bool {
	return !s.IsExpired() && s.Token != ""
}
