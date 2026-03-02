// INPUT PORT - Service Interface
// Package ports defines the input ports (use cases) for the post module.
package ports

import (
	"context"
	"forum/internal/modules/post/domain"
)

// PostService defines post management use cases.
type PostService interface {
	// CreatePost creates a new post with the given details and optional image.
	CreatePost(ctx context.Context, userID int, title, content string, categories []string, image []byte) (*domain.Post, error)
	// GetPost retrieves a post by its public UUID.
	GetPost(ctx context.Context, postID string) (*domain.Post, error)
	// UpdatePost modifies an existing post's title, content, and categories.
	UpdatePost(ctx context.Context, postID string, title, content string, categories []string) error
	// UpdatePostImage updates or removes the image for a post.
	// If image is nil or empty and removeImage is true, the existing image is removed.
	// If image is provided, the existing image is replaced.
	UpdatePostImage(ctx context.Context, postID string, image []byte, removeImage bool) error
	// DeletePost removes a post and its associated data by public UUID.
	DeletePost(ctx context.Context, postID string) error
	// ListPosts returns posts matching the given filter criteria.
	ListPosts(ctx context.Context, filter domain.PostFilter) ([]*domain.Post, error)
	// MaxImageSize returns the maximum allowed image size in bytes.
	MaxImageSize() int64
}

// CategoryService defines category management use cases.
type CategoryService interface {
	// Create creates a new category with the given name and description.
	Create(ctx context.Context, name, description string) (*domain.Category, error)
	// Get retrieves a category by its public UUID.
	Get(ctx context.Context, categoryID string) (*domain.Category, error)
	// List returns all available categories.
	List(ctx context.Context) ([]*domain.Category, error)
	// Delete removes a category by its public UUID.
	Delete(ctx context.Context, categoryID string) error
}
