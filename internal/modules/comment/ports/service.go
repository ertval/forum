// INPUT PORT - Service Interface
// Package ports defines the input ports (use cases) for the comment module.
package ports

import (
	"context"
	"forum/internal/modules/comment/domain"
)

// CommentService defines comment management use cases.
type CommentService interface {
	// CreateComment creates a new comment on a post.
	// postPublicID: public UUID of the post, userID: internal user ID from session
	CreateComment(ctx context.Context, postPublicID string, userID int, content string) (*domain.Comment, error)

	// GetComment retrieves a comment by its public UUID.
	GetComment(ctx context.Context, commentPublicID string) (*domain.Comment, error)

	// UpdateComment updates a comment's content.
	UpdateComment(ctx context.Context, commentPublicID string, content string) error

	// DeleteComment deletes a comment.
	DeleteComment(ctx context.Context, commentPublicID string) error

	// ListCommentsByPost retrieves all comments for a post.
	ListCommentsByPost(ctx context.Context, postPublicID string) ([]*domain.Comment, error)

	// ListCommentsByUser retrieves all comments made by a specific user.
	ListCommentsByUser(ctx context.Context, userPublicID string) ([]*domain.Comment, error)
}
