// Package application implements comment service business logic.
package application

import (
	"context"
	"fmt"
	"forum/internal/modules/comment/domain"
	"forum/internal/modules/comment/ports"
	postPorts "forum/internal/modules/post/ports"
	userPorts "forum/internal/modules/user/ports"
	"time"
)

// Service implements the CommentService interface.
type Service struct {
	commentRepo ports.CommentRepository
	postService postPorts.PostService
	userService userPorts.UserService
}

// NewService creates a new comment service.
func NewService(commentRepo ports.CommentRepository, postService postPorts.PostService, userService userPorts.UserService) *Service {
	return &Service{
		commentRepo: commentRepo,
		postService: postService,
		userService: userService,
	}
}

// CreateComment creates a new comment.
func (s *Service) CreateComment(ctx context.Context, postPublicID string, userID int, content string) (*domain.Comment, error) {
	// Get the post to get its internal ID
	post, err := s.postService.GetPost(ctx, postPublicID)
	if err != nil {
		return nil, fmt.Errorf("post not found: %w", err)
	}

	// Create comment entity with provided data
	comment := &domain.Comment{
		PostID:       post.ID, // Using post's internal ID
		UserID:       userID,
		Content:      content,
		PublicPostID: postPublicID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Validate comment
	if err := comment.Validate(); err != nil {
		return nil, err
	}

	// Save to repository (repository generates PublicID)
	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return nil, err
	}

	// Increment user's comment count asynchronously (non-blocking)
	go func() {
		_ = s.userService.IncrementCommentCount(context.Background(), userID)
	}()

	return comment, nil
}

// GetComment retrieves a comment by its public UUID.
func (s *Service) GetComment(ctx context.Context, commentPublicID string) (*domain.Comment, error) {
	return s.commentRepo.GetByPublicID(ctx, commentPublicID)
}

// UpdateComment updates a comment's content.
func (s *Service) UpdateComment(ctx context.Context, commentPublicID string, content string) error {
	// Retrieve existing comment by public ID
	existingComment, err := s.commentRepo.GetByPublicID(ctx, commentPublicID)
	if err != nil {
		return err
	}

	// Validate new content
	updatedComment := &domain.Comment{
		ID:        existingComment.ID,
		PublicID:  existingComment.PublicID,
		PostID:    existingComment.PostID,
		UserID:    existingComment.UserID,
		Content:   content,
		CreatedAt: existingComment.CreatedAt,
		UpdatedAt: time.Now(),
	}

	if err := updatedComment.Validate(); err != nil {
		return err
	}

	// Check authorization (user owns comment)
	// Note: In a real implementation, we'd pass the authenticated user ID
	// to compare with existingComment.UserID, but for now we'll assume
	// authorization was already checked by the HTTP handler

	// Update comment
	return s.commentRepo.Update(ctx, updatedComment)
}

// DeleteComment deletes a comment.
func (s *Service) DeleteComment(ctx context.Context, commentPublicID string) error {
	// Get the comment first to retrieve the user ID
	comment, err := s.commentRepo.GetByPublicID(ctx, commentPublicID)
	if err != nil {
		return err
	}

	// Delete the comment
	if err := s.commentRepo.DeleteByPublicID(ctx, commentPublicID); err != nil {
		return err
	}

	// Decrement user's comment count asynchronously (non-blocking)
	go func() {
		_ = s.userService.DecrementCommentCount(context.Background(), comment.UserID)
	}()

	return nil
}

// ListCommentsByPost retrieves all comments for a post.
func (s *Service) ListCommentsByPost(ctx context.Context, postPublicID string) ([]*domain.Comment, error) {
	return s.commentRepo.ListByPostPublicID(ctx, postPublicID)
}
