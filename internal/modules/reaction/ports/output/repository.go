package output
// Package output defines the outbound ports for the reaction module.
package output

import (
	"context"

	"forum/internal/modules/reaction/domain"
)

// ReactionRepository defines the interface for reaction persistence.
type ReactionRepository interface {
	Create(ctx context.Context, reaction *domain.Reaction) error
	GetByUserAndTarget(ctx context.Context, userID, targetID string, targetType domain.TargetType) (*domain.Reaction, error)
	Update(ctx context.Context, reaction *domain.Reaction) error
	Delete(ctx context.Context, userID, targetID string, targetType domain.TargetType) error
	CountByTarget(ctx context.Context, targetID string, targetType domain.TargetType) (likes, dislikes int, err error)
	ListLikedPostsByUser(ctx context.Context, userID string, limit, offset int) ([]string, error)
}
