// INPUT PORT - Service Interface
// Package ports defines the input ports (use cases) for the reaction module.
package ports

import (
	"context"
	"forum/internal/modules/reaction/domain"
)

// ReactionService defines reaction management use cases.
type ReactionService interface {
	// React adds or updates a user's reaction to a target.
	// If the user already has a different reaction, it's replaced (toggle behavior).
	// targetPublicID: public UUID of the target (post/comment), userID: internal user ID from session
	React(ctx context.Context, userID int, targetPublicID string, targetType string, reactionType domain.ReactionType) error

	// RemoveReaction removes a user's reaction from a target.
	RemoveReaction(ctx context.Context, userID int, targetPublicID string, targetType string) error

	// GetReactions retrieves all reactions for a target.
	GetReactions(ctx context.Context, targetPublicID string, targetType string) ([]*domain.Reaction, error)

	// CountReactions returns the count of likes and dislikes for a target.
	CountReactions(ctx context.Context, targetPublicID string, targetType string) (likes, dislikes int, err error)

	// GetUserReactionCount returns the total number of reactions given by a user.
	GetUserReactionCount(ctx context.Context, userID int) (int, error)

	// ListUserReactions returns all reactions made by a user, newest first.
	ListUserReactions(ctx context.Context, userID int) ([]*domain.Reaction, error)

	// GetByUserAndTargetPublicID retrieves a user's reaction for a specific target by target's public UUID.
	GetByUserAndTargetPublicID(ctx context.Context, userID int, targetPublicID string, targetType string) (*domain.Reaction, error)

	// CountReactionsBatch returns likes/dislikes counts for multiple targets in a single query.
	// The result is a map keyed by targetPublicID, with inner maps of reaction type ("like"/"dislike") -> count.
	CountReactionsBatch(ctx context.Context, targetPublicIDs []string, targetType string) (map[string]map[string]int, error)
}
