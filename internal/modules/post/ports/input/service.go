package input
// Package input defines the inbound ports for the post module.
package input

import (
	"context"

	"forum/internal/modules/post/domain"
)

// PostService defines the post management use cases.
type PostService interface {
	// Create creates a new post.
	Create(ctx context.Context, post *domain.Post) error

	// GetByID retrieves a post by ID.
	GetByID(ctx context.Context, postID string) (*domain.Post, error)

	// List lists posts with pagination.
	List(ctx context.Context, limit, offset int) ([]*domain.Post, error)

	// ListByCategory lists posts by category.
	ListByCategory(ctx context.Context, categoryID string, limit, offset int) ([]*domain.Post, error)

	// ListByAuthor lists posts by author.
	ListByAuthor(ctx context.Context, authorID string, limit, offset int) ([]*domain.Post, error)

	// Update updates a post.
	Update(ctx context.Context, postID, userID string, title, content string) error

	// Delete deletes a post.
	Delete(ctx context.Context, postID, userID string) error
}

// CategoryService defines the category management use cases.
type CategoryService interface {
	// Create creates a new category.
	Create(ctx context.Context, category *domain.Category) error

	// GetByID retrieves a category by ID.
	GetByID(ctx context.Context, categoryID string) (*domain.Category, error)

	// List lists all categories.
	List(ctx context.Context) ([]*domain.Category, error)

	// Delete deletes a category.
	Delete(ctx context.Context, categoryID string) error
}
