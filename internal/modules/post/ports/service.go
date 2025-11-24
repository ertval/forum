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
	DateFilter    string // "today", "week", "month", "all" (default)
	Offset        int
	Limit         int
}

// FilterService defines post filtering use cases.
type FilterService interface {
	// BuildFilter creates a PostFilter from query parameters and context
	BuildFilter(ctx context.Context, params FilterParams) PostFilter
	// ApplyDateFilter applies date constraints to a filter
	ApplyDateFilter(filter *PostFilter, dateFilter string)
}

// FilterParams represents query parameters for filtering.
type FilterParams struct {
	Category      string
	UserID        string
	MyPosts       bool
	LikedPosts    bool
	DateFilter    string
	Limit         int
	Offset        int
	CurrentUserID string
}
