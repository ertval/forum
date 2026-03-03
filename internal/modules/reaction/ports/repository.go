// OUTPUT PORT - Repository Interface
// Package ports defines the output ports (data access contracts) for the reaction module.
package ports

import (
	"context"
	"forum/internal/modules/reaction/domain"
)

// ReactionRepository defines the data access contract for reactions.
type ReactionRepository interface {
	// Create stores a new reaction in the repository.
	// Must generate and set PublicID (UUID) before persisting.
	Create(ctx context.Context, reaction *domain.Reaction) error

	// DeleteByTargetPublicID removes a user's reaction from a target by target's public UUID.
	DeleteByTargetPublicID(ctx context.Context, userID int, targetPublicID string, targetType string) error

	// GetByTargetPublicID retrieves all reactions for a specific target by its public UUID.
	GetByTargetPublicID(ctx context.Context, targetPublicID string, targetType string) ([]*domain.Reaction, error)

	// GetByUserAndTargetPublicID retrieves a user's reaction for a specific target by target's public UUID.
	GetByUserAndTargetPublicID(ctx context.Context, userID int, targetPublicID string, targetType string) (*domain.Reaction, error)

	// CountByTargetPublicID returns the number of reactions of a specific type for a target by its public UUID.
	CountByTargetPublicID(ctx context.Context, targetPublicID string, targetType string, reactionType domain.ReactionType) (int, error)

	// CountLikesAndDislikesByTargetPublicID returns both likes and dislikes counts in a single query.
	CountLikesAndDislikesByTargetPublicID(ctx context.Context, targetPublicID string, targetType string) (likes, dislikes int, err error)

	// CountByUserID returns the total number of reactions given by a user.
	CountByUserID(ctx context.Context, userID int) (int, error)

	// ListByUserID returns all reactions made by a user, newest first.
	// Each returned reaction must include PublicTargetID.
	ListByUserID(ctx context.Context, userID int) ([]*domain.Reaction, error)

	// CountBatchByTargetPublicIDs returns like and dislike counts for multiple targets in a single query.
	// The result is a map keyed by targetPublicID, with inner maps of reaction type ("like"/"dislike") -> count.
	CountBatchByTargetPublicIDs(ctx context.Context, targetPublicIDs []string, targetType string) (map[string]map[string]int, error)

	// ToggleReaction atomically handles the full reaction toggle flow in a single transaction.
	// It resolves the target, checks for an existing reaction, and either:
	// - Deletes the reaction if the same type already exists (toggle off, removed=true)
	// - Updates the reaction type if a different type exists (removed=false)
	// - Creates a new reaction if none exists (removed=false)
	ToggleReaction(ctx context.Context, reaction *domain.Reaction) (action domain.ToggleAction, err error)
}
