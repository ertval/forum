package input
// Package input defines the inbound ports (use cases) for the auth module.
// These interfaces represent the operations that the auth module exposes
// to external actors (HTTP handlers, CLI, etc.).
package input

import (
	"context"

	"forum/internal/modules/auth/domain"
)

// AuthService defines the authentication use cases.
type AuthService interface {
	// Register creates a new user account.
	Register(ctx context.Context, email, username, password string) error

	// Login authenticates a user and creates a session.
	Login(ctx context.Context, email, password string) (*domain.Session, error)

	// Logout terminates a user session.
	Logout(ctx context.Context, sessionToken string) error

	// ValidateSession checks if a session is valid and returns the user ID.
	ValidateSession(ctx context.Context, sessionToken string) (userID string, err error)

	// RefreshSession extends the session expiration.
	RefreshSession(ctx context.Context, sessionToken string) (*domain.Session, error)

	// LoginWithOAuth authenticates a user using OAuth provider.
	LoginWithOAuth(ctx context.Context, provider string, code string) (*domain.Session, error)
}
