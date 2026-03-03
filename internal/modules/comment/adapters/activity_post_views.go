// INPUT ADAPTER - Activity Post View Helpers
package adapters

import (
	"context"
	"time"

	postDomain "forum/internal/modules/post/domain"
)

type activityPostView struct {
	PublicID       string
	Title          string
	AuthorUsername string
	Categories     []string
	UserID         int
	LikeCount      int
	DislikeCount   int
	CommentCount   int
	CreatedAt      time.Time
}

func toActivityPostView(post *postDomain.Post) *activityPostView {
	if post == nil {
		return nil
	}
	return &activityPostView{
		PublicID:       post.PublicID,
		Title:          post.Title,
		AuthorUsername: post.AuthorUsername,
		Categories:     post.Categories,
		UserID:         post.UserID,
		LikeCount:      post.LikeCount,
		DislikeCount:   post.DislikeCount,
		CommentCount:   post.CommentCount,
		CreatedAt:      post.CreatedAt,
	}
}

func (h *HTTPHandler) listCreatedPostsForActivity(ctx context.Context, userPublicID string) ([]*activityPostView, error) {
	posts, err := h.postService.ListPosts(ctx, postDomain.PostFilter{UserID: userPublicID, Limit: 50, Offset: 0})
	if err != nil {
		return nil, err
	}
	views := make([]*activityPostView, 0, len(posts))
	for _, post := range posts {
		if view := toActivityPostView(post); view != nil {
			views = append(views, view)
		}
	}
	return views, nil
}

func (h *HTTPHandler) listLikedPostsForActivity(ctx context.Context, userPublicID string) ([]*activityPostView, error) {
	posts, err := h.postService.ListPosts(ctx, postDomain.PostFilter{LikedByUserID: userPublicID, Limit: 50, Offset: 0})
	if err != nil {
		return nil, err
	}
	views := make([]*activityPostView, 0, len(posts))
	for _, post := range posts {
		if view := toActivityPostView(post); view != nil {
			views = append(views, view)
		}
	}
	return views, nil
}

func (h *HTTPHandler) listDislikedPostsForActivity(ctx context.Context, userPublicID string) ([]*activityPostView, error) {
	posts, err := h.postService.ListPosts(ctx, postDomain.PostFilter{DislikedByUserID: userPublicID, Limit: 50, Offset: 0})
	if err != nil {
		return nil, err
	}
	views := make([]*activityPostView, 0, len(posts))
	for _, post := range posts {
		if view := toActivityPostView(post); view != nil {
			views = append(views, view)
		}
	}
	return views, nil
}

func (h *HTTPHandler) listCommentedPostsForActivity(ctx context.Context, userPublicID string) ([]*activityPostView, error) {
	posts, err := h.postService.ListPosts(ctx, postDomain.PostFilter{CommenterID: userPublicID})
	if err != nil {
		return nil, err
	}
	views := make([]*activityPostView, 0, len(posts))
	for _, post := range posts {
		if view := toActivityPostView(post); view != nil {
			views = append(views, view)
		}
	}
	return views, nil
}

func (h *HTTPHandler) getPostViewForActivity(ctx context.Context, postPublicID string) (*activityPostView, error) {
	post, err := h.postService.GetPost(ctx, postPublicID)
	if err != nil {
		return nil, err
	}
	return toActivityPostView(post), nil
}
