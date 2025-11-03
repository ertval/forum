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
	Create(ctx context.Context, comment *domain.Comment) error

	// GetByID retrieves a comment by its ID.
	GetByID(ctx context.Context, commentID int) (*domain.Comment, error)

	// Update updates an existing comment in the repository.
	Update(ctx context.Context, comment *domain.Comment) error

	// Delete removes a comment from the repository.
	Delete(ctx context.Context, commentID int) error

	// ListByPostID retrieves all comments for a specific post.
	ListByPostID(ctx context.Context, postID int) ([]*domain.Comment, error)
}
