package output
// Package output defines the outbound ports for the post module.
package output

import (
	"context"
	"io"

	"forum/internal/modules/post/domain"
)

// PostRepository defines the interface for post persistence.
type PostRepository interface {
	Create(ctx context.Context, post *domain.Post) error
	GetByID(ctx context.Context, postID string) (*domain.Post, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Post, error)
	ListByCategory(ctx context.Context, categoryID string, limit, offset int) ([]*domain.Post, error)
	ListByAuthor(ctx context.Context, authorID string, limit, offset int) ([]*domain.Post, error)
	Update(ctx context.Context, post *domain.Post) error
	Delete(ctx context.Context, postID string) error
}

// CategoryRepository defines the interface for category persistence.
type CategoryRepository interface {
	Create(ctx context.Context, category *domain.Category) error
	GetByID(ctx context.Context, categoryID string) (*domain.Category, error)
	List(ctx context.Context) ([]*domain.Category, error)
	Delete(ctx context.Context, categoryID string) error
}

// ImageStorage defines the interface for image storage operations.
type ImageStorage interface {
	// Save saves an image and returns its path.
	Save(ctx context.Context, file io.Reader, filename string) (string, error)

	// Delete deletes an image by path.
	Delete(ctx context.Context, path string) error

	// ValidateImage validates image format and size.
	ValidateImage(file io.Reader) error
}
