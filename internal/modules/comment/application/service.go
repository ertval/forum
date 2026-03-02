// Package application implements comment service business logic.
package application

import (
	"context"
	"fmt"
	"log"
	"time"

	"forum/internal/modules/comment/domain"
	"forum/internal/modules/comment/ports"
)

// Notification type constants mirrored locally to avoid cross-module port imports.
const notifTypeComment = "comment"

type PostRecord struct {
	ID       int
	PublicID string
	UserID   int
}

// postService defines the minimal post operations required by the comment service.
// This avoids a direct import of the post module's ports package.
type postService interface {
	GetPostForComment(ctx context.Context, postID string) (*PostRecord, error)
}

// userService defines the minimal user operations required by the comment service.
// This avoids a direct import of the user module's ports package.
type userService interface {
	ResolveUserIDByPublicID(ctx context.Context, publicID string) (int, error)
	IncrementCommentCount(ctx context.Context, userID int) error
	DecrementCommentCount(ctx context.Context, userID int) error
}

// notificationService defines the minimal notification operations required by the comment service.
// This avoids a direct import of the notification module's ports package.
type notificationService interface {
	CreateNotification(ctx context.Context, userID, actorID int, notifType, message string, targetPublicID string) error
}

// Service implements the CommentService interface.
type Service struct {
	commentRepo         ports.CommentRepository
	postService         postService
	userService         userService
	notificationService notificationService
}

// NewService creates a new comment service.
func NewService(commentRepo ports.CommentRepository, postService postService, userService userService, notificationService notificationService) *Service {
	return &Service{
		commentRepo:         commentRepo,
		postService:         postService,
		userService:         userService,
		notificationService: notificationService,
	}
}

// CreateComment creates a new comment.
func (s *Service) CreateComment(ctx context.Context, postPublicID string, userID int, content string) (*domain.Comment, error) {
	// Get the post to get its internal ID
	post, err := s.postService.GetPostForComment(ctx, postPublicID)
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

	if s.notificationService != nil && post.UserID != userID {
		message := "Someone commented on your post"
		if err := s.notificationService.CreateNotification(ctx, post.UserID, userID, notifTypeComment, message, post.PublicID); err != nil {
			log.Printf("WARNING: failed to create comment notification for post owner %d: %v", post.UserID, err)
		}
	}

	runInBackground(fmt.Sprintf("increment comment count for user %d", userID), func() error {
		return s.userService.IncrementCommentCount(context.Background(), userID)
	})

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

	runInBackground(fmt.Sprintf("decrement comment count for user %d", comment.UserID), func() error {
		return s.userService.DecrementCommentCount(context.Background(), comment.UserID)
	})

	return nil
}

// ListCommentsByPost retrieves all comments for a post.
func (s *Service) ListCommentsByPost(ctx context.Context, postPublicID string) ([]*domain.Comment, error) {
	return s.commentRepo.ListByPostPublicID(ctx, postPublicID)
}

// ListCommentsByUser retrieves all comments made by a specific user.
func (s *Service) ListCommentsByUser(ctx context.Context, userPublicID string) ([]*domain.Comment, error) {
	// First get the internal user ID from the public ID
	userID, err := s.userService.ResolveUserIDByPublicID(ctx, userPublicID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Call the repository to get comments by user ID
	return s.commentRepo.ListByUser(ctx, userID)
}

// ListCommentsByUserPaginated retrieves comments made by a user with pagination.
func (s *Service) ListCommentsByUserPaginated(ctx context.Context, userPublicID string, limit, offset int) ([]*domain.Comment, error) {
	// First get the internal user ID from the public ID
	userID, err := s.userService.ResolveUserIDByPublicID(ctx, userPublicID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Call the repository to get comments by user ID with pagination
	return s.commentRepo.ListByUserPaginated(ctx, userID, limit, offset)
}

func runInBackground(action string, fn func() error) {
	go func() {
		if err := fn(); err != nil {
			log.Printf("WARNING: background action failed (%s): %v", action, err)
		}
	}()
}
