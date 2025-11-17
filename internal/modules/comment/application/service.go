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
func (s *Service) CreateComment(ctx context.Context, postPublicID string, userID int, content string) (*domain.Comment, error) {
	// Implementation placeholder
	// 1. Validate content (non-empty, length limits)
	// 2. Resolve postPublicID to internal post ID
	// 3. Create comment entity
	// 4. Save to repository (repo generates PublicID)
	// 5. Return created comment
	return nil, nil
}

// GetComment retrieves a comment by its public UUID.
func (s *Service) GetComment(ctx context.Context, commentPublicID string) (*domain.Comment, error) {
	return s.commentRepo.GetByPublicID(ctx, commentPublicID)
}

// UpdateComment updates a comment's content.
// TODO: Implement comment update with validation and authorization.
func (s *Service) UpdateComment(ctx context.Context, commentPublicID string, content string) error {
	// Implementation placeholder
	// 1. Retrieve existing comment by public ID
	// 2. Validate new content
	// 3. Check authorization (user owns comment)
	// 4. Update comment
	return nil
}

// DeleteComment deletes a comment.
func (s *Service) DeleteComment(ctx context.Context, commentPublicID string) error {
	return s.commentRepo.DeleteByPublicID(ctx, commentPublicID)
}

// ListCommentsByPost retrieves all comments for a post.
func (s *Service) ListCommentsByPost(ctx context.Context, postPublicID string) ([]*domain.Comment, error) {
	return s.commentRepo.ListByPostPublicID(ctx, postPublicID)
}
