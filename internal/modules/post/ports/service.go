// INPUT PORT - Service Interface
// Package ports defines the input ports for the post module.
package ports

import (
	"context"
	"forum/internal/modules/post/domain"
)

// PostService defines post management use cases.
type PostService interface {
	CreatePost(ctx context.Context, userID int, title, content string, categories []string, image []byte) (*domain.Post, error)
	GetPost(ctx context.Context, postID string) (*domain.Post, error)
	UpdatePost(ctx context.Context, postID string, title, content string, categories []string) error
	DeletePost(ctx context.Context, postID string) error
	ListPosts(ctx context.Context, filter domain.PostFilter) ([]*domain.Post, error)
}

// CategoryService defines category management use cases.
type CategoryService interface {
	Create(ctx context.Context, name, description string) (*domain.Category, error)
	Get(ctx context.Context, categoryID string) (*domain.Category, error)
	List(ctx context.Context) ([]*domain.Category, error)
	Delete(ctx context.Context, categoryID string) error
}

// FilterService defines post filtering use cases.
type FilterService interface {
	// BuildFilter creates a PostFilter from query parameters and context
	BuildFilter(ctx context.Context, params domain.FilterParams) domain.PostFilter
	// ApplyDateFilter applies date constraints to a filter
	ApplyDateFilter(filter *domain.PostFilter, dateFilter string)
}
