package application

import (
	"context"
	"forum/internal/modules/comment/domain"
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

func TestService_GetComment(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockCommentRepository{}
	service := NewService(mockRepo)

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
	service := NewService(mockRepo)

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
	service := NewService(mockRepo)

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
	service := NewService(mockRepo)

	// Test the current implementation (returns nil, nil since it's a placeholder)
	comment, err := service.CreateComment(ctx, "post-uuid-10", 5, "Test content")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if comment != nil {
		t.Errorf("Expected nil comment (placeholder implementation), got %v", comment)
	}
}

func TestService_UpdateComment(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockCommentRepository{}
	service := NewService(mockRepo)

	// Test the current implementation (returns nil since it's a placeholder)
	err := service.UpdateComment(ctx, "comment-uuid-1", "Updated content")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}
