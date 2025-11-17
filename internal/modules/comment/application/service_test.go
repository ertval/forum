package application

import (
	"context"
	"errors"
	"forum/internal/modules/comment/domain"
	"forum/internal/modules/comment/ports"
	"testing"
	"time"
)

// MockCommentRepository implements CommentRepository for testing
type MockCommentRepository struct {
	comments       map[int]*domain.Comment
	listByPostIDFn func(ctx context.Context, postID int) ([]*domain.Comment, error)
	getByIDFn      func(ctx context.Context, commentID int) (*domain.Comment, error)
	createFn       func(ctx context.Context, comment *domain.Comment) error
	updateFn       func(ctx context.Context, comment *domain.Comment) error
	deleteFn       func(ctx context.Context, commentID int) error
}

func (m *MockCommentRepository) ListByPostID(ctx context.Context, postID int) ([]*domain.Comment, error) {
	if m.listByPostIDFn != nil {
		return m.listByPostIDFn(ctx, postID)
	}

	var result []*domain.Comment
	for _, comment := range m.comments {
		if comment.PostID == postID {
			result = append(result, comment)
		}
	}
	return result, nil
}

func (m *MockCommentRepository) GetByID(ctx context.Context, commentID int) (*domain.Comment, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, commentID)
	}

	if comment, exists := m.comments[commentID]; exists {
		return comment, nil
	}
	return nil, domain.ErrCommentNotFound
}

func (m *MockCommentRepository) Create(ctx context.Context, comment *domain.Comment) error {
	if m.createFn != nil {
		return m.createFn(ctx, comment)
	}

	if m.comments == nil {
		m.comments = make(map[int]*domain.Comment)
	}
	m.comments[comment.ID] = comment
	return nil
}

func (m *MockCommentRepository) Update(ctx context.Context, comment *domain.Comment) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, comment)
	}

	if m.comments == nil {
		m.comments = make(map[int]*domain.Comment)
	}
	m.comments[comment.ID] = comment
	return nil
}

func (m *MockCommentRepository) Delete(ctx context.Context, commentID int) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, commentID)
	}

	delete(m.comments, commentID)
	return nil
}

func TestService_GetComment(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockCommentRepository{}
	service := NewService(mockRepo)

	// Add a test comment to the mock
	testTime := time.Now()
	testComment := &domain.Comment{
		ID:        1,
		PostID:    10,
		UserID:    5,
		Content:   "Test comment",
		CreatedAt: testTime,
		UpdatedAt: testTime,
	}
	mockRepo.comments = map[int]*domain.Comment{
		1: testComment,
	}

	t.Run("successful get comment", func(t *testing.T) {
		comment, err := service.GetComment(ctx, 1)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if comment == nil {
			t.Fatal("Expected comment, got nil")
		}
		if comment.ID != 1 {
			t.Errorf("Expected comment ID 1, got %d", comment.ID)
		}
		if comment.Content != "Test comment" {
			t.Errorf("Expected content 'Test comment', got '%s'", comment.Content)
		}
	})

	t.Run("comment not found", func(t *testing.T) {
		_, err := service.GetComment(ctx, 999)
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
		ID:        1,
		PostID:    10,
		UserID:    5,
		Content:   "Test comment",
		CreatedAt: testTime,
		UpdatedAt: testTime,
	}
	mockRepo.comments = map[int]*domain.Comment{
		1: testComment,
	}

	err := service.DeleteComment(ctx, 1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify the comment was removed by trying to get it again
	_, err = service.GetComment(ctx, 1)
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
		{ID: 1, PostID: 10, UserID: 5, Content: "First comment", CreatedAt: testTime, UpdatedAt: testTime},
		{ID: 2, PostID: 10, UserID: 6, Content: "Second comment", CreatedAt: testTime, UpdatedAt: testTime},
		{ID: 3, PostID: 11, UserID: 5, Content: "Third comment", CreatedAt: testTime, UpdatedAt: testTime}, // Different post
	}
	mockRepo.comments = map[int]*domain.Comment{}
	for _, comment := range comments {
		mockRepo.comments[comment.ID] = comment
	}

	t.Run("list comments for post", func(t *testing.T) {
		result, err := service.ListCommentsByPost(ctx, 10)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(result) != 2 {
			t.Errorf("Expected 2 comments for post 10, got %d", len(result))
		}

		// Verify all returned comments belong to the correct post
		for _, comment := range result {
			if comment.PostID != 10 {
				t.Errorf("Expected PostID 10, got %d", comment.PostID)
			}
		}
	})

	t.Run("list comments for post with no comments", func(t *testing.T) {
		result, err := service.ListCommentsByPost(ctx, 999)
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
	comment, err := service.CreateComment(ctx, 10, 5, "Test content")
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
	err := service.UpdateComment(ctx, 1, "Updated content")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}