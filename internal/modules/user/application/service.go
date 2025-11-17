// Package application implements the user service business logic.
package application

import (
	"context"
	"forum/internal/modules/user/domain"
	"forum/internal/modules/user/ports"
)

// Service implements the UserService interface.
type Service struct {
	userRepo ports.UserRepository
}

// NewService creates a new user service.
func NewService(userRepo ports.UserRepository) *Service {
	return &Service{
		userRepo: userRepo,
	}
}

// GetByID retrieves a user by their internal ID (for internal use only).
// TODO: Implement user retrieval.
func (s *Service) GetByID(ctx context.Context, userID int) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// GetByPublicID retrieves a user by their public UUID (for external API access).
func (s *Service) GetByPublicID(ctx context.Context, publicID string) (*domain.User, error) {
	return s.userRepo.GetByPublicID(ctx, publicID)
}

// GetByUsername retrieves a user by their username.
// TODO: Implement username-based retrieval.
func (s *Service) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	return s.userRepo.GetByUsername(ctx, username)
}

// GetByEmail retrieves a user by their email address.
// TODO: Implement email-based retrieval.
func (s *Service) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.userRepo.GetByEmail(ctx, email)
}

// GetProfile retrieves a user's public profile.
// TODO: Implement profile retrieval.
func (s *Service) GetProfile(ctx context.Context, userID int) (*domain.UserProfile, error) {
	// Implementation placeholder
	return nil, nil
}

// UpdateRole updates a user's role.
// TODO: Implement role update with permission checks.
func (s *Service) UpdateRole(ctx context.Context, userID int, newRole domain.Role) error {
	// Implementation placeholder
	return nil
}

// DeactivateUser deactivates a user account.
// TODO: Implement user deactivation.
func (s *Service) DeactivateUser(ctx context.Context, userID int) error {
	// Implementation placeholder
	return nil
}

// ActivateUser reactivates a user account.
// TODO: Implement user activation.
func (s *Service) ActivateUser(ctx context.Context, userID int) error {
	// Implementation placeholder
	return nil
}

// ListUsers returns a paginated list of users.
// TODO: Implement user listing.
func (s *Service) ListUsers(ctx context.Context, offset, limit int) ([]*domain.User, error) {
	return s.userRepo.List(ctx, offset, limit)
}

// GetUserStats retrieves statistics about a user's activity.
func (s *Service) GetUserStats(ctx context.Context, userID int) (*ports.UserStats, error) {
	return s.userRepo.GetUserStats(ctx, userID)
}
