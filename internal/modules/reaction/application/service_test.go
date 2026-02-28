package application

import (
	"context"
	"fmt"
	commentDomain "forum/internal/modules/comment/domain"
	notificationDomain "forum/internal/modules/notification/domain"
	postDomain "forum/internal/modules/post/domain"
	"forum/internal/modules/reaction/domain"
	userDomain "forum/internal/modules/user/domain"
	"testing"
	"time"
)

type MockNotificationService struct {
	called         bool
	notifType      string
	userID         int
	actorID        int
	targetPublicID string
}

func (m *MockNotificationService) CreateNotification(ctx context.Context, userID, actorID int, notifType, message string, targetPublicID string) error {
	m.called = true
	m.notifType = notifType
	m.userID = userID
	m.actorID = actorID
	m.targetPublicID = targetPublicID
	return nil
}

func (m *MockNotificationService) GetUserNotifications(ctx context.Context, userID int) ([]*notificationDomain.Notification, error) {
	return nil, nil
}

func (m *MockNotificationService) MarkAsRead(ctx context.Context, notificationPublicID string) error {
	return nil
}

func (m *MockNotificationService) MarkAllAsRead(ctx context.Context, userID int) error {
	return nil
}

// MockReactionRepository implements ReactionRepository for testing
type MockReactionRepository struct {
	reactions       map[string]*domain.Reaction // Key: userID:targetPublicID:targetType
	countFn         func(ctx context.Context, targetPublicID string, targetType string, reactionType domain.ReactionType) (int, error)
	getByTargetFn   func(ctx context.Context, targetPublicID string, targetType string) ([]*domain.Reaction, error)
	deleteFn        func(ctx context.Context, userID int, targetPublicID string, targetType string) error
	countByUserIDFn func(ctx context.Context, userID int) (int, error)
}

func (m *MockReactionRepository) CountByTargetPublicID(ctx context.Context, targetPublicID string, targetType string, reactionType domain.ReactionType) (int, error) {
	if m.countFn != nil {
		return m.countFn(ctx, targetPublicID, targetType, reactionType)
	}

	count := 0
	for _, reaction := range m.reactions {
		if reaction.PublicTargetID == targetPublicID && reaction.TargetType == targetType && reaction.Type == reactionType {
			count++
		}
	}
	return count, nil
}

func (m *MockReactionRepository) GetByTargetPublicID(ctx context.Context, targetPublicID string, targetType string) ([]*domain.Reaction, error) {
	if m.getByTargetFn != nil {
		return m.getByTargetFn(ctx, targetPublicID, targetType)
	}

	var result []*domain.Reaction
	for _, reaction := range m.reactions {
		if reaction.PublicTargetID == targetPublicID && reaction.TargetType == targetType {
			result = append(result, reaction)
		}
	}
	return result, nil
}

func (m *MockReactionRepository) DeleteByTargetPublicID(ctx context.Context, userID int, targetPublicID string, targetType string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, userID, targetPublicID, targetType)
	}
	if m.reactions == nil {
		return nil
	}
	key := fmt.Sprintf("%d:%s:%s", userID, targetPublicID, targetType)
	delete(m.reactions, key)
	return nil
}

func (m *MockReactionRepository) GetByUserAndTargetPublicID(ctx context.Context, userID int, targetPublicID string, targetType string) (*domain.Reaction, error) {
	if m.reactions == nil {
		return nil, nil
	}
	key := fmt.Sprintf("%d:%s:%s", userID, targetPublicID, targetType)
	if r, ok := m.reactions[key]; ok {
		return r, nil
	}
	return nil, nil
}

func (m *MockReactionRepository) Create(ctx context.Context, reaction *domain.Reaction) error {
	if m.reactions == nil {
		m.reactions = make(map[string]*domain.Reaction)
	}
	key := fmt.Sprintf("%d:%s:%s", reaction.UserID, reaction.PublicTargetID, reaction.TargetType)
	m.reactions[key] = reaction
	return nil
}

func (m *MockReactionRepository) CountByUserID(ctx context.Context, userID int) (int, error) {
	if m.countByUserIDFn != nil {
		return m.countByUserIDFn(ctx, userID)
	}

	// Count reactions by this user
	count := 0
	for _, reaction := range m.reactions {
		if reaction.UserID == userID {
			count++
		}
	}
	return count, nil
}

// MockPostRepository implements PostRepository for testing
type MockPostRepository struct {
	posts     map[string]*postDomain.Post // Key: public_id
	getByIDFn func(ctx context.Context, postID string) (*postDomain.Post, error)
}

func (m *MockPostRepository) Create(ctx context.Context, post *postDomain.Post) error {
	if m.posts == nil {
		m.posts = make(map[string]*postDomain.Post)
	}
	m.posts[post.PublicID] = post
	return nil
}

func (m *MockPostRepository) GetByID(ctx context.Context, postID string) (*postDomain.Post, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, postID)
	}

	if m.posts == nil {
		return nil, fmt.Errorf("post not found")
	}
	if post, exists := m.posts[postID]; exists {
		return post, nil
	}
	return nil, fmt.Errorf("post not found")
}

func (m *MockPostRepository) Update(ctx context.Context, post *postDomain.Post) error {
	if m.posts == nil {
		m.posts = make(map[string]*postDomain.Post)
	}
	m.posts[post.PublicID] = post
	return nil
}

func (m *MockPostRepository) Delete(ctx context.Context, postID string) error {
	if m.posts == nil {
		return nil
	}
	delete(m.posts, postID)
	return nil
}

func (m *MockPostRepository) List(ctx context.Context, filter postDomain.PostFilter) ([]*postDomain.Post, error) {
	var posts []*postDomain.Post
	for _, post := range m.posts {
		posts = append(posts, post)
	}
	return posts, nil
}

func (m *MockPostRepository) UpdateImagePath(ctx context.Context, postID string, imagePath string) error {
	return nil
}

func (m *MockPostRepository) GetImagePath(ctx context.Context, postID string) (string, error) {
	return "", nil
}

// MockCommentRepository implements CommentRepository for testing
type MockCommentRepository struct {
	comments        map[string]*commentDomain.Comment // Key: public_id
	getByPublicIDFn func(ctx context.Context, commentPublicID string) (*commentDomain.Comment, error)
}

func (m *MockCommentRepository) Create(ctx context.Context, comment *commentDomain.Comment) error {
	if m.comments == nil {
		m.comments = make(map[string]*commentDomain.Comment)
	}
	m.comments[comment.PublicID] = comment
	return nil
}

func (m *MockCommentRepository) GetByPublicID(ctx context.Context, commentPublicID string) (*commentDomain.Comment, error) {
	if m.getByPublicIDFn != nil {
		return m.getByPublicIDFn(ctx, commentPublicID)
	}

	if m.comments == nil {
		return nil, fmt.Errorf("comment not found")
	}
	if comment, exists := m.comments[commentPublicID]; exists {
		return comment, nil
	}
	return nil, fmt.Errorf("comment not found")
}

func (m *MockCommentRepository) Update(ctx context.Context, comment *commentDomain.Comment) error {
	if m.comments == nil {
		m.comments = make(map[string]*commentDomain.Comment)
	}
	m.comments[comment.PublicID] = comment
	return nil
}

func (m *MockCommentRepository) DeleteByPublicID(ctx context.Context, commentPublicID string) error {
	if m.comments == nil {
		return nil
	}
	delete(m.comments, commentPublicID)
	return nil
}

func (m *MockCommentRepository) ListByPostPublicID(ctx context.Context, postPublicID string) ([]*commentDomain.Comment, error) {
	return nil, nil
}

func (m *MockCommentRepository) ListByUser(ctx context.Context, userID int) ([]*commentDomain.Comment, error) {
	return nil, nil
}

func (m *MockCommentRepository) ListByUserPaginated(ctx context.Context, userID int, limit, offset int) ([]*commentDomain.Comment, error) {
	return nil, nil
}

// MockUserService implements UserService for testing
type MockUserService struct{}

func (m *MockUserService) CreateUser(ctx context.Context, email, username, passwordHash string) (userID int, err error) {
	return 0, nil
}

func (m *MockUserService) GetByID(ctx context.Context, userID int) (*userDomain.User, error) {
	return &userDomain.User{ID: userID, PublicID: fmt.Sprintf("public-%d", userID)}, nil
}

func (m *MockUserService) GetByPublicID(ctx context.Context, publicID string) (*userDomain.User, error) {
	return &userDomain.User{PublicID: publicID, ID: 123}, nil
}

func (m *MockUserService) GetByUsername(ctx context.Context, username string) (*userDomain.User, error) {
	return nil, nil
}

func (m *MockUserService) GetByEmail(ctx context.Context, email string) (*userDomain.User, error) {
	return nil, nil
}

func (m *MockUserService) UpdateRole(ctx context.Context, userID int, newRole userDomain.Role) error {
	return nil
}

func (m *MockUserService) DeactivateUser(ctx context.Context, userID int) error {
	return nil
}

func (m *MockUserService) ActivateUser(ctx context.Context, userID int) error {
	return nil
}

func (m *MockUserService) ListUsers(ctx context.Context, offset, limit int) ([]*userDomain.User, error) {
	return nil, nil
}

func (m *MockUserService) IncrementPostCount(ctx context.Context, userID int) error {
	return nil
}

func (m *MockUserService) DecrementPostCount(ctx context.Context, userID int) error {
	return nil
}

func (m *MockUserService) IncrementCommentCount(ctx context.Context, userID int) error {
	return nil
}

func (m *MockUserService) DecrementCommentCount(ctx context.Context, userID int) error {
	return nil
}

func (m *MockUserService) IncrementReactionCount(ctx context.Context, userID int) error {
	return nil
}

func (m *MockUserService) DecrementReactionCount(ctx context.Context, userID int) error {
	return nil
}

func (m *MockUserService) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return false, nil
}

func (m *MockUserService) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return false, nil
}

func (m *MockUserService) UpdateSettings(ctx context.Context, publicID, username, email, newPassword, avatarPath string) (*userDomain.User, error) {
	return nil, nil
}

func TestService_React(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockReactionRepository{}
	mockPostRepo := &MockPostRepository{}
	mockCommentRepo := &MockCommentRepository{}
	mockUserService := &MockUserService{}
	service := NewService(mockRepo, mockPostRepo, mockCommentRepo, mockUserService)

	// Test with a mock post that exists
	mockPostRepo.posts = map[string]*postDomain.Post{
		"public-10": {ID: 10, PublicID: "public-10", UserID: 1},
	}

	// Test the implementation
	err := service.React(ctx, 1, "public-10", "post", domain.ReactionLike)
	if err != nil {
		// Since there's no real error handling in the mock, any error means an issue
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestService_RemoveReaction(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockReactionRepository{}
	mockPostRepo := &MockPostRepository{}
	mockCommentRepo := &MockCommentRepository{}
	mockUserService := &MockUserService{}
	service := NewService(mockRepo, mockPostRepo, mockCommentRepo, mockUserService)

	// Test with a mock post that exists
	mockPostRepo.posts = map[string]*postDomain.Post{
		"public-10": {ID: 10, PublicID: "public-10", UserID: 1},
	}

	// Test the implementation
	err := service.RemoveReaction(ctx, 1, "public-10", "post")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestService_GetReactions(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockReactionRepository{}
	mockPostRepo := &MockPostRepository{}
	mockCommentRepo := &MockCommentRepository{}
	mockUserService := &MockUserService{}
	service := NewService(mockRepo, mockPostRepo, mockCommentRepo, mockUserService)

	// Set up mock to return a post when GetByID is called
	mockPostRepo.getByIDFn = func(ctx context.Context, postID string) (*postDomain.Post, error) {
		if postID == "public-10" {
			return &postDomain.Post{ID: 10, PublicID: "public-10", UserID: 1}, nil
		}
		return nil, fmt.Errorf("post not found")
	}

	// Add test reactions to the mock
	now := time.Now()
	reactions := []*domain.Reaction{
		{ID: 1, UserID: 1, TargetID: 10, PublicTargetID: "public-10", TargetType: "post", Type: domain.ReactionLike, CreatedAt: now},
		{ID: 2, UserID: 2, TargetID: 10, PublicTargetID: "public-10", TargetType: "post", Type: domain.ReactionDislike, CreatedAt: now},
		{ID: 3, UserID: 3, TargetID: 15, PublicTargetID: "public-15", TargetType: "comment", Type: domain.ReactionLike, CreatedAt: now}, // Different target
	}

	// Create a map since the mock repo uses it
	if mockRepo.reactions == nil {
		mockRepo.reactions = make(map[string]*domain.Reaction)
	}

	for _, reaction := range reactions {
		key := fmt.Sprintf("%d:%s:%s", reaction.UserID, reaction.PublicTargetID, reaction.TargetType)
		mockRepo.reactions[key] = reaction
	}

	t.Run("get reactions for target", func(t *testing.T) {
		result, err := service.GetReactions(ctx, "public-10", "post")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(result) != 2 {
			t.Errorf("Expected 2 reactions for post 10, got %d", len(result))
		}

		// Verify all returned reactions belong to the correct target
		for _, reaction := range result {
			if reaction.PublicTargetID != "public-10" || reaction.TargetType != "post" {
				t.Errorf("Expected PublicTargetID 'public-10' and TargetType 'post', got %s and %s", reaction.PublicTargetID, reaction.TargetType)
			}
		}
	})

	t.Run("get reactions for target with no reactions", func(t *testing.T) {
		// For non-existing target, we expect an error because the post validation will fail
		_, err := service.GetReactions(ctx, "public-999", "post")
		if err == nil {
			t.Errorf("Expected error for non-existent post, got nil")
		}
	})
}

func TestService_CountReactions(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockReactionRepository{}
	mockPostRepo := &MockPostRepository{}
	mockCommentRepo := &MockCommentRepository{}
	mockUserService := &MockUserService{}
	service := NewService(mockRepo, mockPostRepo, mockCommentRepo, mockUserService)

	// Set up mock to return a post when GetByID is called
	mockPostRepo.getByIDFn = func(ctx context.Context, postID string) (*postDomain.Post, error) {
		if postID == "public-10" {
			return &postDomain.Post{ID: 10, PublicID: "public-10", UserID: 1}, nil
		}
		return nil, fmt.Errorf("post not found")
	}

	// Mock the expected count behavior
	mockRepo.countFn = func(ctx context.Context, targetPublicID string, targetType string, reactionType domain.ReactionType) (int, error) {
		if targetPublicID == "public-10" && targetType == "post" {
			switch reactionType {
			case domain.ReactionLike:
				return 2, nil // 2 likes
			case domain.ReactionDislike:
				return 1, nil // 1 dislike
			}
		}
		return 0, nil
	}

	likes, dislikes, err := service.CountReactions(ctx, "public-10", "post")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if likes != 2 {
		t.Errorf("Expected 2 likes, got %d", likes)
	}
	if dislikes != 1 {
		t.Errorf("Expected 1 dislike, got %d", dislikes)
	}
}

func TestService_React_SendsLikeNotification(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockReactionRepository{}
	mockPostRepo := &MockPostRepository{posts: map[string]*postDomain.Post{
		"public-10": {ID: 10, PublicID: "public-10", UserID: 1},
	}}
	mockCommentRepo := &MockCommentRepository{}
	mockUserService := &MockUserService{}
	notificationService := &MockNotificationService{}

	service := NewService(mockRepo, mockPostRepo, mockCommentRepo, mockUserService)
	service.SetNotificationService(notificationService)

	if err := service.React(ctx, 2, "public-10", "post", domain.ReactionLike); err != nil {
		t.Fatalf("React returned error: %v", err)
	}

	if !notificationService.called {
		t.Fatal("expected notification to be sent")
	}
	if notificationService.notifType != notificationDomain.TypeLike {
		t.Fatalf("expected like notification, got %s", notificationService.notifType)
	}
}

func TestService_React_SendsDislikeNotification(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockReactionRepository{}
	mockPostRepo := &MockPostRepository{posts: map[string]*postDomain.Post{
		"public-10": {ID: 10, PublicID: "public-10", UserID: 1},
	}}
	mockCommentRepo := &MockCommentRepository{}
	mockUserService := &MockUserService{}
	notificationService := &MockNotificationService{}

	service := NewService(mockRepo, mockPostRepo, mockCommentRepo, mockUserService)
	service.SetNotificationService(notificationService)

	if err := service.React(ctx, 2, "public-10", "post", domain.ReactionDislike); err != nil {
		t.Fatalf("React returned error: %v", err)
	}

	if !notificationService.called {
		t.Fatal("expected notification to be sent")
	}
	if notificationService.notifType != notificationDomain.TypeDislike {
		t.Fatalf("expected dislike notification, got %s", notificationService.notifType)
	}
}
