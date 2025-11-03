package input
// Package input defines the inbound ports for the user module.
package input

import (
	"context"

	"forum/internal/modules/user/domain"
)

// UserService defines the user management use cases.
type UserService interface {
	// GetByID retrieves a user by ID.
	GetByID(ctx context.Context, userID string) (*domain.User, error)

	// GetByEmail retrieves a user by email.
	GetByEmail(ctx context.Context, email string) (*domain.User, error)

	// GetByUsername retrieves a user by username.
	GetByUsername(ctx context.Context, username string) (*domain.User, error)

	// UpdateProfile updates user profile information.
	UpdateProfile(ctx context.Context, userID string, username string) error

	// PromoteToModerator promotes a user to moderator role.
	PromoteToModerator(ctx context.Context, adminID, userID string) error

	// DemoteFromModerator demotes a moderator to user role.
	DemoteFromModerator(ctx context.Context, adminID, userID string) error

	// Delete deletes a user account.
	Delete(ctx context.Context, userID string) error

	// List lists users with pagination.
	List(ctx context.Context, limit, offset int) ([]*domain.User, error)
}
