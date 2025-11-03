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
	Create(ctx context.Context, reaction *domain.Reaction) error

	// Delete removes a user's reaction from a target.
	Delete(ctx context.Context, userID, targetID int, targetType string) error

	// GetByTarget retrieves all reactions for a specific target.
	GetByTarget(ctx context.Context, targetID int, targetType string) ([]*domain.Reaction, error)

	// GetByUserAndTarget retrieves a user's reaction for a specific target.
	GetByUserAndTarget(ctx context.Context, userID, targetID int, targetType string) (*domain.Reaction, error)

	// Count returns the number of reactions of a specific type for a target.
	Count(ctx context.Context, targetID int, targetType string, reactionType domain.ReactionType) (int, error)
}
