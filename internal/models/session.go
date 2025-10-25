package models

// session.go defines the Session model and related database operations.
// It handles user session creation, validation, and cleanup.

import (
	"time"
)

// Session represents a user login session
type Session struct {
	ID        int
	UserID    int
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// CreateSession creates a new session for a user with a unique token
func CreateSession(userID int, token string, duration time.Duration) (*Session, error) {
	// Insert session into database
	// Set expiration time based on duration
	// Return created session
	return nil, nil
}

// GetSessionByToken retrieves a session by its token
func GetSessionByToken(token string) (*Session, error) {
	// Query session from database by token
	// Check if session is expired
	return nil, nil
}

// DeleteSession removes a session from the database
func DeleteSession(token string) error {
	// Delete session by token
	return nil
}

// DeleteExpiredSessions removes all expired sessions from the database
func DeleteExpiredSessions() error {
	// Delete all sessions where expires_at < current time
	return nil
}

// IsValid checks if a session is still valid (not expired)
func (s *Session) IsValid() bool {
	// Check if session expiration time is in the future
	return time.Now().Before(s.ExpiresAt)
}
