// Package application implements reaction service business logic.
package application

import (
	"context"
	"fmt"
	"log"
	"time"

	"forum/internal/modules/reaction/domain"
	"forum/internal/modules/reaction/ports"
	"forum/internal/platform/async"
)

// Notification type constants mirrored locally to avoid cross-module port imports.
const (
	notifTypeLike    = "like"
	notifTypeDislike = "dislike"
)

type PostRecord struct {
	UserID int
}

// postRepository defines the minimal post data access required by the reaction service.
// This avoids a direct import of the post module's ports package.
type postRepository interface {
	GetPostForReaction(ctx context.Context, postID string) (*PostRecord, error)
}

// commentRepository defines the minimal comment data access required by the reaction service.
// This avoids a direct import of the comment module's ports package.
type commentRepository interface {
	EnsureCommentExists(ctx context.Context, commentPublicID string) error
}

// userService defines the minimal user operations required by the reaction service.
// This avoids a direct import of the user module's ports package.
type userService interface {
	IncrementReactionCount(ctx context.Context, userID int) error
	DecrementReactionCount(ctx context.Context, userID int) error
}

// notificationService defines the minimal notification operations required by the reaction service.
// This avoids a direct import of the notification module's ports package.
type notificationService interface {
	CreateNotification(ctx context.Context, userID, actorID int, notifType, message string, targetPublicID string) error
}

// Service implements the ReactionService interface.
type Service struct {
	reactionRepo        ports.ReactionRepository
	postRepo            postRepository
	commentRepo         commentRepository
	userService         userService
	notificationService notificationService
}

// NewService creates a new reaction service with all required dependencies.
func NewService(
	reactionRepo ports.ReactionRepository,
	postRepo postRepository,
	commentRepo commentRepository,
	userService userService,
	notificationService notificationService,
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
	action, err := s.reactionRepo.ToggleReaction(ctx, reaction)
	if err != nil {
		return err
	}

	if action == domain.ToggleActionRemoved {
		// Reaction was toggled off — decrement user's reaction count
		async.Run(func(ctx context.Context) error {
			return s.userService.DecrementReactionCount(ctx, userID)
		}, fmt.Sprintf("decrement reaction count for user %d", userID))
		return nil
	}

	// Reaction was created or updated — send notification for posts
	if s.notificationService != nil && targetType == "post" {
		post, postErr := s.postRepo.GetPostForReaction(ctx, targetPublicID)
		if postErr == nil && post.UserID != userID {
			notificationType := notifTypeLike
			message := "Someone liked your post"
			if reactionType == domain.ReactionDislike {
				notificationType = notifTypeDislike
				message = "Someone disliked your post"
			}
			if err := s.notificationService.CreateNotification(ctx, post.UserID, userID, notificationType, message, targetPublicID); err != nil {
				log.Printf("WARNING: failed to create reaction notification for post owner %d: %v", post.UserID, err)
			}
		}
	}

	// Increment user's reaction count only when a new reaction is created.
	if action == domain.ToggleActionCreated {
		async.Run(func(ctx context.Context) error {
			return s.userService.IncrementReactionCount(ctx, userID)
		}, fmt.Sprintf("increment reaction count for user %d", userID))
	}

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
		_, err := s.postRepo.GetPostForReaction(ctx, targetPublicID)
		if err != nil {
			return nil, err
		}
	case "comment":
		err := s.commentRepo.EnsureCommentExists(ctx, targetPublicID)
		if err != nil {
			return nil, err
		}
	}

	return s.reactionRepo.GetByTargetPublicID(ctx, targetPublicID, targetType)
}

// CountReactions returns the count of likes and dislikes for a target.
// Uses a single optimized query that resolves the target once and counts both types.
func (s *Service) CountReactions(ctx context.Context, targetPublicID string, targetType string) (likes, dislikes int, err error) {
	// Validate inputs
	if targetType != "post" && targetType != "comment" {
		return 0, 0, domain.ErrInvalidTarget
	}

	return s.reactionRepo.CountLikesAndDislikesByTargetPublicID(ctx, targetPublicID, targetType)
}

// GetUserReactionCount returns the total number of reactions given by a user.
func (s *Service) GetUserReactionCount(ctx context.Context, userID int) (int, error) {
	if userID <= 0 {
		return 0, domain.ErrInvalidUserID
	}

	return s.reactionRepo.CountByUserID(ctx, userID)
}

// ListUserReactions returns all reactions made by a user, newest first.
func (s *Service) ListUserReactions(ctx context.Context, userID int) ([]*domain.Reaction, error) {
	if userID <= 0 {
		return nil, domain.ErrInvalidUserID
	}

	return s.reactionRepo.ListByUserID(ctx, userID)
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
		_, err := s.postRepo.GetPostForReaction(ctx, targetPublicID)
		if err != nil {
			return nil, err
		}
	case "comment":
		err := s.commentRepo.EnsureCommentExists(ctx, targetPublicID)
		if err != nil {
			return nil, err
		}
	}

	return s.reactionRepo.GetByUserAndTargetPublicID(ctx, userID, targetPublicID, targetType)
}
