package input
// Package input defines the inbound ports for the comment module.
package input

import (
	"context"

	"forum/internal/modules/comment/domain"
)

// CommentService defines the comment management use cases.
type CommentService interface {
	// Create creates a new comment.
	Create(ctx context.Context, comment *domain.Comment) error

	// GetByID retrieves a comment by ID.
	GetByID(ctx context.Context, commentID string) (*domain.Comment, error)

	// ListByPost lists comments for a post.
	ListByPost(ctx context.Context, postID string) ([]*domain.Comment, error)

	// Update updates a comment.
	Update(ctx context.Context, commentID, userID, content string) error

	// Delete deletes a comment.
	Delete(ctx context.Context, commentID, userID string) error
}
