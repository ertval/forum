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
    GetPost(ctx context.Context, postID int) (*domain.Post, error)
    UpdatePost(ctx context.Context, postID int, title, content string) error
    DeletePost(ctx context.Context, postID int) error
    ListPosts(ctx context.Context, filter PostFilter) ([]*domain.Post, error)
}

// PostFilter represents post filtering options.
type PostFilter struct {
    UserID     int
    Categories []string
    LikedByUserID int
    Offset     int
    Limit      int
}
