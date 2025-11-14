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
}

// CategoryRepository defines data access for categories.
type CategoryRepository interface {
	Create(ctx context.Context, category *domain.Category) error
	GetByID(ctx context.Context, categoryID string) (*domain.Category, error)
	GetByName(ctx context.Context, name string) (*domain.Category, error)
	List(ctx context.Context) ([]*domain.Category, error)
	Delete(ctx context.Context, categoryID string) error
}
