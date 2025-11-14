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

// CategoryService implements the CategoryService interface.
type CategoryService struct {
	categoryRepo ports.CategoryRepository
}

// NewCategoryService creates a new category service.
func NewCategoryService(categoryRepo ports.CategoryRepository) ports.CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
	}
}

// Create creates a new category.
func (s *CategoryService) Create(ctx context.Context, name, description string) (*domain.Category, error) {
	category := &domain.Category{
		Name:        name,
		Description: description,
	}

	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

// Get retrieves a category by ID.
func (s *CategoryService) Get(ctx context.Context, categoryID string) (*domain.Category, error) {
	return s.categoryRepo.GetByID(ctx, categoryID)
}

// List retrieves all categories.
func (s *CategoryService) List(ctx context.Context) ([]*domain.Category, error) {
	return s.categoryRepo.List(ctx)
}

// Delete deletes a category.
func (s *CategoryService) Delete(ctx context.Context, categoryID string) error {
	return s.categoryRepo.Delete(ctx, categoryID)
}

// CreatePost creates a new post.
// TODO: Implement post creation with image upload.
func (s *Service) CreatePost(ctx context.Context, userID string, title, content string, categories []string, image []byte) (*domain.Post, error) {
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
func (s *Service) GetPost(ctx context.Context, postID string) (*domain.Post, error) {
	return s.postRepo.GetByID(ctx, postID)
}

// UpdatePost updates a post.
// TODO: Implement post update.
func (s *Service) UpdatePost(ctx context.Context, postID string, title, content string) error {
	// Implementation placeholder
	// 1. Retrieve existing post
	// 2. Validate new title and content
	// 3. Update post entity
	// 4. Save to repository
	return nil
}

// DeletePost deletes a post.
// TODO: Implement post deletion.
func (s *Service) DeletePost(ctx context.Context, postID string) error {
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
