// Package application implements the post service business logic.
package application

import (
    "context"
    "forum/internal/modules/post/domain"
    "forum/internal/modules/post/ports"
)

// Service implements the PostService interface.
type Service struct {
    postRepo     ports.PostRepository
    categoryRepo ports.CategoryRepository
}

// NewService creates a new post service.
func NewService(postRepo ports.PostRepository, categoryRepo ports.CategoryRepository) *Service {
    return &Service{
        postRepo:     postRepo,
        categoryRepo: categoryRepo,
    }
}

// CreatePost creates a new post.
// TODO: Implement post creation with image upload.
func (s *Service) CreatePost(ctx context.Context, userID int, title, content string, categories []string, image []byte) (*domain.Post, error) {
    // Implementation placeholder
    // 1. Validate title, content, and categories
    // 2. Verify categories exist in database
    // 3. If image provided, validate format (JPEG/PNG/GIF) and size (< 20MB)
    // 4. Save image to static/uploads/ with unique filename
    // 5. Create post entity with image URL
    // 6. Save to repository
    // 7. Associate categories with post
    // 8. Return created post
    return nil, nil
}

// GetPost retrieves a post by ID.
func (s *Service) GetPost(ctx context.Context, postID int) (*domain.Post, error) {
    return s.postRepo.GetByID(ctx, postID)
}

// UpdatePost updates a post.
// TODO: Implement post update.
func (s *Service) UpdatePost(ctx context.Context, postID int, title, content string) error {
    // Implementation placeholder
    // 1. Retrieve existing post
    // 2. Validate new title and content
    // 3. Update post entity
    // 4. Save to repository
    return nil
}

// DeletePost deletes a post.
// TODO: Implement post deletion.
func (s *Service) DeletePost(ctx context.Context, postID int) error {
    return s.postRepo.Delete(ctx, postID)
}

// ListPosts lists posts with optional filters.
// TODO: Implement post filtering.
func (s *Service) ListPosts(ctx context.Context, filter ports.PostFilter) ([]*domain.Post, error) {
    // Implementation placeholder
    // 1. Apply filters (by category, by user, by liked posts)
    // 2. Apply pagination (offset, limit)
    // 3. Return filtered posts
    return s.postRepo.List(ctx, filter)
}
