package domain

import (
	"testing"
	"time"
)

func TestComment_Validate(t *testing.T) {
	tests := []struct {
		name        string
		comment     *Comment
		expectError bool
	}{
		{
			name: "valid comment with content",
			comment: &Comment{
				ID:        1,
				PostID:    1,
				UserID:    1,
				Content:   "This is a valid comment",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectError: false,
		},
		{
			name: "comment with empty content",
			comment: &Comment{
				ID:        1,
				PostID:    1,
				UserID:    1,
				Content:   "",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectError: true, // Now Validate returns ErrEmptyContent
		},
		{
			name: "comment with only whitespace content",
			comment: &Comment{
				ID:        1,
				PostID:    1,
				UserID:    1,
				Content:   "   ",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectError: true, // Now Validate returns ErrEmptyContent
		},
		{
			name: "comment with content exceeding 5000 characters",
			comment: &Comment{
				ID:        1,
				PostID:    1,
				UserID:    1,
				Content:   string(make([]byte, 5001)), // 5001 characters
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectError: true, // ErrContentTooLong
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.comment.Validate()
			if tt.expectError && err == nil {
				t.Errorf("Expected error for comment validation, but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error for comment validation, but got %v", err)
			}
		})
	}
}

func TestComment_StructFields(t *testing.T) {
	now := time.Now()
	comment := &Comment{
		ID:        1,
		PostID:    10,
		UserID:    5,
		Content:   "Test comment content",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if comment.ID != 1 {
		t.Errorf("Expected ID 1, got %d", comment.ID)
	}
	if comment.PostID != 10 {
		t.Errorf("Expected PostID 10, got %d", comment.PostID)
	}
	if comment.UserID != 5 {
		t.Errorf("Expected UserID 5, got %d", comment.UserID)
	}
	if comment.Content != "Test comment content" {
		t.Errorf("Expected Content 'Test comment content', got '%s'", comment.Content)
	}
	if !comment.CreatedAt.Equal(now) {
		t.Errorf("Expected CreatedAt %v, got %v", now, comment.CreatedAt)
	}
	if !comment.UpdatedAt.Equal(now) {
		t.Errorf("Expected UpdatedAt %v, got %v", now, comment.UpdatedAt)
	}
}
