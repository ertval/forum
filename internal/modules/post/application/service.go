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
    return s.postRepo.List(ctx, filter)
}
