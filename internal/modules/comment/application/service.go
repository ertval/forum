// Package application implements comment service business logic.
package application

import (
	"context"
	"forum/internal/modules/comment/domain"
	"forum/internal/modules/comment/ports"
)

// Service implements the CommentService interface.
type Service struct {
	commentRepo ports.CommentRepository
}

// NewService creates a new comment service.
func NewService(commentRepo ports.CommentRepository) *Service {
	return &Service{commentRepo: commentRepo}
}

// CreateComment creates a new comment.
// TODO: Implement comment creation with validation.
func (s *Service) CreateComment(ctx context.Context, postID, userID int, content string) (*domain.Comment, error) {
	// Implementation placeholder
	// 1. Validate content (non-empty, length limits)
	// 2. Create comment entity
	// 3. Save to repository
	// 4. Return created comment
	return nil, nil
}

// GetComment retrieves a comment by ID.
func (s *Service) GetComment(ctx context.Context, commentID int) (*domain.Comment, error) {
	return s.commentRepo.GetByID(ctx, commentID)
}

// UpdateComment updates a comment's content.
// TODO: Implement comment update with validation and authorization.
func (s *Service) UpdateComment(ctx context.Context, commentID int, content string) error {
	// Implementation placeholder
	// 1. Retrieve existing comment
	// 2. Validate new content
	// 3. Check authorization (user owns comment)
	// 4. Update comment
	return nil
}

// DeleteComment deletes a comment.
func (s *Service) DeleteComment(ctx context.Context, commentID int) error {
	return s.commentRepo.Delete(ctx, commentID)
}

// ListCommentsByPost retrieves all comments for a post.
func (s *Service) ListCommentsByPost(ctx context.Context, postID int) ([]*domain.Comment, error) {
	return s.commentRepo.ListByPostID(ctx, postID)
}
