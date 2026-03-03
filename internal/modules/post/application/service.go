// Package application implements the post service business logic.
package application

import (
	"context"
	"fmt"
	"time"

	"forum/internal/modules/post/domain"
	"forum/internal/modules/post/ports"
	"forum/internal/platform/async"
)

// userService defines the minimal user operations required by the post service.
// This avoids a direct import of the user module's ports package.
type userService interface {
	IncrementPostCount(ctx context.Context, userID int) error
	DecrementPostCount(ctx context.Context, userID int) error
}

// Service implements the PostService interface.
type Service struct {
	postRepo     ports.PostRepository
	categoryRepo ports.CategoryRepository
	userService  userService
	imageHandler ports.ImageHandler
	maxImageSize int64
}

// NewService creates a new post service.
func NewService(postRepo ports.PostRepository, categoryRepo ports.CategoryRepository, userService userService, imageHandler ports.ImageHandler, maxImageSize int64) *Service {
	return &Service{
		postRepo:     postRepo,
		categoryRepo: categoryRepo,
		userService:  userService,
		imageHandler: imageHandler,
		maxImageSize: maxImageSize,
	}
}

// MaxImageSize returns the maximum allowed image size in bytes.
func (s *Service) MaxImageSize() int64 {
	return s.maxImageSize
}

// CreatePost creates a new post.
func (s *Service) CreatePost(ctx context.Context, userID int, title, content string, categories []string, image []byte) (*domain.Post, error) {
	// Create post entity - repository will generate both internal ID and public_id
	post := &domain.Post{
		UserID:     userID,
		Title:      title,
		Content:    content,
		Categories: categories,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Validate post
	if err := post.Validate(); err != nil {
		return nil, err
	}

	// Verify all categories exist (batch lookup)
	foundCats, err := s.categoryRepo.GetByNames(ctx, categories)
	if err != nil {
		return nil, err
	}
	if len(foundCats) != len(categories) {
		return nil, domain.ErrCategoryNotFound
	}

	// Handle image upload if provided
	var savedFilename string
	if len(image) > 0 && s.imageHandler != nil {
		filename, err := s.imageHandler.Save(image)
		if err != nil {
			return nil, err
		}
		savedFilename = filename
		post.ImageURL = filename // Store just the filename, repository prepends path
	}

	// Save post to repository
	if err := s.postRepo.Create(ctx, post); err != nil {
		// Rollback: delete saved image if DB insert fails
		if savedFilename != "" && s.imageHandler != nil {
			_ = s.imageHandler.Delete(savedFilename)
		}
		return nil, err
	}

	// Increment user's post count asynchronously (non-blocking)
	async.Run(func(ctx context.Context) error {
		return s.userService.IncrementPostCount(ctx, userID)
	}, fmt.Sprintf("increment post count for user %d", userID))

	return post, nil
}

// GetPost retrieves a post by ID.
func (s *Service) GetPost(ctx context.Context, postID string) (*domain.Post, error) {
	return s.postRepo.GetByID(ctx, postID)
}

// UpdatePost updates a post.
func (s *Service) UpdatePost(ctx context.Context, postID string, title, content string, categories []string) error {
	// Retrieve existing post
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return err
	}

	// Update fields
	post.Title = title
	post.Content = content
	post.Categories = categories
	post.UpdatedAt = time.Now()

	// Validate updated post
	if err := post.Validate(); err != nil {
		return err
	}

	// Verify all categories exist (batch lookup)
	foundCats, err := s.categoryRepo.GetByNames(ctx, categories)
	if err != nil {
		return err
	}
	if len(foundCats) != len(categories) {
		return domain.ErrCategoryNotFound
	}

	// Save to repository
	return s.postRepo.Update(ctx, post)
}

// DeletePost deletes a post.
func (s *Service) DeletePost(ctx context.Context, postID string) error {
	// Get the post first to retrieve the user ID and image path
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return err
	}

	// Get the stored image path before deletion
	imagePath, _ := s.postRepo.GetImagePath(ctx, postID)

	// Delete the post
	if err := s.postRepo.Delete(ctx, postID); err != nil {
		return err
	}

	// Delete associated image file if present
	if imagePath != "" && s.imageHandler != nil {
		_ = s.imageHandler.Delete(imagePath) // Best effort, don't fail if image deletion fails
	}

	// Decrement user's post count asynchronously (non-blocking)
	async.Run(func(ctx context.Context) error {
		return s.userService.DecrementPostCount(ctx, post.UserID)
	}, fmt.Sprintf("decrement post count for user %d", post.UserID))

	return nil
}

// ListPosts lists posts with optional filters.
func (s *Service) ListPosts(ctx context.Context, filter domain.PostFilter) ([]*domain.Post, error) {
	return s.postRepo.List(ctx, filter)
}

// UpdatePostImage updates or removes the image for a post.
func (s *Service) UpdatePostImage(ctx context.Context, postID string, image []byte, removeImage bool) error {
	// Verify post exists
	_, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return err
	}

	// Get current image path
	oldImagePath, _ := s.postRepo.GetImagePath(ctx, postID)

	// Handle removal
	if removeImage && len(image) == 0 {
		// Remove image from DB
		if err := s.postRepo.UpdateImagePath(ctx, postID, ""); err != nil {
			return err
		}
		// Delete file if exists
		if oldImagePath != "" && s.imageHandler != nil {
			_ = s.imageHandler.Delete(oldImagePath)
		}
		return nil
	}

	// Handle new image upload
	if len(image) > 0 && s.imageHandler != nil {
		// Save new image
		newFilename, err := s.imageHandler.Save(image)
		if err != nil {
			return err
		}

		// Update DB with new image path
		if err := s.postRepo.UpdateImagePath(ctx, postID, newFilename); err != nil {
			// Rollback: delete newly saved image
			_ = s.imageHandler.Delete(newFilename)
			return err
		}

		// Delete old image if it existed
		if oldImagePath != "" {
			_ = s.imageHandler.Delete(oldImagePath)
		}
	}

	return nil
}
