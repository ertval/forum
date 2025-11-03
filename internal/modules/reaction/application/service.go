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
func (s *Service) React(ctx context.Context, userID, targetID int, targetType string, reactionType domain.ReactionType) error {
	// Implementation placeholder
	// 1. Validate target type and reaction type
	// 2. Check if user already has a reaction on this target
	// 3. If same reaction exists, do nothing (idempotent)
	// 4. If different reaction exists, delete old and create new
	// 5. If no reaction exists, create new
	return nil
}

// RemoveReaction removes a user's reaction from a target.
func (s *Service) RemoveReaction(ctx context.Context, userID, targetID int, targetType string) error {
	return s.reactionRepo.Delete(ctx, userID, targetID, targetType)
}

// GetReactions retrieves all reactions for a target.
func (s *Service) GetReactions(ctx context.Context, targetID int, targetType string) ([]*domain.Reaction, error) {
	return s.reactionRepo.GetByTarget(ctx, targetID, targetType)
}

// CountReactions returns the count of likes and dislikes for a target.
// TODO: Implement reaction counting.
func (s *Service) CountReactions(ctx context.Context, targetID int, targetType string) (likes, dislikes int, err error) {
	// Implementation placeholder
	// 1. Count likes using reactionRepo.Count
	// 2. Count dislikes using reactionRepo.Count
	// 3. Return both counts
	return 0, 0, nil
}
