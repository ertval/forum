// INPUT PORT - Service Interface
// Package ports defines the input ports (use cases) for the auth module.
// These interfaces define what the auth module can do from an external perspective.
package ports

import (
	"context"
	"forum/internal/modules/auth/domain"
)

// AuthService defines the authentication use cases.
// This is the primary interface for authentication operations.
type AuthService interface {
	// Register creates a new user account.
	// Returns the created user ID and session, or an error if registration fails.
	Register(ctx context.Context, email, username, password string) (userID int, session *domain.Session, err error)

	// Login authenticates a user with email and password.
	// Returns a session if authentication succeeds, or an error if it fails.
	Login(ctx context.Context, email, password string) (*domain.Session, error)

	// Logout invalidates the session with the given token.
	// Returns an error if the session doesn't exist or logout fails.
	Logout(ctx context.Context, sessionToken string) error

	// ValidateSession checks if a session token is valid and not expired.
	// Returns the session if valid, or an error if invalid/expired.
	ValidateSession(ctx context.Context, sessionToken string) (*domain.Session, error)

	// RefreshSession extends the expiration time of an existing session.
	// Returns the updated session or an error if refresh fails.
	RefreshSession(ctx context.Context, sessionToken string) (*domain.Session, error)

	// GetSession retrieves a session by its token.
	// Returns the session or an error if not found.
	GetSession(ctx context.Context, sessionToken string) (*domain.Session, error)
}
