// Package application implements reaction service business logic.
package application

import (
	"context"
	"fmt"
	"log"
	"time"

	commentPorts "forum/internal/modules/comment/ports"
	notificationDomain "forum/internal/modules/notification/domain"
	notificationPorts "forum/internal/modules/notification/ports"
	postPorts "forum/internal/modules/post/ports"
	"forum/internal/modules/reaction/domain"
	"forum/internal/modules/reaction/ports"
	userPorts "forum/internal/modules/user/ports"
	"forum/internal/platform/async"
)

// Service implements the ReactionService interface.
type Service struct {
	reactionRepo        ports.ReactionRepository
	postRepo            postPorts.PostRepository
	commentRepo         commentPorts.CommentRepository
	userService         userPorts.UserService
	notificationService notificationPorts.NotificationService
}

// NewService creates a new reaction service with all required dependencies.
func NewService(
	reactionRepo ports.ReactionRepository,
	postRepo postPorts.PostRepository,
	commentRepo commentPorts.CommentRepository,
	userService userPorts.UserService,
	notificationService notificationPorts.NotificationService,
) *Service {
	return &Service{
		reactionRepo:        reactionRepo,
		postRepo:            postRepo,
		commentRepo:         commentRepo,
		userService:         userService,
		notificationService: notificationService,
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

	// Build the reaction domain object
	reaction := &domain.Reaction{
		UserID:         userID,
		PublicTargetID: targetPublicID,
		TargetType:     targetType,
		Type:           reactionType,
		CreatedAt:      time.Now(),
	}

	// Atomically toggle/create the reaction in a single transaction.
	// This avoids the TOCTOU race of separate read → delete → create steps.
	removed, err := s.reactionRepo.ToggleReaction(ctx, reaction)
	if err != nil {
		return err
	}

	if removed {
		// Reaction was toggled off — decrement user's reaction count
		async.Run(func(ctx context.Context) error {
			return s.userService.DecrementReactionCount(ctx, userID)
		}, fmt.Sprintf("decrement reaction count for user %d", userID))
		return nil
	}

	// Reaction was created or updated — send notification for posts
	if s.notificationService != nil && targetType == "post" {
		post, postErr := s.postRepo.GetByID(ctx, targetPublicID)
		if postErr == nil && post.UserID != userID {
			notificationType := notificationDomain.TypeLike
			message := "Someone liked your post"
			if reactionType == domain.ReactionDislike {
				notificationType = notificationDomain.TypeDislike
				message = "Someone disliked your post"
			}
			if err := s.notificationService.CreateNotification(ctx, post.UserID, userID, notificationType, message, targetPublicID); err != nil {
				log.Printf("WARNING: failed to create reaction notification for post owner %d: %v", post.UserID, err)
			}
		}
	}

	// Increment user's reaction count asynchronously (non-blocking)
	async.Run(func(ctx context.Context) error {
		return s.userService.IncrementReactionCount(ctx, userID)
	}, fmt.Sprintf("increment reaction count for user %d", userID))

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

	// Delete the reaction — the repository handles target resolution and returns
	// ErrTargetNotFound or ErrReactionNotFound as appropriate.
	err := s.reactionRepo.DeleteByTargetPublicID(ctx, userID, targetPublicID, targetType)
	if err != nil {
		return err
	}

	// Decrement user's reaction count asynchronously (non-blocking)
	async.Run(func(ctx context.Context) error {
		return s.userService.DecrementReactionCount(ctx, userID)
	}, fmt.Sprintf("decrement reaction count for user %d", userID))

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

// CountReactionsBatch returns likes/dislikes counts for multiple targets in a single query.
func (s *Service) CountReactionsBatch(ctx context.Context, targetPublicIDs []string, targetType string) (map[string]map[string]int, error) {
	if targetType != "post" && targetType != "comment" {
		return nil, domain.ErrInvalidTarget
	}
	if len(targetPublicIDs) == 0 {
		return make(map[string]map[string]int), nil
	}
	return s.reactionRepo.CountBatchByTargetPublicIDs(ctx, targetPublicIDs, targetType)
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
