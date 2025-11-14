// INPUT PORT - Service Interface
// Package ports defines the input ports for the post module.
package ports

import (
	"context"
	"forum/internal/modules/post/domain"
)

// PostService defines post management use cases.
type PostService interface {
	CreatePost(ctx context.Context, userID string, title, content string, categories []string, image []byte) (*domain.Post, error)
	GetPost(ctx context.Context, postID string) (*domain.Post, error)
	UpdatePost(ctx context.Context, postID string, title, content string) error
	DeletePost(ctx context.Context, postID string) error
	ListPosts(ctx context.Context, filter PostFilter) ([]*domain.Post, error)
}

// CategoryService defines category management use cases.
type CategoryService interface {
	Create(ctx context.Context, name, description string) (*domain.Category, error)
	Get(ctx context.Context, categoryID string) (*domain.Category, error)
	List(ctx context.Context) ([]*domain.Category, error)
	Delete(ctx context.Context, categoryID string) error
}

// PostFilter represents post filtering options.
type PostFilter struct {
	UserID        string
	Categories    []string
	LikedByUserID string
	Offset        int
	Limit         int
}
