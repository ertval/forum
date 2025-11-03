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
	CreateComment(ctx context.Context, postID, userID int, content string) (*domain.Comment, error)

	// GetComment retrieves a comment by ID.
	GetComment(ctx context.Context, commentID int) (*domain.Comment, error)

	// UpdateComment updates a comment's content.
	UpdateComment(ctx context.Context, commentID int, content string) error

	// DeleteComment deletes a comment.
	DeleteComment(ctx context.Context, commentID int) error

	// ListCommentsByPost retrieves all comments for a post.
	ListCommentsByPost(ctx context.Context, postID int) ([]*domain.Comment, error)
}
