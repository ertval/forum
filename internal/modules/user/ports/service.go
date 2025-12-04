package ports

// INPUT PORT - Service Interface
// Package ports defines the input ports (use cases) for the user module.

import (
	"context"
	"forum/internal/modules/user/domain"
)

// UserService defines the user management use cases.
type UserService interface {
	// CreateUser creates a new user with the given details.
	// Returns the created user's internal ID or an error.
	CreateUser(ctx context.Context, email, username, passwordHash string) (userID int, err error)

	// GetByID retrieves a user by their internal ID (for internal use only).
	GetByID(ctx context.Context, userID int) (*domain.User, error)

	// GetByPublicID retrieves a user by their public UUID (for external API access).
	GetByPublicID(ctx context.Context, publicID string) (*domain.User, error)

	// GetByUsername retrieves a user by their username.
	GetByUsername(ctx context.Context, username string) (*domain.User, error)

	// GetByEmail retrieves a user by their email address.
	GetByEmail(ctx context.Context, email string) (*domain.User, error)

	// UpdateRole updates a user's role (requires admin permissions).
	UpdateRole(ctx context.Context, userID int, newRole domain.Role) error

	// DeactivateUser deactivates a user account.
	DeactivateUser(ctx context.Context, userID int) error

	// ActivateUser reactivates a user account.
	ActivateUser(ctx context.Context, userID int) error

	// ListUsers returns a paginated list of users.
	ListUsers(ctx context.Context, offset, limit int) ([]*domain.User, error)

	// IncrementPostCount atomically increments the user's post count.
	IncrementPostCount(ctx context.Context, userID int) error

	// DecrementPostCount atomically decrements the user's post count.
	DecrementPostCount(ctx context.Context, userID int) error

	// IncrementCommentCount atomically increments the user's comment count.
	IncrementCommentCount(ctx context.Context, userID int) error

	// DecrementCommentCount atomically decrements the user's comment count.
	DecrementCommentCount(ctx context.Context, userID int) error

	// ExistsByEmail checks if a user with the given email exists.
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// ExistsByUsername checks if a user with the given username exists.
	ExistsByUsername(ctx context.Context, username string) (bool, error)
}
