// Package application implements reaction service business logic.
package application

import (
	"context"
	"forum/internal/modules/reaction/domain"
	"forum/internal/modules/reaction/ports"
)

// Service implements the ReactionService interface.
type Service struct {
	reactionRepo ports.ReactionRepository
}

// NewService creates a new reaction service.
func NewService(reactionRepo ports.ReactionRepository) *Service {
	return &Service{reactionRepo: reactionRepo}
}

// React adds or updates a user's reaction to a target.
// TODO: Implement reaction toggle logic (replace existing opposite reaction).
func (s *Service) React(ctx context.Context, userID int, targetPublicID string, targetType string, reactionType domain.ReactionType) error {
	// Implementation placeholder
	// 1. Validate target type and reaction type
	// 2. Resolve targetPublicID to internal target ID
	// 3. Check if user already has a reaction on this target
	// 4. If same reaction exists, do nothing (idempotent)
	// 5. If different reaction exists, delete old and create new
	// 6. If no reaction exists, create new
	return nil
}

// RemoveReaction removes a user's reaction from a target.
func (s *Service) RemoveReaction(ctx context.Context, userID int, targetPublicID string, targetType string) error {
	return s.reactionRepo.DeleteByTargetPublicID(ctx, userID, targetPublicID, targetType)
}

// GetReactions retrieves all reactions for a target.
func (s *Service) GetReactions(ctx context.Context, targetPublicID string, targetType string) ([]*domain.Reaction, error) {
	return s.reactionRepo.GetByTargetPublicID(ctx, targetPublicID, targetType)
}

// CountReactions returns the count of likes and dislikes for a target.
// TODO: Implement reaction counting.
func (s *Service) CountReactions(ctx context.Context, targetPublicID string, targetType string) (likes, dislikes int, err error) {
	// Implementation placeholder
	// 1. Count likes using reactionRepo.CountByTargetPublicID
	// 2. Count dislikes using reactionRepo.CountByTargetPublicID
	// 3. Return both counts
	return 0, 0, nil
}

// GetUserReactionCount returns the total number of reactions given by a user.
func (s *Service) GetUserReactionCount(ctx context.Context, userID int) (int, error) {
	return s.reactionRepo.CountByUserID(ctx, userID)
}
