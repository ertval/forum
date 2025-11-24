package application

import (
	"context"
	"fmt"
	"forum/internal/modules/reaction/domain"
	"testing"
	"time"
)

// MockReactionRepository implements ReactionRepository for testing
type MockReactionRepository struct {
	reactions     map[string]*domain.Reaction // Key: userID:targetPublicID:targetType
	countFn       func(ctx context.Context, targetPublicID string, targetType string, reactionType domain.ReactionType) (int, error)
	getByTargetFn func(ctx context.Context, targetPublicID string, targetType string) ([]*domain.Reaction, error)
	deleteFn      func(ctx context.Context, userID int, targetPublicID string, targetType string) error
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

func TestService_React(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockReactionRepository{}
	service := NewService(mockRepo)

	// Test the current implementation (returns nil since it's a placeholder)
	err := service.React(ctx, 1, "public-10", "post", domain.ReactionLike)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestService_RemoveReaction(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockReactionRepository{}
	service := NewService(mockRepo)

	// Test the current implementation (returns nil since it's a placeholder)
	err := service.RemoveReaction(ctx, 1, "public-10", "post")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestService_GetReactions(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockReactionRepository{}
	service := NewService(mockRepo)

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
		result, err := service.GetReactions(ctx, "public-999", "post")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(result) != 0 {
			t.Errorf("Expected 0 reactions for non-existent target, got %d", len(result))
		}
	})
}

func TestService_CountReactions(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockReactionRepository{}
	service := NewService(mockRepo)

	// Test the current implementation (returns 0, 0, nil since it's a placeholder)
	likes, dislikes, err := service.CountReactions(ctx, "public-10", "post")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if likes != 0 {
		t.Errorf("Expected 0 likes, got %d", likes)
	}
	if dislikes != 0 {
		t.Errorf("Expected 0 dislikes, got %d", dislikes)
	}
}
