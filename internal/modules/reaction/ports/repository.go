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

	// CountByUserID returns the total number of reactions given by a user.
	CountByUserID(ctx context.Context, userID int) (int, error)
}
