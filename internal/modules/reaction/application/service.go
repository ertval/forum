// Package application implements reaction service business logic.
package application

import (
	"context"
	"log"
	"time"

	commentPorts "forum/internal/modules/comment/ports"
	postPorts "forum/internal/modules/post/ports"
	"forum/internal/modules/reaction/domain"
	"forum/internal/modules/reaction/ports"
	userPorts "forum/internal/modules/user/ports"
)

// Service implements the ReactionService interface.
type Service struct {
	reactionRepo ports.ReactionRepository
	postRepo     postPorts.PostRepository
	commentRepo  commentPorts.CommentRepository
	userService  userPorts.UserService
}

// NewService creates a new reaction service with all required dependencies.
func NewService(
	reactionRepo ports.ReactionRepository,
	postRepo postPorts.PostRepository,
	commentRepo commentPorts.CommentRepository,
	userService userPorts.UserService,
) *Service {
	return &Service{
		reactionRepo: reactionRepo,
		postRepo:     postRepo,
		commentRepo:  commentRepo,
		userService:  userService,
	}
}

// React adds or updates a user's reaction to a target.
// If the user already has a different reaction, it's replaced (toggle behavior).
func (s *Service) React(ctx context.Context, userID int, targetPublicID string, targetType string, reactionType domain.ReactionType) error {
	// Validate inputs
	if userID <= 0 {
		return domain.ErrInvalidUserID
	}

	if targetType != "post" && targetType != "comment" {
		return domain.ErrInvalidTarget
	}

	if reactionType != domain.ReactionLike && reactionType != domain.ReactionDislike {
		return domain.ErrInvalidReactionType
	}

	// Validate that the target exists
	switch targetType {
	case "post":
		_, err := s.postRepo.GetByID(ctx, targetPublicID)
		if err != nil {
			return err
		}
	case "comment":
		_, err := s.commentRepo.GetByPublicID(ctx, targetPublicID)
		if err != nil {
			return err
		}
	}

	// Check if user already has a reaction on this target
	existingReaction, err := s.reactionRepo.GetByUserAndTargetPublicID(
		ctx,
		userID,
		targetPublicID,
		targetType,
	)

	if err != nil && err != domain.ErrReactionNotFound {
		return err // Some other error occurred
	}

	if existingReaction != nil {
		// If the existing reaction is the same type, remove it (toggle behavior)
		if existingReaction.Type == reactionType {
			return s.RemoveReaction(ctx, userID, targetPublicID, targetType)
		}

		// If it's a different type, remove the old reaction first
		err = s.reactionRepo.DeleteByTargetPublicID(ctx, userID, targetPublicID, targetType)
		if err != nil {
			return err
		}
	}

	// Create new reaction
	reaction := &domain.Reaction{
		UserID:         userID,
		TargetID:       0,              // Will be resolved by repository based on targetPublicID
		PublicTargetID: targetPublicID, // Set the public target ID for repository to resolve
		TargetType:     targetType,
		Type:           reactionType,
		CreatedAt:      time.Now(),
	}

	// The repository will handle resolving the targetPublicID to internal ID
	err = s.reactionRepo.Create(ctx, reaction)
	if err != nil {
		return err
	}

	// Increment user's reaction count asynchronously (non-blocking)
	go func(uid int) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.userService.IncrementReactionCount(ctx, uid); err != nil {
			log.Printf("WARNING: failed to increment reaction count for user %d: %v", uid, err)
		}
	}(userID)

	return nil
}

// RemoveReaction removes a user's reaction from a target.
func (s *Service) RemoveReaction(ctx context.Context, userID int, targetPublicID string, targetType string) error {
	// Validate inputs
	if userID <= 0 {
		return domain.ErrInvalidUserID
	}

	if targetType != "post" && targetType != "comment" {
		return domain.ErrInvalidTarget
	}

	// Verify target exists before attempting to remove reaction
	switch targetType {
	case "post":
		_, err := s.postRepo.GetByID(ctx, targetPublicID)
		if err != nil {
			return err
		}
	case "comment":
		_, err := s.commentRepo.GetByPublicID(ctx, targetPublicID)
		if err != nil {
			return err
		}
	}

	err := s.reactionRepo.DeleteByTargetPublicID(ctx, userID, targetPublicID, targetType)
	if err != nil {
		return err
	}

	// Decrement user's reaction count asynchronously (non-blocking)
	go func(uid int) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.userService.DecrementReactionCount(ctx, uid); err != nil {
			log.Printf("WARNING: failed to decrement reaction count for user %d: %v", uid, err)
		}
	}(userID)

	return nil
}

// GetReactions retrieves all reactions for a target.
func (s *Service) GetReactions(ctx context.Context, targetPublicID string, targetType string) ([]*domain.Reaction, error) {
	// Validate inputs
	if targetType != "post" && targetType != "comment" {
		return nil, domain.ErrInvalidTarget
	}

	// Verify target exists
	switch targetType {
	case "post":
		_, err := s.postRepo.GetByID(ctx, targetPublicID)
		if err != nil {
			return nil, err
		}
	case "comment":
		_, err := s.commentRepo.GetByPublicID(ctx, targetPublicID)
		if err != nil {
			return nil, err
		}
	}

	return s.reactionRepo.GetByTargetPublicID(ctx, targetPublicID, targetType)
}

// CountReactions returns the count of likes and dislikes for a target.
func (s *Service) CountReactions(ctx context.Context, targetPublicID string, targetType string) (likes, dislikes int, err error) {
	// Validate inputs
	if targetType != "post" && targetType != "comment" {
		return 0, 0, domain.ErrInvalidTarget
	}

	// Verify target exists
	switch targetType {
	case "post":
		_, err := s.postRepo.GetByID(ctx, targetPublicID)
		if err != nil {
			return 0, 0, err
		}
	case "comment":
		_, err := s.commentRepo.GetByPublicID(ctx, targetPublicID)
		if err != nil {
			return 0, 0, err
		}
	}

	// Count likes
	likes, err = s.reactionRepo.CountByTargetPublicID(ctx, targetPublicID, targetType, domain.ReactionLike)
	if err != nil {
		return 0, 0, err
	}

	// Count dislikes
	dislikes, err = s.reactionRepo.CountByTargetPublicID(ctx, targetPublicID, targetType, domain.ReactionDislike)
	if err != nil {
		return 0, 0, err
	}

	return likes, dislikes, nil
}

// GetUserReactionCount returns the total number of reactions given by a user.
func (s *Service) GetUserReactionCount(ctx context.Context, userID int) (int, error) {
	if userID <= 0 {
		return 0, domain.ErrInvalidUserID
	}

	return s.reactionRepo.CountByUserID(ctx, userID)
}

// GetByUserAndTargetPublicID retrieves a user's reaction for a specific target by target's public UUID.
func (s *Service) GetByUserAndTargetPublicID(ctx context.Context, userID int, targetPublicID string, targetType string) (*domain.Reaction, error) {
	// Validate inputs
	if userID <= 0 {
		return nil, domain.ErrInvalidUserID
	}

	if targetType != "post" && targetType != "comment" {
		return nil, domain.ErrInvalidTarget
	}

	// Verify target exists
	switch targetType {
	case "post":
		_, err := s.postRepo.GetByID(ctx, targetPublicID)
		if err != nil {
			return nil, err
		}
	case "comment":
		_, err := s.commentRepo.GetByPublicID(ctx, targetPublicID)
		if err != nil {
			return nil, err
		}
	}

	return s.reactionRepo.GetByUserAndTargetPublicID(ctx, userID, targetPublicID, targetType)
}
