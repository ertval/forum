// OUTPUT PORT - Repository Interface
// Package ports defines the output ports (data access contracts) for the auth module.
// These interfaces define how the auth module persists and retrieves data.
package ports

import (
	"context"
	"forum/internal/modules/auth/domain"
)

// SessionRepository defines the data access contract for sessions.
// Implementations will provide concrete database operations.
type SessionRepository interface {
	// Create stores a new session in the repository.
	// Returns an error if the session cannot be created.
	Create(ctx context.Context, session *domain.Session) error

	// GetByToken retrieves a session by its token.
	// Returns ErrSessionNotFound if the session doesn't exist.
	GetByToken(ctx context.Context, token string) (*domain.Session, error)

	// GetByUserID retrieves all active sessions for a user.
	// Returns an empty slice if no sessions are found.
	GetByUserID(ctx context.Context, userID int) ([]*domain.Session, error)

	// Update updates an existing session in the repository.
	// Returns an error if the session doesn't exist or update fails.
	Update(ctx context.Context, session *domain.Session) error

	// Delete removes a session from the repository by its token.
	// Returns an error if the session doesn't exist or deletion fails.
	Delete(ctx context.Context, token string) error

	// DeleteByUserID removes all sessions for a user.
	// Used when a user logs out from all devices.
	DeleteByUserID(ctx context.Context, userID int) error

	// DeleteExpired removes all expired sessions from the repository.
	// This should be called periodically to clean up old sessions.
	DeleteExpired(ctx context.Context) error
}

// UserRepository defines the data access contract for user-related operations needed by auth.
// This interface is defined here to avoid circular dependencies with the user module.
type UserRepository interface {
	// Create stores a new user in the repository.
	// Returns the created user's ID or an error if creation fails.
	Create(ctx context.Context, email, username, passwordHash string) (int, error)

	// GetByEmail retrieves a user by email address.
	// Returns the user ID and password hash, or an error if not found.
	GetByEmail(ctx context.Context, email string) (userID int, passwordHash string, err error)

	// ExistsByEmail checks if a user with the given email already exists.
	// Returns true if the user exists, false otherwise.
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// ExistsByUsername checks if a user with the given username already exists.
	// Returns true if the user exists, false otherwise.
	ExistsByUsername(ctx context.Context, username string) (bool, error)
}
