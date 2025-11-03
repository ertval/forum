package application
// Package application contains post application services.
package application

import (
	"context"

	"forum/internal/modules/post/domain"
	"forum/internal/modules/post/ports/input"
	"forum/internal/modules/post/ports/output"
)

// PostService implements the PostService interface.
type PostService struct {
	postRepo     output.PostRepository
	imageStorage output.ImageStorage
}

// NewPostService creates a new post service.
func NewPostService(postRepo output.PostRepository, imageStorage output.ImageStorage) input.PostService {
	return &PostService{
		postRepo:     postRepo,
		imageStorage: imageStorage,
	}
}

// Create creates a new post.
func (s *PostService) Create(ctx context.Context, post *domain.Post) error {
	// TODO: Implement
	return nil
}

// GetByID retrieves a post by ID.
func (s *PostService) GetByID(ctx context.Context, postID string) (*domain.Post, error) {
	// TODO: Implement
	return nil, nil
}

// List lists posts.
func (s *PostService) List(ctx context.Context, limit, offset int) ([]*domain.Post, error) {
	// TODO: Implement
	return nil, nil
}

// ListByCategory lists posts by category.
func (s *PostService) ListByCategory(ctx context.Context, categoryID string, limit, offset int) ([]*domain.Post, error) {
	// TODO: Implement
	return nil, nil
}

// ListByAuthor lists posts by author.
func (s *PostService) ListByAuthor(ctx context.Context, authorID string, limit, offset int) ([]*domain.Post, error) {
	// TODO: Implement
	return nil, nil
}

// Update updates a post.
func (s *PostService) Update(ctx context.Context, postID, userID string, title, content string) error {
	// TODO: Implement
	return nil
}

// Delete deletes a post.
func (s *PostService) Delete(ctx context.Context, postID, userID string) error {
	// TODO: Implement
	return nil
}

// CategoryService implements the CategoryService interface.
type CategoryService struct {
	categoryRepo output.CategoryRepository
}

// NewCategoryService creates a new category service.
func NewCategoryService(categoryRepo output.CategoryRepository) input.CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
	}
}

// Create creates a new category.
func (s *CategoryService) Create(ctx context.Context, category *domain.Category) error {
	// TODO: Implement
	return nil
}

// GetByID retrieves a category by ID.
func (s *CategoryService) GetByID(ctx context.Context, categoryID string) (*domain.Category, error) {
	// TODO: Implement
	return nil, nil
}

// List lists all categories.
func (s *CategoryService) List(ctx context.Context) ([]*domain.Category, error) {
	// TODO: Implement
	return nil, nil
}

// Delete deletes a category.
func (s *CategoryService) Delete(ctx context.Context, categoryID string) error {
	// TODO: Implement
	return nil
}
