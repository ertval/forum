// OUTPUT PORT - Repository Interface
// Package ports defines the output ports (data access contracts) for the user module.
package ports

import (
	"context"
	"forum/internal/modules/user/domain"
)

// UserRepository defines the data access contract for users.
type UserRepository interface {
	// Create stores a new user in the repository.
	Create(ctx context.Context, user *domain.User) error

	// GetByID retrieves a user by their internal ID (for internal use only).
	GetByID(ctx context.Context, userID int) (*domain.User, error)

	// GetByPublicID retrieves a user by their public UUID (for external API access).
	GetByPublicID(ctx context.Context, publicID string) (*domain.User, error)

	// GetByEmail retrieves a user by their email address.
	GetByEmail(ctx context.Context, email string) (*domain.User, error)

	// GetByUsername retrieves a user by their username.
	GetByUsername(ctx context.Context, username string) (*domain.User, error)

	// Update updates an existing user in the repository.
	Update(ctx context.Context, user *domain.User) error

	// Delete removes a user from the repository.
	Delete(ctx context.Context, userID int) error

	// List returns a paginated list of users.
	List(ctx context.Context, offset, limit int) ([]*domain.User, error)

	// Count returns the total number of users.
	Count(ctx context.Context) (int, error)

	// ExistsByEmail checks if a user with the given email exists.
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// ExistsByUsername checks if a user with the given username exists.
	ExistsByUsername(ctx context.Context, username string) (bool, error)

	// IncrementPostCount atomically increments the user's post count.
	IncrementPostCount(ctx context.Context, userID int) error

	// DecrementPostCount atomically decrements the user's post count.
	DecrementPostCount(ctx context.Context, userID int) error

	// IncrementCommentCount atomically increments the user's comment count.
	IncrementCommentCount(ctx context.Context, userID int) error

	// DecrementCommentCount atomically decrements the user's comment count.
	DecrementCommentCount(ctx context.Context, userID int) error
}
