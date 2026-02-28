package ports

import (
	"context"
	"forum/internal/modules/reaction/domain"
	"testing"
)

// This test file verifies that the interfaces are properly defined and can be implemented
func TestReactionServiceInterface(t *testing.T) {
	// This test ensures that the ReactionService interface is properly defined
	// and that we can create a variable of the interface type

	var reactionService ReactionService
	_ = reactionService // ensure the interface is defined and can be used
}

func TestReactionRepositoryInterface(t *testing.T) {
	// This test ensures that the ReactionRepository interface is properly defined
	var reactionRepo ReactionRepository
	_ = reactionRepo // ensure the interface is defined and can be used
}

// Mock implementations for interface compatibility testing
type mockReactionService struct{}

func (m *mockReactionService) React(ctx context.Context, userID int, targetPublicID string, targetType string, reactionType domain.ReactionType) error {
	return nil
}

func (m *mockReactionService) RemoveReaction(ctx context.Context, userID int, targetPublicID string, targetType string) error {
	return nil
}

func (m *mockReactionService) GetReactions(ctx context.Context, targetPublicID string, targetType string) ([]*domain.Reaction, error) {
	return nil, nil
}

func (m *mockReactionService) CountReactions(ctx context.Context, targetPublicID string, targetType string) (likes, dislikes int, err error) {
	return 0, 0, nil
}

func (m *mockReactionService) GetUserReactionCount(ctx context.Context, userID int) (int, error) {
	return 0, nil
}

func (m *mockReactionService) GetByUserAndTargetPublicID(ctx context.Context, userID int, targetPublicID string, targetType string) (*domain.Reaction, error) {
	return nil, nil
}

type mockReactionRepository struct{}

func (m *mockReactionRepository) Create(ctx context.Context, reaction *domain.Reaction) error {
	return nil
}

func (m *mockReactionRepository) DeleteByTargetPublicID(ctx context.Context, userID int, targetPublicID string, targetType string) error {
	return nil
}

func (m *mockReactionRepository) GetByTargetPublicID(ctx context.Context, targetPublicID string, targetType string) ([]*domain.Reaction, error) {
	return nil, nil
}

func (m *mockReactionRepository) GetByUserAndTargetPublicID(ctx context.Context, userID int, targetPublicID string, targetType string) (*domain.Reaction, error) {
	return nil, nil
}

func (m *mockReactionRepository) CountByTargetPublicID(ctx context.Context, targetPublicID string, targetType string, reactionType domain.ReactionType) (int, error) {
	return 0, nil
}

func (m *mockReactionRepository) CountByUserID(ctx context.Context, userID int) (int, error) {
	return 0, nil
}

func (m *mockReactionRepository) ToggleReaction(ctx context.Context, reaction *domain.Reaction) (removed bool, err error) {
	return false, nil
}

// Compile-time interface satisfaction checks
var _ ReactionService = (*mockReactionService)(nil)
var _ ReactionRepository = (*mockReactionRepository)(nil)

func TestReactionServiceInterfaceMethods(t *testing.T) {
	ctx := context.Background()
	service := &mockReactionService{}

	// Test each method signature with correct parameter types
	err := service.React(ctx, 1, "test-uuid", "post", domain.ReactionLike)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = service.RemoveReaction(ctx, 1, "test-uuid", "post")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = service.GetReactions(ctx, "test-uuid", "post")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, _, err = service.CountReactions(ctx, "test-uuid", "post")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = service.GetUserReactionCount(ctx, 1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = service.GetByUserAndTargetPublicID(ctx, 1, "test-uuid", "post")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestReactionRepositoryInterfaceMethods(t *testing.T) {
	ctx := context.Background()
	repo := &mockReactionRepository{}

	// Test Create
	err := repo.Create(ctx, &domain.Reaction{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test DeleteByTargetPublicID
	err = repo.DeleteByTargetPublicID(ctx, 1, "test-uuid", "post")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test GetByTargetPublicID
	_, err = repo.GetByTargetPublicID(ctx, "test-uuid", "post")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test GetByUserAndTargetPublicID
	_, err = repo.GetByUserAndTargetPublicID(ctx, 1, "test-uuid", "post")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test CountByTargetPublicID
	_, err = repo.CountByTargetPublicID(ctx, "test-uuid", "post", domain.ReactionLike)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test CountByUserID
	_, err = repo.CountByUserID(ctx, 1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test ToggleReaction
	_, err = repo.ToggleReaction(ctx, &domain.Reaction{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
