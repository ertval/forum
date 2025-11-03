package output
// Package output defines the outbound ports for the user module.
package output

import (
	"context"

	"forum/internal/modules/user/domain"
)

// UserRepository defines the interface for user persistence.
type UserRepository interface {
	// Create creates a new user.
	Create(ctx context.Context, user *domain.User) error

	// GetByID retrieves a user by ID.
	GetByID(ctx context.Context, userID string) (*domain.User, error)

	// GetByEmail retrieves a user by email.
	GetByEmail(ctx context.Context, email string) (*domain.User, error)

	// GetByUsername retrieves a user by username.
	GetByUsername(ctx context.Context, username string) (*domain.User, error)

	// Update updates a user.
	Update(ctx context.Context, user *domain.User) error

	// Delete deletes a user.
	Delete(ctx context.Context, userID string) error

	// List lists users with pagination.
	List(ctx context.Context, limit, offset int) ([]*domain.User, error)

	// EmailExists checks if an email exists.
	EmailExists(ctx context.Context, email string) (bool, error)

	// UsernameExists checks if a username exists.
	UsernameExists(ctx context.Context, username string) (bool, error)
}
