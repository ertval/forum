package adapters

import (
	"context"
	"testing"
	"time"

	commentDomain "forum/internal/modules/comment/domain"
	postDomain "forum/internal/modules/post/domain"
	reactionDomain "forum/internal/modules/reaction/domain"
	userDomain "forum/internal/modules/user/domain"
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
	comments   []*commentDomain.Comment
	byPublicID map[string]*commentDomain.Comment
}

func (m *activityMockCommentService) CreateComment(ctx context.Context, postPublicID string, userID int, content string) (*commentDomain.Comment, error) {
	return nil, nil
}
func (m *activityMockCommentService) GetComment(ctx context.Context, commentPublicID string) (*commentDomain.Comment, error) {
	if m.byPublicID != nil {
		if c, ok := m.byPublicID[commentPublicID]; ok {
			return c, nil
		}
	}
	return nil, commentDomain.ErrCommentNotFound
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

type activityMockReactionService struct {
	reactions []*reactionDomain.Reaction
}

func (m *activityMockReactionService) React(ctx context.Context, userID int, targetPublicID string, targetType string, reactionType reactionDomain.ReactionType) error {
	return nil
}

func (m *activityMockReactionService) RemoveReaction(ctx context.Context, userID int, targetPublicID string, targetType string) error {
	return nil
}

func (m *activityMockReactionService) GetReactions(ctx context.Context, targetPublicID string, targetType string) ([]*reactionDomain.Reaction, error) {
	return nil, nil
}

func (m *activityMockReactionService) CountReactions(ctx context.Context, targetPublicID string, targetType string) (likes, dislikes int, err error) {
	return 0, 0, nil
}

func (m *activityMockReactionService) GetUserReactionCount(ctx context.Context, userID int) (int, error) {
	return len(m.reactions), nil
}

func (m *activityMockReactionService) ListUserReactions(ctx context.Context, userID int) ([]*reactionDomain.Reaction, error) {
	return m.reactions, nil
}

func (m *activityMockReactionService) GetByUserAndTargetPublicID(ctx context.Context, userID int, targetPublicID string, targetType string) (*reactionDomain.Reaction, error) {
	return nil, nil
}

func (m *activityMockReactionService) CountReactionsBatch(ctx context.Context, targetPublicIDs []string, targetType string) (map[string]map[string]int, error) {
	return make(map[string]map[string]int), nil
}

type activityMockUserService struct{}

func (m *activityMockUserService) CreateUser(ctx context.Context, email, username, passwordHash string) (userID int, err error) {
	return 0, nil
}

func (m *activityMockUserService) GetByID(ctx context.Context, userID int) (*userDomain.User, error) {
	return &userDomain.User{ID: userID, PublicID: "user-123"}, nil
}

func (m *activityMockUserService) GetByPublicID(ctx context.Context, publicID string) (*userDomain.User, error) {
	return &userDomain.User{ID: 123, PublicID: publicID}, nil
}

func (m *activityMockUserService) GetByUsername(ctx context.Context, username string) (*userDomain.User, error) {
	return nil, nil
}

func (m *activityMockUserService) GetByEmail(ctx context.Context, email string) (*userDomain.User, error) {
	return nil, nil
}

func (m *activityMockUserService) UpdateRole(ctx context.Context, userID int, newRole userDomain.Role) error {
	return nil
}

func (m *activityMockUserService) DeactivateUser(ctx context.Context, userID int) error {
	return nil
}

func (m *activityMockUserService) ActivateUser(ctx context.Context, userID int) error {
	return nil
}

func (m *activityMockUserService) ListUsers(ctx context.Context, offset, limit int) ([]*userDomain.User, error) {
	return nil, nil
}

func (m *activityMockUserService) IncrementPostCount(ctx context.Context, userID int) error {
	return nil
}

func (m *activityMockUserService) DecrementPostCount(ctx context.Context, userID int) error {
	return nil
}

func (m *activityMockUserService) IncrementCommentCount(ctx context.Context, userID int) error {
	return nil
}

func (m *activityMockUserService) DecrementCommentCount(ctx context.Context, userID int) error {
	return nil
}

func (m *activityMockUserService) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return false, nil
}

func (m *activityMockUserService) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return false, nil
}

func (m *activityMockUserService) IncrementReactionCount(ctx context.Context, userID int) error {
	return nil
}

func (m *activityMockUserService) DecrementReactionCount(ctx context.Context, userID int) error {
	return nil
}

func (m *activityMockUserService) UpdateSettings(ctx context.Context, publicID, username, email, newPassword, avatarPath string) (*userDomain.User, error) {
	return nil, nil
}

func TestAggregateUserActivity_IncludesCreatedLikedDislikedAndComments(t *testing.T) {
	now := time.Now()
	h := &HTTPHandler{
		postService: &activityMockPostService{
			created:  []*postDomain.Post{{PublicID: "post-created", Title: "Created", Categories: []string{"Go"}, CreatedAt: now}},
			liked:    []*postDomain.Post{},
			disliked: []*postDomain.Post{},
			byID: map[string]*postDomain.Post{
				"post-liked":     {PublicID: "post-liked", Title: "Liked Post", Categories: []string{"Go"}, CreatedAt: now},
				"post-commented": {PublicID: "post-commented", Title: "Commented Post", Categories: []string{"Go"}, CreatedAt: now},
			},
		},
		commentService: &activityMockCommentService{
			comments: []*commentDomain.Comment{{
				PublicID:     "comment-1",
				PublicPostID: "post-commented",
				Content:      "hello",
				CreatedAt:    now,
			}},
			byPublicID: map[string]*commentDomain.Comment{
				"comment-1": {PublicID: "comment-1", PublicPostID: "post-commented", Content: "hello", CreatedAt: now},
			},
		},
		reactionService: &activityMockReactionService{reactions: []*reactionDomain.Reaction{
			{UserID: 123, TargetType: "post", PublicTargetID: "post-liked", Type: reactionDomain.ReactionLike, CreatedAt: now},
			{UserID: 123, TargetType: "comment", PublicTargetID: "comment-1", Type: reactionDomain.ReactionDislike, CreatedAt: now},
		}},
		userService: &activityMockUserService{},
	}

	activity, err := h.aggregateUserActivity(context.Background(), "user-123", activityFilters{ActivityType: "all", Time: "all", ReactionType: "all"})
	if err != nil {
		t.Fatalf("aggregateUserActivity returned error: %v", err)
	}

	createdPosts, ok := activity["created_posts"].([]map[string]interface{})
	if !ok || len(createdPosts) != 1 {
		t.Fatalf("expected one created post, got %#v", activity["created_posts"])
	}

	reactions, ok := activity["reactions"].([]map[string]interface{})
	if !ok || len(reactions) != 2 {
		t.Fatalf("expected two reactions (post + comment), got %#v", activity["reactions"])
	}

	hasPostLike := false
	hasCommentDislike := false
	for _, reaction := range reactions {
		typeValue, _ := reaction["ReactionType"].(string)
		targetType, _ := reaction["ReactionTargetType"].(string)
		if typeValue == "like" && targetType == "post" {
			hasPostLike = true
		}
		if typeValue == "dislike" && targetType == "comment" {
			hasCommentDislike = true
		}
	}
	if !hasPostLike || !hasCommentDislike {
		t.Fatalf("expected post like and comment dislike reactions, got %#v", reactions)
	}

	comments, ok := activity["comments"].([]map[string]interface{})
	if !ok || len(comments) != 1 {
		t.Fatalf("expected one comment item, got %#v", activity["comments"])
	}
	if comments[0]["PostTitle"] != "Commented Post" {
		t.Fatalf("expected comment post context title 'Commented Post', got %#v", comments[0]["PostTitle"])
	}
	if comments[0]["PostCategories"] == nil {
		t.Fatalf("expected comment post categories to be included")
	}
}

func TestAggregateUserActivity_AppliesCategoryTimeAndReactionTypeFilters(t *testing.T) {
	now := time.Now()
	old := now.AddDate(0, -2, 0)

	h := &HTTPHandler{
		postService: &activityMockPostService{
			created: []*postDomain.Post{
				{PublicID: "post-created-go", Title: "Created Go", Categories: []string{"Go"}, CreatedAt: now},
				{PublicID: "post-created-old", Title: "Created Old", Categories: []string{"Go"}, CreatedAt: old},
			},
			liked:    []*postDomain.Post{},
			disliked: []*postDomain.Post{},
			byID: map[string]*postDomain.Post{
				"post-liked-go":          {PublicID: "post-liked-go", Title: "Liked Go", Categories: []string{"Go"}, CreatedAt: now},
				"post-commented-go":      {PublicID: "post-commented-go", Title: "Commented Go", Categories: []string{"Go"}, CreatedAt: now},
				"post-commented-general": {PublicID: "post-commented-general", Title: "Commented General", Categories: []string{"General"}, CreatedAt: now},
			},
		},
		commentService: &activityMockCommentService{comments: []*commentDomain.Comment{
			{PublicID: "comment-go", PublicPostID: "post-commented-go", Content: "hello go", CreatedAt: now},
			{PublicID: "comment-general", PublicPostID: "post-commented-general", Content: "hello general", CreatedAt: now},
		}, byPublicID: map[string]*commentDomain.Comment{
			"comment-go":      {PublicID: "comment-go", PublicPostID: "post-commented-go", Content: "hello go", CreatedAt: now},
			"comment-general": {PublicID: "comment-general", PublicPostID: "post-commented-general", Content: "hello general", CreatedAt: now},
		}},
		reactionService: &activityMockReactionService{reactions: []*reactionDomain.Reaction{
			{UserID: 123, TargetType: "post", PublicTargetID: "post-liked-go", Type: reactionDomain.ReactionLike, CreatedAt: now},
			{UserID: 123, TargetType: "comment", PublicTargetID: "comment-general", Type: reactionDomain.ReactionDislike, CreatedAt: now},
		}},
		userService: &activityMockUserService{},
	}

	filters := activityFilters{
		ActivityType: "all",
		Category:     "Go",
		Time:         "month",
		ReactionType: "like",
	}

	activity, err := h.aggregateUserActivity(context.Background(), "user-123", filters)
	if err != nil {
		t.Fatalf("aggregateUserActivity returned error: %v", err)
	}

	createdPosts, ok := activity["created_posts"].([]map[string]interface{})
	if !ok || len(createdPosts) != 1 {
		t.Fatalf("expected one filtered created post, got %#v", activity["created_posts"])
	}
	if createdPosts[0]["PublicID"] != "post-created-go" {
		t.Fatalf("expected post-created-go after filtering, got %#v", createdPosts[0]["PublicID"])
	}

	reactions, ok := activity["reactions"].([]map[string]interface{})
	if !ok || len(reactions) != 1 {
		t.Fatalf("expected one filtered reaction, got %#v", activity["reactions"])
	}
	if reactions[0]["ReactionType"] != "like" {
		t.Fatalf("expected only like reactions, got %#v", reactions[0]["ReactionType"])
	}

	comments, ok := activity["comments"].([]map[string]interface{})
	if !ok || len(comments) != 1 {
		t.Fatalf("expected one filtered comment, got %#v", activity["comments"])
	}
	if comments[0]["CommentPublicID"] != "comment-go" {
		t.Fatalf("expected only Go comment, got %#v", comments[0]["CommentPublicID"])
	}
}

func TestSplitReactionItemsByTarget(t *testing.T) {
	items := []map[string]interface{}{
		{"ReactionTargetType": "post", "ReactionType": "like", "PostPublicID": "post-1"},
		{"ReactionTargetType": "comment", "ReactionType": "dislike", "PostPublicID": "post-2", "CommentPublicID": "comment-2"},
		{"ReactionTargetType": "post", "ReactionType": "dislike", "PostPublicID": "post-3"},
	}

	postReactions, commentReactions := splitReactionItemsByTarget(items)

	if len(postReactions) != 2 {
		t.Fatalf("expected 2 post reactions, got %d", len(postReactions))
	}
	if len(commentReactions) != 1 {
		t.Fatalf("expected 1 comment reaction, got %d", len(commentReactions))
	}

	if postReactions[0]["PostPublicID"] != "post-1" || postReactions[1]["PostPublicID"] != "post-3" {
		t.Fatalf("unexpected post reaction ordering/content: %#v", postReactions)
	}
	if commentReactions[0]["CommentPublicID"] != "comment-2" {
		t.Fatalf("unexpected comment reaction content: %#v", commentReactions)
	}
}
