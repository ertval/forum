package output
// Package output defines the outbound ports for the auth module.
// These interfaces represent the dependencies that the auth module needs
// from external systems (database, external services, etc.).
package output

import (
	"context"

	"forum/internal/modules/auth/domain"
)

// SessionRepository defines the interface for session persistence.
type SessionRepository interface {
	// Create creates a new session in the database.
	Create(ctx context.Context, session *domain.Session) error

	// GetByToken retrieves a session by its token.
	GetByToken(ctx context.Context, token string) (*domain.Session, error)

	// Delete removes a session from the database.
	Delete(ctx context.Context, token string) error

	// DeleteByUserID removes all sessions for a user.
	DeleteByUserID(ctx context.Context, userID string) error

	// DeleteExpired removes all expired sessions.
	DeleteExpired(ctx context.Context) error

	// Update updates an existing session.
	Update(ctx context.Context, session *domain.Session) error
}

// UserRepository defines the interface for user-related operations needed by auth.
// Note: This is a minimal interface. The full user repository is in the user module.
type UserRepository interface {
	// GetByEmail retrieves a user by email.
	GetByEmail(ctx context.Context, email string) (userID, hashedPassword string, err error)

	// GetByUsername retrieves a user by username.
	GetByUsername(ctx context.Context, username string) (userID string, err error)

	// EmailExists checks if an email is already registered.
	EmailExists(ctx context.Context, email string) (bool, error)

	// UsernameExists checks if a username is already taken.
	UsernameExists(ctx context.Context, username string) (bool, error)

	// Create creates a new user. Returns the user ID.
	Create(ctx context.Context, email, username, hashedPassword string) (string, error)

	// GetByOAuthProvider retrieves a user by OAuth provider and provider ID.
	GetByOAuthProvider(ctx context.Context, provider, providerID string) (userID string, err error)

	// CreateOAuthUser creates a new user from OAuth credentials.
	CreateOAuthUser(ctx context.Context, creds domain.OAuthCredentials) (userID string, err error)
}

// PasswordHasher defines the interface for password hashing operations.
type PasswordHasher interface {
	// Hash generates a hash from a plain text password.
	Hash(password string) (string, error)

	// Compare compares a plain text password with a hash.
	Compare(hashedPassword, password string) error
}

// OAuthProvider defines the interface for OAuth operations.
type OAuthProvider interface {
	// GetAuthURL returns the OAuth authorization URL.
	GetAuthURL(state string) string

	// ExchangeCode exchanges an authorization code for an access token.
	ExchangeCode(ctx context.Context, code string) (string, error)

	// GetUserInfo retrieves user information from the OAuth provider.
	GetUserInfo(ctx context.Context, accessToken string) (*domain.OAuthCredentials, error)
}
