package output
// Package output defines the outbound ports for the comment module.
package output

import (
	"context"

	"forum/internal/modules/comment/domain"
)

// CommentRepository defines the interface for comment persistence.
type CommentRepository interface {
	Create(ctx context.Context, comment *domain.Comment) error
	GetByID(ctx context.Context, commentID string) (*domain.Comment, error)
	ListByPost(ctx context.Context, postID string) ([]*domain.Comment, error)
	Update(ctx context.Context, comment *domain.Comment) error
	Delete(ctx context.Context, commentID string) error
}
