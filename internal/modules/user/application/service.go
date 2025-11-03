package application
// Package application contains user application services.
package application

import (
	"context"

	"forum/internal/modules/user/domain"
	"forum/internal/modules/user/ports/input"
	"forum/internal/modules/user/ports/output"
)

// Service implements the UserService interface.
type Service struct {
	userRepo output.UserRepository
}

// NewService creates a new user service.
func NewService(userRepo output.UserRepository) input.UserService {
	return &Service{
		userRepo: userRepo,
	}
}

// GetByID retrieves a user by ID.
func (s *Service) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	// TODO: Implement
	return nil, nil
}

// GetByEmail retrieves a user by email.
func (s *Service) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	// TODO: Implement
	return nil, nil
}

// GetByUsername retrieves a user by username.
func (s *Service) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	// TODO: Implement
	return nil, nil
}

// UpdateProfile updates user profile.
func (s *Service) UpdateProfile(ctx context.Context, userID string, username string) error {
	// TODO: Implement
	return nil
}

// PromoteToModerator promotes a user to moderator.
func (s *Service) PromoteToModerator(ctx context.Context, adminID, userID string) error {
	// TODO: Implement
	return nil
}

// DemoteFromModerator demotes a moderator to user.
func (s *Service) DemoteFromModerator(ctx context.Context, adminID, userID string) error {
	// TODO: Implement
	return nil
}

// Delete deletes a user.
func (s *Service) Delete(ctx context.Context, userID string) error {
	// TODO: Implement
	return nil
}

// List lists users.
func (s *Service) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	// TODO: Implement
	return nil, nil
}
