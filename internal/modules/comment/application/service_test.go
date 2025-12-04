package application

import (
	"context"
	"forum/internal/modules/comment/domain"
	postDomain "forum/internal/modules/post/domain"
	userDomain "forum/internal/modules/user/domain"
	"testing"
	"time"
)

// MockCommentRepository implements CommentRepository for testing
type MockCommentRepository struct {
	comments             map[string]*domain.Comment
	listByPostPublicIDFn func(ctx context.Context, postPublicID string) ([]*domain.Comment, error)
	getByPublicIDFn      func(ctx context.Context, commentPublicID string) (*domain.Comment, error)
	createFn             func(ctx context.Context, comment *domain.Comment) error
	updateFn             func(ctx context.Context, comment *domain.Comment) error
	deleteByPublicIDFn   func(ctx context.Context, commentPublicID string) error
}

func (m *MockCommentRepository) ListByPostPublicID(ctx context.Context, postPublicID string) ([]*domain.Comment, error) {
	if m.listByPostPublicIDFn != nil {
		return m.listByPostPublicIDFn(ctx, postPublicID)
	}

	var result []*domain.Comment
	for _, comment := range m.comments {
		if comment.PublicPostID == postPublicID {
			result = append(result, comment)
		}
	}
	return result, nil
}

func (m *MockCommentRepository) GetByPublicID(ctx context.Context, commentPublicID string) (*domain.Comment, error) {
	if m.getByPublicIDFn != nil {
		return m.getByPublicIDFn(ctx, commentPublicID)
	}

	if comment, exists := m.comments[commentPublicID]; exists {
		return comment, nil
	}
	return nil, domain.ErrCommentNotFound
}

func (m *MockCommentRepository) Create(ctx context.Context, comment *domain.Comment) error {
	if m.createFn != nil {
		return m.createFn(ctx, comment)
	}

	if m.comments == nil {
		m.comments = make(map[string]*domain.Comment)
	}
	// Simulate generating a PublicID
	if comment.PublicID == "" {
		comment.PublicID = "comment-uuid-" + string(rune(len(m.comments)+1))
	}
	m.comments[comment.PublicID] = comment
	return nil
}

func (m *MockCommentRepository) Update(ctx context.Context, comment *domain.Comment) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, comment)
	}

	if m.comments == nil {
		m.comments = make(map[string]*domain.Comment)
	}
	m.comments[comment.PublicID] = comment
	return nil
}

func (m *MockCommentRepository) DeleteByPublicID(ctx context.Context, commentPublicID string) error {
	if m.deleteByPublicIDFn != nil {
		return m.deleteByPublicIDFn(ctx, commentPublicID)
	}

	delete(m.comments, commentPublicID)
	return nil
}

func (m *MockCommentRepository) ListByUser(ctx context.Context, userID int) ([]*domain.Comment, error) {
	var result []*domain.Comment
	for _, comment := range m.comments {
		if comment.UserID == userID {
			result = append(result, comment)
		}
	}
	return result, nil
}

// MockUserService implements a minimal UserService for testing
type MockUserService struct {
	getByPublicIDFn func(ctx context.Context, publicID string) (*userDomain.User, error)
}

func (m *MockUserService) CreateUser(ctx context.Context, email, username, passwordHash string) (userID int, err error) {
	return 1, nil
}

func (m *MockUserService) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return false, nil
}

func (m *MockUserService) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return false, nil
}

func (m *MockUserService) GetByID(ctx context.Context, userID int) (*userDomain.User, error) {
	return nil, nil
}

func (m *MockUserService) GetByPublicID(ctx context.Context, publicID string) (*userDomain.User, error) {
	if m.getByPublicIDFn != nil {
		return m.getByPublicIDFn(ctx, publicID)
	}
	return nil, nil
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

// MockPostService implements PostService for testing
type MockPostService struct{}

func (m *MockPostService) GetPost(ctx context.Context, publicID string) (*postDomain.Post, error) {
	// For testing, return a mock post
	return &postDomain.Post{
		ID:       10, // Internal Post ID
		PublicID: publicID,
		UserID:   1, // Author ID
	}, nil
}

func (m *MockPostService) CreatePost(ctx context.Context, userID int, title, content string, categories []string, imageData []byte) (*postDomain.Post, error) {
	return nil, nil
}

func (m *MockPostService) UpdatePost(ctx context.Context, publicID, title, content string, categories []string) error {
	return nil
}

func (m *MockPostService) UpdatePostImage(ctx context.Context, postID string, image []byte, removeImage bool) error {
	return nil
}

func (m *MockPostService) DeletePost(ctx context.Context, publicID string) error {
	return nil
}

func (m *MockPostService) ListPosts(ctx context.Context, filter postDomain.PostFilter) ([]*postDomain.Post, error) {
	return nil, nil
}

func TestService_GetComment(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockCommentRepository{}
	mockPostService := &MockPostService{}
	mockUserService := &MockUserService{}
	service := NewService(mockRepo, mockPostService, mockUserService)

	// Add a test comment to the mock
	testTime := time.Now()
	testComment := &domain.Comment{
		ID:           1,
		PublicID:     "comment-uuid-1",
		PostID:       10,
		PublicPostID: "post-uuid-10",
		UserID:       5,
		Content:      "Test comment",
		CreatedAt:    testTime,
		UpdatedAt:    testTime,
	}
	mockRepo.comments = map[string]*domain.Comment{
		"comment-uuid-1": testComment,
	}

	t.Run("successful get comment", func(t *testing.T) {
		comment, err := service.GetComment(ctx, "comment-uuid-1")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if comment == nil {
			t.Fatal("Expected comment, got nil")
		}
		if comment.PublicID != "comment-uuid-1" {
			t.Errorf("Expected comment PublicID comment-uuid-1, got %s", comment.PublicID)
		}
		if comment.Content != "Test comment" {
			t.Errorf("Expected content 'Test comment', got '%s'", comment.Content)
		}
	})

	t.Run("comment not found", func(t *testing.T) {
		_, err := service.GetComment(ctx, "comment-uuid-999")
		if err == nil {
			t.Error("Expected error for non-existent comment, got nil")
		}
	})
}

func TestService_DeleteComment(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockCommentRepository{}
	mockPostService := &MockPostService{}
	mockUserService := &MockUserService{}
	service := NewService(mockRepo, mockPostService, mockUserService)

	// Add a test comment to the mock
	testTime := time.Now()
	testComment := &domain.Comment{
		ID:           1,
		PublicID:     "comment-uuid-1",
		PostID:       10,
		PublicPostID: "post-uuid-10",
		UserID:       5,
		Content:      "Test comment",
		CreatedAt:    testTime,
		UpdatedAt:    testTime,
	}
	mockRepo.comments = map[string]*domain.Comment{
		"comment-uuid-1": testComment,
	}

	err := service.DeleteComment(ctx, "comment-uuid-1")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify the comment was removed by trying to get it again
	_, err = service.GetComment(ctx, "comment-uuid-1")
	if err == nil {
		t.Error("Expected error after deletion, got nil")
	}
}

func TestService_ListCommentsByPost(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockCommentRepository{}
	mockPostService := &MockPostService{}
	mockUserService := &MockUserService{}
	service := NewService(mockRepo, mockPostService, mockUserService)

	// Add test comments to the mock
	testTime := time.Now()
	comments := []*domain.Comment{
		{ID: 1, PublicID: "comment-uuid-1", PostID: 10, PublicPostID: "post-uuid-10", UserID: 5, Content: "First comment", CreatedAt: testTime, UpdatedAt: testTime},
		{ID: 2, PublicID: "comment-uuid-2", PostID: 10, PublicPostID: "post-uuid-10", UserID: 6, Content: "Second comment", CreatedAt: testTime, UpdatedAt: testTime},
		{ID: 3, PublicID: "comment-uuid-3", PostID: 11, PublicPostID: "post-uuid-11", UserID: 5, Content: "Third comment", CreatedAt: testTime, UpdatedAt: testTime}, // Different post
	}
	mockRepo.comments = map[string]*domain.Comment{}
	for _, comment := range comments {
		mockRepo.comments[comment.PublicID] = comment
	}

	t.Run("list comments for post", func(t *testing.T) {
		result, err := service.ListCommentsByPost(ctx, "post-uuid-10")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(result) != 2 {
			t.Errorf("Expected 2 comments for post 10, got %d", len(result))
		}

		// Verify all returned comments belong to the correct post
		for _, comment := range result {
			if comment.PublicPostID != "post-uuid-10" {
				t.Errorf("Expected PublicPostID post-uuid-10, got %s", comment.PublicPostID)
			}
		}
	})

	t.Run("list comments for post with no comments", func(t *testing.T) {
		result, err := service.ListCommentsByPost(ctx, "post-uuid-999")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(result) != 0 {
			t.Errorf("Expected 0 comments for non-existent post, got %d", len(result))
		}
	})
}

func TestService_CreateComment(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockCommentRepository{}
	mockPostService := &MockPostService{}
	mockUserService := &MockUserService{}
	service := NewService(mockRepo, mockPostService, mockUserService)

	t.Run("successful create comment", func(t *testing.T) {
		comment, err := service.CreateComment(ctx, "post-uuid-10", 5, "Test content")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if comment == nil {
			t.Fatal("Expected comment, got nil")
		}
		if comment.PostID != 10 { // PostID should now be resolved from the post service
			t.Errorf("Expected PostID 10, got %d", comment.PostID)
		}
		if comment.UserID != 5 {
			t.Errorf("Expected UserID 5, got %d", comment.UserID)
		}
		if comment.Content != "Test content" {
			t.Errorf("Expected Content 'Test content', got '%s'", comment.Content)
		}
		if comment.PublicPostID != "post-uuid-10" {
			t.Errorf("Expected PublicPostID 'post-uuid-10', got '%s'", comment.PublicPostID)
		}
		if comment.PublicID == "" {
			t.Error("Expected PublicID to be generated, got empty string")
		}
	})

	t.Run("create comment with empty content", func(t *testing.T) {
		comment, err := service.CreateComment(ctx, "post-uuid-10", 5, "")
		if err == nil {
			t.Error("Expected error for empty content, got nil")
		}
		if comment != nil {
			t.Errorf("Expected nil comment for empty content, got %v", comment)
		}
		if err != domain.ErrEmptyContent {
			t.Errorf("Expected ErrEmptyContent, got %v", err)
		}
	})
}

func TestService_UpdateComment(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockCommentRepository{}
	mockPostService := &MockPostService{}
	mockUserService := &MockUserService{}
	service := NewService(mockRepo, mockPostService, mockUserService)

	// Add a test comment to update
	testTime := time.Now()
	testComment := &domain.Comment{
		ID:        1,
		PublicID:  "comment-uuid-1",
		PostID:    10,
		UserID:    5,
		Content:   "Original content",
		CreatedAt: testTime,
		UpdatedAt: testTime,
	}
	mockRepo.comments = map[string]*domain.Comment{
		"comment-uuid-1": testComment,
	}

	t.Run("successful update comment", func(t *testing.T) {
		err := service.UpdateComment(ctx, "comment-uuid-1", "Updated content")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify the comment was updated
		updatedComment, err := service.GetComment(ctx, "comment-uuid-1")
		if err != nil {
			t.Fatalf("Failed to get updated comment: %v", err)
		}
		if updatedComment.Content != "Updated content" {
			t.Errorf("Expected updated content 'Updated content', got '%s'", updatedComment.Content)
		}
		if updatedComment.UpdatedAt.Equal(testTime) {
			t.Error("Expected UpdatedAt to be changed, but it wasn't")
		}
	})

	t.Run("update non-existent comment", func(t *testing.T) {
		err := service.UpdateComment(ctx, "comment-uuid-999", "Updated content")
		if err == nil {
			t.Error("Expected error for non-existent comment, got nil")
		}
	})

	t.Run("update comment with empty content", func(t *testing.T) {
		err := service.UpdateComment(ctx, "comment-uuid-1", "")
		if err == nil {
			t.Error("Expected error for empty content, got nil")
		}
	})
}

func TestService_ListCommentsByUser(t *testing.T) {
	ctx := context.Background()

	t.Run("successful list comments by user", func(t *testing.T) {
		testTime := time.Now()
		mockRepo := &MockCommentRepository{
			comments: map[string]*domain.Comment{
				"comment-1": {
					ID:        1,
					PublicID:  "comment-1",
					UserID:    42,
					Content:   "First comment",
					CreatedAt: testTime,
					UpdatedAt: testTime,
				},
				"comment-2": {
					ID:        2,
					PublicID:  "comment-2",
					UserID:    42,
					Content:   "Second comment",
					CreatedAt: testTime,
					UpdatedAt: testTime,
				},
			},
		}
		mockPostService := &MockPostService{}
		mockUserService := &MockUserService{
			getByPublicIDFn: func(ctx context.Context, publicID string) (*userDomain.User, error) {
				return &userDomain.User{
					ID:       42,
					PublicID: publicID,
					Username: "testuser",
				}, nil
			},
		}
		service := NewService(mockRepo, mockPostService, mockUserService)

		comments, err := service.ListCommentsByUser(ctx, "user-public-id")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(comments) != 2 {
			t.Errorf("Expected 2 comments, got %d", len(comments))
		}
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo := &MockCommentRepository{}
		mockPostService := &MockPostService{}
		mockUserService := &MockUserService{
			getByPublicIDFn: func(ctx context.Context, publicID string) (*userDomain.User, error) {
				return nil, userDomain.ErrUserNotFound
			},
		}
		service := NewService(mockRepo, mockPostService, mockUserService)

		_, err := service.ListCommentsByUser(ctx, "non-existent-user")
		if err == nil {
			t.Error("Expected error for non-existent user, got nil")
		}
	})
}

func TestService_CreateComment_ValidationError(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockCommentRepository{}
	mockPostService := &MockPostService{}
	mockUserService := &MockUserService{}
	service := NewService(mockRepo, mockPostService, mockUserService)

	// Test with empty content (validation error)
	_, err := service.CreateComment(ctx, "post-uuid-1", 1, "")
	if err == nil {
		t.Error("Expected validation error for empty content, got nil")
	}
}

func TestService_DeleteComment_NotFound(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockCommentRepository{
		getByPublicIDFn: func(ctx context.Context, commentPublicID string) (*domain.Comment, error) {
			return nil, domain.ErrCommentNotFound
		},
	}
	mockPostService := &MockPostService{}
	mockUserService := &MockUserService{}
	service := NewService(mockRepo, mockPostService, mockUserService)

	err := service.DeleteComment(ctx, "non-existent-comment")
	if err == nil {
		t.Error("Expected error for non-existent comment, got nil")
	}
}
