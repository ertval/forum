// Package application implements the user service business logic.
package application

import (
	"context"
	"fmt"
	"time"

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

// CreateUser creates a new user with the given details.
// Returns the created user's internal ID or an error.
func (s *Service) CreateUser(ctx context.Context, email, username, passwordHash string) (userID int, err error) {
	user := &domain.User{
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
		Role:         domain.RoleUser, // Default role is User
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return 0, fmt.Errorf("creating user: %w", err)
	}

	return user.ID, nil
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

// UpdateRole updates a user's role.
// Validates the new role and updates the user in the repository.
func (s *Service) UpdateRole(ctx context.Context, userID int, newRole domain.Role) error {
	// Validate role
	if !isValidRole(newRole) {
		return domain.ErrInvalidRole
	}

	// Get the user first to ensure they exist
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("getting user for role update: %w", err)
	}

	// Update role and timestamp
	user.Role = newRole
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(ctx, user)
}

// DeactivateUser deactivates a user account.
// Sets IsActive to false and updates the repository.
func (s *Service) DeactivateUser(ctx context.Context, userID int) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("getting user for deactivation: %w", err)
	}

	if !user.IsActive {
		return nil // Already inactive
	}

	user.IsActive = false
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(ctx, user)
}

// ActivateUser reactivates a user account.
// Sets IsActive to true and updates the repository.
func (s *Service) ActivateUser(ctx context.Context, userID int) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("getting user for activation: %w", err)
	}

	if user.IsActive {
		return nil // Already active
	}

	user.IsActive = true
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(ctx, user)
}

// isValidRole checks if a role string is a valid role.
func isValidRole(role domain.Role) bool {
	switch role {
	case domain.RoleGuest, domain.RoleUser, domain.RoleModerator, domain.RoleAdmin:
		return true
	default:
		return false
	}
}

// ListUsers returns a paginated list of users.
// TODO: Implement user listing.
func (s *Service) ListUsers(ctx context.Context, offset, limit int) ([]*domain.User, error) {
	return s.userRepo.List(ctx, offset, limit)
}

// IncrementPostCount atomically increments the user's post count.
func (s *Service) IncrementPostCount(ctx context.Context, userID int) error {
	return s.userRepo.IncrementPostCount(ctx, userID)
}

// DecrementPostCount atomically decrements the user's post count.
func (s *Service) DecrementPostCount(ctx context.Context, userID int) error {
	return s.userRepo.DecrementPostCount(ctx, userID)
}

// IncrementCommentCount atomically increments the user's comment count.
func (s *Service) IncrementCommentCount(ctx context.Context, userID int) error {
	return s.userRepo.IncrementCommentCount(ctx, userID)
}

// DecrementCommentCount atomically decrements the user's comment count.
func (s *Service) DecrementCommentCount(ctx context.Context, userID int) error {
	return s.userRepo.DecrementCommentCount(ctx, userID)
}

// ExistsByEmail checks if a user with the given email exists.
func (s *Service) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return s.userRepo.ExistsByEmail(ctx, email)
}

// ExistsByUsername checks if a user with the given username exists.
func (s *Service) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return s.userRepo.ExistsByUsername(ctx, username)
}

// IncrementReactionCount atomically increments the user's reaction count.
func (s *Service) IncrementReactionCount(ctx context.Context, userID int) error {
	return s.userRepo.IncrementReactionCount(ctx, userID)
}

// DecrementReactionCount atomically decrements the user's reaction count.
func (s *Service) DecrementReactionCount(ctx context.Context, userID int) error {
	return s.userRepo.DecrementReactionCount(ctx, userID)
}
