// Package application implements the category service business logic.
package application

import (
	"context"

	"forum/internal/modules/post/domain"
	"forum/internal/modules/post/ports"
)

// CategoryService implements the CategoryService interface.
type CategoryService struct {
	categoryRepo ports.CategoryRepository
}

// NewCategoryService creates a new category service.
func NewCategoryService(categoryRepo ports.CategoryRepository) *CategoryService {
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

	// Validate category
	if err := category.Validate(); err != nil {
		return nil, err
	}

	// Repository will generate both internal ID and public_id
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
