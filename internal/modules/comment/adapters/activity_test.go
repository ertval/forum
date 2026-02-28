package adapters

import (
	"context"
	"testing"
	"time"

	commentDomain "forum/internal/modules/comment/domain"
	postDomain "forum/internal/modules/post/domain"
)

type activityMockPostService struct {
	created  []*postDomain.Post
	liked    []*postDomain.Post
	disliked []*postDomain.Post
	byID     map[string]*postDomain.Post
}

func (m *activityMockPostService) CreatePost(ctx context.Context, userID int, title, content string, categories []string, image []byte) (*postDomain.Post, error) {
	return nil, nil
}
func (m *activityMockPostService) GetPost(ctx context.Context, postID string) (*postDomain.Post, error) {
	if m.byID != nil {
		if post, ok := m.byID[postID]; ok {
			return post, nil
		}
	}
	return nil, postDomain.ErrPostNotFound
}
func (m *activityMockPostService) UpdatePost(ctx context.Context, postID string, title, content string, categories []string) error {
	return nil
}
func (m *activityMockPostService) UpdatePostImage(ctx context.Context, postID string, image []byte, removeImage bool) error {
	return nil
}
func (m *activityMockPostService) DeletePost(ctx context.Context, postID string) error { return nil }
func (m *activityMockPostService) ListPosts(ctx context.Context, filter postDomain.PostFilter) ([]*postDomain.Post, error) {
	switch {
	case filter.UserID != "":
		return m.created, nil
	case filter.LikedByUserID != "":
		return m.liked, nil
	case filter.DislikedByUserID != "":
		return m.disliked, nil
	default:
		return []*postDomain.Post{}, nil
	}
}
func (m *activityMockPostService) MaxImageSize() int64 { return 0 }

type activityMockCommentService struct {
	comments []*commentDomain.Comment
}

func (m *activityMockCommentService) CreateComment(ctx context.Context, postPublicID string, userID int, content string) (*commentDomain.Comment, error) {
	return nil, nil
}
func (m *activityMockCommentService) GetComment(ctx context.Context, commentPublicID string) (*commentDomain.Comment, error) {
	return nil, nil
}
func (m *activityMockCommentService) UpdateComment(ctx context.Context, commentPublicID string, content string) error {
	return nil
}
func (m *activityMockCommentService) DeleteComment(ctx context.Context, commentPublicID string) error {
	return nil
}
func (m *activityMockCommentService) ListCommentsByPost(ctx context.Context, postPublicID string) ([]*commentDomain.Comment, error) {
	return nil, nil
}
func (m *activityMockCommentService) ListCommentsByUser(ctx context.Context, userPublicID string) ([]*commentDomain.Comment, error) {
	return m.comments, nil
}
func (m *activityMockCommentService) ListCommentsByUserPaginated(ctx context.Context, userPublicID string, limit, offset int) ([]*commentDomain.Comment, error) {
	return m.comments, nil
}

func TestAggregateUserActivity_IncludesCreatedLikedDislikedAndComments(t *testing.T) {
	now := time.Now()
	h := &HTTPHandler{
		postService: &activityMockPostService{
			created:  []*postDomain.Post{{PublicID: "post-created", Title: "Created", CreatedAt: now}},
			liked:    []*postDomain.Post{{PublicID: "post-liked", Title: "Liked Post", CreatedAt: now}},
			disliked: []*postDomain.Post{{PublicID: "post-disliked", Title: "Disliked Post", CreatedAt: now}},
			byID: map[string]*postDomain.Post{
				"post-commented": {PublicID: "post-commented", Title: "Commented Post", CreatedAt: now},
			},
		},
		commentService: &activityMockCommentService{comments: []*commentDomain.Comment{{
			PublicID:     "comment-1",
			PublicPostID: "post-commented",
			Content:      "hello",
			CreatedAt:    now,
		}}},
	}

	activity, err := h.aggregateUserActivity(context.Background(), "user-123")
	if err != nil {
		t.Fatalf("aggregateUserActivity returned error: %v", err)
	}

	createdPosts, ok := activity["created_posts"].([]map[string]interface{})
	if !ok || len(createdPosts) != 1 {
		t.Fatalf("expected one created post, got %#v", activity["created_posts"])
	}

	reactions, ok := activity["reactions"].([]map[string]interface{})
	if !ok || len(reactions) != 2 {
		t.Fatalf("expected two reactions (like + dislike), got %#v", activity["reactions"])
	}

	hasLike := false
	hasDislike := false
	for _, reaction := range reactions {
		typeValue, _ := reaction["ReactionType"].(string)
		if typeValue == "like" {
			hasLike = true
		}
		if typeValue == "dislike" {
			hasDislike = true
		}
	}
	if !hasLike || !hasDislike {
		t.Fatalf("expected both like and dislike reactions, got %#v", reactions)
	}

	comments, ok := activity["comments"].([]map[string]interface{})
	if !ok || len(comments) != 1 {
		t.Fatalf("expected one comment item, got %#v", activity["comments"])
	}
	if comments[0]["PostTitle"] != "Commented Post" {
		t.Fatalf("expected comment post context title 'Commented Post', got %#v", comments[0]["PostTitle"])
	}
}
