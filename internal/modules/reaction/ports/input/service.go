package input
// Package input defines the inbound ports for the reaction module.
package input

import (
	"context"

	"forum/internal/modules/reaction/domain"
)

// ReactionService defines the reaction management use cases.
type ReactionService interface {
	// AddReaction adds or updates a reaction.
	AddReaction(ctx context.Context, userID, targetID string, targetType domain.TargetType, reactionType domain.ReactionType) error

	// RemoveReaction removes a reaction.
	RemoveReaction(ctx context.Context, userID, targetID string, targetType domain.TargetType) error

	// GetReaction gets a user's reaction to a target.
	GetReaction(ctx context.Context, userID, targetID string, targetType domain.TargetType) (*domain.Reaction, error)

	// CountReactions counts likes and dislikes for a target.
	CountReactions(ctx context.Context, targetID string, targetType domain.TargetType) (likes, dislikes int, err error)

	// ListLikedPosts lists posts liked by a user.
	ListLikedPosts(ctx context.Context, userID string, limit, offset int) ([]string, error)
}
