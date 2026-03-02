// Package application implements reaction service business logic.
package application

import (
	"context"
	"fmt"
	"log"
	"time"

	"forum/internal/modules/reaction/domain"
	"forum/internal/modules/reaction/ports"
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
	removed, err := s.reactionRepo.ToggleReaction(ctx, reaction)
	if err != nil {
		return err
	}

	if removed {
		// Reaction was toggled off — decrement user's reaction count
		runInBackground(fmt.Sprintf("decrement reaction count for user %d", userID), func() error {
			return s.userService.DecrementReactionCount(context.Background(), userID)
		})
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

	// Increment user's reaction count asynchronously (non-blocking)
	runInBackground(fmt.Sprintf("increment reaction count for user %d", userID), func() error {
		return s.userService.IncrementReactionCount(context.Background(), userID)
	})

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
	runInBackground(fmt.Sprintf("decrement reaction count for user %d", userID), func() error {
		return s.userService.DecrementReactionCount(context.Background(), userID)
	})

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
func (s *Service) CountReactions(ctx context.Context, targetPublicID string, targetType string) (likes, dislikes int, err error) {
	// Validate inputs
	if targetType != "post" && targetType != "comment" {
		return 0, 0, domain.ErrInvalidTarget
	}

	// Verify target exists
	switch targetType {
	case "post":
		_, err := s.postRepo.GetPostForReaction(ctx, targetPublicID)
		if err != nil {
			return 0, 0, err
		}
	case "comment":
		err := s.commentRepo.EnsureCommentExists(ctx, targetPublicID)
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

func runInBackground(action string, fn func() error) {
	go func() {
		if err := fn(); err != nil {
			log.Printf("WARNING: background action failed (%s): %v", action, err)
		}
	}()
}
