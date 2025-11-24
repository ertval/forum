// Package application implements the post service business logic.
package application

import (
	"context"
	"forum/internal/modules/post/domain"
	"forum/internal/modules/post/ports"
	userPorts "forum/internal/modules/user/ports"
	"time"

	"github.com/gofrs/uuid/v5"
)

// Service implements the PostService interface.
type Service struct {
	postRepo     ports.PostRepository
	categoryRepo ports.CategoryRepository
	userService  userPorts.UserService
}

// NewService creates a new post service.
func NewService(postRepo ports.PostRepository, categoryRepo ports.CategoryRepository, userService userPorts.UserService) *Service {
	return &Service{
		postRepo:     postRepo,
		categoryRepo: categoryRepo,
		userService:  userService,
	}
}

// generateID generates a new UUID for entities.
func generateID() string {
	id, _ := uuid.NewV4()
	return id.String()
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

	// Verify all categories exist
	for _, categoryName := range categories {
		_, err := s.categoryRepo.GetByName(ctx, categoryName)
		if err != nil {
			return nil, err
		}
	}

	// TODO: Handle image upload when needed
	// If image provided:
	// 1. Validate format (JPEG/PNG/GIF) and size (< 20MB)
	// 2. Save to static/uploads/ with unique filename
	// 3. Set post.ImageURL

	// Save post to repository
	if err := s.postRepo.Create(ctx, post); err != nil {
		return nil, err
	}

	// Increment user's post count asynchronously (non-blocking)
	go func() {
		_ = s.userService.IncrementPostCount(context.Background(), userID)
	}()

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

	// Verify all categories exist
	for _, categoryName := range categories {
		_, err := s.categoryRepo.GetByName(ctx, categoryName)
		if err != nil {
			return err
		}
	}

	// Save to repository
	return s.postRepo.Update(ctx, post)
}

// DeletePost deletes a post.
func (s *Service) DeletePost(ctx context.Context, postID string) error {
	// Get the post first to retrieve the user ID
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return err
	}

	// Delete the post
	if err := s.postRepo.Delete(ctx, postID); err != nil {
		return err
	}

	// Decrement user's post count asynchronously (non-blocking)
	go func() {
		_ = s.userService.DecrementPostCount(context.Background(), post.UserID)
	}()

	return nil
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
