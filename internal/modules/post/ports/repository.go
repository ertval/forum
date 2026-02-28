// OUTPUT PORT - Repository Interface
// Package ports defines the output ports for the post module.
package ports

import (
	"context"
	"forum/internal/modules/post/domain"
)

// PostRepository defines data access for posts.
type PostRepository interface {
	// Create persists a new post to the database.
	Create(ctx context.Context, post *domain.Post) error
	// GetByID retrieves a post by its public UUID.
	GetByID(ctx context.Context, postID string) (*domain.Post, error)
	// Update modifies an existing post in the database.
	Update(ctx context.Context, post *domain.Post) error
	// Delete removes a post by its public UUID.
	Delete(ctx context.Context, postID string) error
	// List returns posts matching the given filter criteria.
	List(ctx context.Context, filter domain.PostFilter) ([]*domain.Post, error)
	// UpdateImagePath updates only the image_path field for a post.
	UpdateImagePath(ctx context.Context, postID string, imagePath string) error
	// GetImagePath retrieves the image_path for a post by its public ID.
	GetImagePath(ctx context.Context, postID string) (string, error)
}

// CategoryRepository defines data access for categories.
type CategoryRepository interface {
	// Create persists a new category to the database.
	Create(ctx context.Context, category *domain.Category) error
	// GetByID retrieves a category by its public UUID.
	GetByID(ctx context.Context, categoryID string) (*domain.Category, error)
	// GetByName retrieves a category by its name.
	GetByName(ctx context.Context, name string) (*domain.Category, error)
	// List returns all categories.
	List(ctx context.Context) ([]*domain.Category, error)
	// Delete removes a category by its public UUID.
	Delete(ctx context.Context, categoryID string) error
}

// NOTE: ImageHandler interface is defined in image.go to maintain separation of concerns.
