// OUTPUT PORT - Repository Interface
// Package ports defines the output ports for the post module.
package ports

import (
	"context"
	"forum/internal/modules/post/domain"
)

// PostRepository defines data access for posts.
type PostRepository interface {
	Create(ctx context.Context, post *domain.Post) error
	GetByID(ctx context.Context, postID string) (*domain.Post, error)
	Update(ctx context.Context, post *domain.Post) error
	Delete(ctx context.Context, postID string) error
	List(ctx context.Context, filter PostFilter) ([]*domain.Post, error)
	// UpdateImagePath updates only the image_path field for a post.
	UpdateImagePath(ctx context.Context, postID string, imagePath string) error
	// GetImagePath retrieves the image_path for a post by its public ID.
	GetImagePath(ctx context.Context, postID string) (string, error)
}

// CategoryRepository defines data access for categories.
type CategoryRepository interface {
	Create(ctx context.Context, category *domain.Category) error
	GetByID(ctx context.Context, categoryID string) (*domain.Category, error)
	GetByName(ctx context.Context, name string) (*domain.Category, error)
	List(ctx context.Context) ([]*domain.Category, error)
	Delete(ctx context.Context, categoryID string) error
}

// ImageHandler defines file operations for image uploads.
type ImageHandler interface {
	// Save saves image data and returns the filename (without path prefix).
	Save(data []byte) (filename string, err error)
	// Delete removes an image file by filename.
	Delete(filename string) error
}
