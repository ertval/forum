// OUTPUT PORT - Repository Interface
// Package ports defines the output ports (data access contracts) for the comment module.
package ports

import (
	"context"
	"forum/internal/modules/comment/domain"
)

// CommentRepository defines the data access contract for comments.
type CommentRepository interface {
	// Create stores a new comment in the repository.
	// Must generate and set PublicID (UUID) before persisting.
	Create(ctx context.Context, comment *domain.Comment) error

	// GetByPublicID retrieves a comment by its public UUID.
	GetByPublicID(ctx context.Context, commentPublicID string) (*domain.Comment, error)

	// Update updates an existing comment in the repository.
	// Uses internal ID from the comment entity.
	Update(ctx context.Context, comment *domain.Comment) error

	// DeleteByPublicID removes a comment by its public UUID.
	DeleteByPublicID(ctx context.Context, commentPublicID string) error

	// ListByPostPublicID retrieves all comments for a specific post by post's public UUID.
	ListByPostPublicID(ctx context.Context, postPublicID string) ([]*domain.Comment, error)

	// ListByUser retrieves all comments made by a specific user by internal user ID.
	ListByUser(ctx context.Context, userID int) ([]*domain.Comment, error)
}
