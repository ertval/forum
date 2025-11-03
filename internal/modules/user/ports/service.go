package ports
// INPUT PORT - Service Interface
// Package ports defines the input ports (use cases) for the user module.

import (
	"context"
	"forum/internal/modules/user/domain"
)

// UserService defines the user management use cases.
type UserService interface {
	// GetByID retrieves a user by their ID.
	GetByID(ctx context.Context, userID int) (*domain.User, error)
	
	// GetByUsername retrieves a user by their username.
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	
	// GetByEmail retrieves a user by their email address.
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	
	// GetProfile retrieves a user's public profile.
	GetProfile(ctx context.Context, userID int) (*domain.UserProfile, error)
	
	// UpdateRole updates a user's role (requires admin permissions).
	UpdateRole(ctx context.Context, userID int, newRole domain.Role) error
	
	// DeactivateUser deactivates a user account.
	DeactivateUser(ctx context.Context, userID int) error
	
	// ActivateUser reactivates a user account.
	ActivateUser(ctx context.Context, userID int) error
	
	// ListUsers returns a paginated list of users.
	ListUsers(ctx context.Context, offset, limit int) ([]*domain.User, error)
	
	// GetUserStats retrieves statistics about a user's activity.
	GetUserStats(ctx context.Context, userID int) (*UserStats, error)
}

// UserStats represents a user's activity statistics.
type UserStats struct {
	PostCount     int
	CommentCount  int
	LikeCount     int
	DislikeCount  int
}
