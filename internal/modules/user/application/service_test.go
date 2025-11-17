package application

import (
	"context"
	"forum/internal/modules/user/domain"
	"forum/internal/modules/user/ports"
	"testing"
	"time"
)

// MockUserRepository implements UserRepository for testing
type MockUserRepository struct {
	users              map[int]*domain.User
	createFn           func(ctx context.Context, user *domain.User) error
	getByIDFn          func(ctx context.Context, id int) (*domain.User, error)
	getByPublicIDFn    func(ctx context.Context, publicID string) (*domain.User, error)
	getByEmailFn       func(ctx context.Context, email string) (*domain.User, error)
	getByUsernameFn    func(ctx context.Context, username string) (*domain.User, error)
	updateFn           func(ctx context.Context, user *domain.User) error
	deleteFn           func(ctx context.Context, id int) error
	updatePasswordFn   func(ctx context.Context, userID int, newPasswordHash string) error
	listFn             func(ctx context.Context, offset, limit int) ([]*domain.User, error)
	existsByEmailFn    func(ctx context.Context, email string) (bool, error)
	existsByUsernameFn func(ctx context.Context, username string) (bool, error)
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	if m.createFn != nil {
		return m.createFn(ctx, user)
	}

	if m.users == nil {
		m.users = make(map[int]*domain.User)
	}
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int) (*domain.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}

	if user, exists := m.users[id]; exists {
		return user, nil
	}
	return nil, nil
}

func (m *MockUserRepository) GetByPublicID(ctx context.Context, publicID string) (*domain.User, error) {
	if m.getByPublicIDFn != nil {
		return m.getByPublicIDFn(ctx, publicID)
	}

	for _, user := range m.users {
		if user.PublicID == publicID {
			return user, nil
		}
	}
	return nil, nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.getByEmailFn != nil {
		return m.getByEmailFn(ctx, email)
	}

	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, nil
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	if m.getByUsernameFn != nil {
		return m.getByUsernameFn(ctx, username)
	}

	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, user)
	}

	if m.users == nil {
		m.users = make(map[int]*domain.User)
	}
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id int) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}

	delete(m.users, id)
	return nil
}

func (m *MockUserRepository) UpdatePassword(ctx context.Context, userID int, newPasswordHash string) error {
	if m.updatePasswordFn != nil {
		return m.updatePasswordFn(ctx, userID, newPasswordHash)
	}

	if m.users != nil {
		if user, exists := m.users[userID]; exists {
			user.PasswordHash = newPasswordHash
		}
	}
	return nil
}

func (m *MockUserRepository) List(ctx context.Context, offset, limit int) ([]*domain.User, error) {
	if m.listFn != nil {
		return m.listFn(ctx, offset, limit)
	}

	var users []*domain.User
	count := 0
	for _, user := range m.users {
		if count >= offset && (limit == 0 || len(users) < limit) {
			users = append(users, user)
		}
		count++
	}
	return users, nil
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	if m.existsByEmailFn != nil {
		return m.existsByEmailFn(ctx, email)
	}

	for _, user := range m.users {
		if user.Email == email {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	if m.existsByUsernameFn != nil {
		return m.existsByUsernameFn(ctx, username)
	}

	for _, user := range m.users {
		if user.Username == username {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockUserRepository) Count(ctx context.Context) (int, error) {
	return len(m.users), nil
}

func (m *MockUserRepository) GetUserStats(ctx context.Context, userID int) (*ports.UserStats, error) {
	return &ports.UserStats{
		PostCount:    0,
		CommentCount: 0,
		LikeCount:    0,
		DislikeCount: 0,
	}, nil
}

func TestService_GetByID(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockUserRepository{}
	service := NewService(mockRepo)

	// Add a test user to the mock
	now := time.Now()
	user := &domain.User{
		ID:           1,
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashed_password",
		Role:         domain.RoleUser,
		CreatedAt:    now,
		UpdatedAt:    now,
		IsActive:     true,
	}
	mockRepo.users = map[int]*domain.User{
		1: user,
	}

	t.Run("successful get user by ID", func(t *testing.T) {
		result, err := service.GetByID(ctx, 1)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result == nil {
			t.Fatal("Expected user, got nil")
		}
		if result.ID != 1 {
			t.Errorf("Expected ID 1, got %d", result.ID)
		}
		if result.Email != "test@example.com" {
			t.Errorf("Expected Email 'test@example.com', got '%s'", result.Email)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		_, err := service.GetByID(ctx, 999)
		if err != nil {
			t.Errorf("Expected no error for non-existent user, got %v", err)
		}
	})
}

func TestService_GetByUsername(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockUserRepository{}
	service := NewService(mockRepo)

	// Add a test user to the mock
	now := time.Now()
	user := &domain.User{
		ID:           1,
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashed_password",
		Role:         domain.RoleUser,
		CreatedAt:    now,
		UpdatedAt:    now,
		IsActive:     true,
	}
	mockRepo.users = map[int]*domain.User{
		1: user,
	}

	t.Run("successful get user by username", func(t *testing.T) {
		result, err := service.GetByUsername(ctx, "testuser")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result == nil {
			t.Fatal("Expected user, got nil")
		}
		if result.Username != "testuser" {
			t.Errorf("Expected Username 'testuser', got '%s'", result.Username)
		}
		if result.Email != "test@example.com" {
			t.Errorf("Expected Email 'test@example.com', got '%s'", result.Email)
		}
	})

	t.Run("user not found by username", func(t *testing.T) {
		_, err := service.GetByUsername(ctx, "nonexistent")
		if err != nil {
			t.Errorf("Expected no error for non-existent user, got %v", err)
		}
	})
}

func TestService_GetByEmail(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockUserRepository{}
	service := NewService(mockRepo)

	// Add a test user to the mock
	now := time.Now()
	user := &domain.User{
		ID:           1,
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashed_password",
		Role:         domain.RoleUser,
		CreatedAt:    now,
		UpdatedAt:    now,
		IsActive:     true,
	}
	mockRepo.users = map[int]*domain.User{
		1: user,
	}

	t.Run("successful get user by email", func(t *testing.T) {
		result, err := service.GetByEmail(ctx, "test@example.com")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result == nil {
			t.Fatal("Expected user, got nil")
		}
		if result.Email != "test@example.com" {
			t.Errorf("Expected Email 'test@example.com', got '%s'", result.Email)
		}
		if result.Username != "testuser" {
			t.Errorf("Expected Username 'testuser', got '%s'", result.Username)
		}
	})

	t.Run("user not found by email", func(t *testing.T) {
		_, err := service.GetByEmail(ctx, "nonexistent@example.com")
		if err != nil {
			t.Errorf("Expected no error for non-existent user, got %v", err)
		}
	})
}

func TestService_ListUsers(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockUserRepository{}
	service := NewService(mockRepo)

	// Add test users to the mock
	now := time.Now()
	users := []*domain.User{
		{ID: 1, Email: "user1@example.com", Username: "user1", PasswordHash: "hash1", Role: domain.RoleUser, CreatedAt: now, UpdatedAt: now, IsActive: true},
		{ID: 2, Email: "user2@example.com", Username: "user2", PasswordHash: "hash2", Role: domain.RoleUser, CreatedAt: now, UpdatedAt: now, IsActive: true},
		{ID: 3, Email: "user3@example.com", Username: "user3", PasswordHash: "hash3", Role: domain.RoleAdmin, CreatedAt: now, UpdatedAt: now, IsActive: true},
	}
	mockRepo.users = map[int]*domain.User{}
	for _, user := range users {
		mockRepo.users[user.ID] = user
	}

	result, err := service.ListUsers(ctx, 0, 0) // Get all users
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 users, got %d", len(result))
	}

	// Verify we got the right users
	expectedEmails := map[string]bool{
		"user1@example.com": true,
		"user2@example.com": true,
		"user3@example.com": true,
	}

	for _, user := range result {
		if !expectedEmails[user.Email] {
			t.Errorf("Unexpected user email: %s", user.Email)
		}
	}
}

func TestService_UpdateRole(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockUserRepository{}
	service := NewService(mockRepo)

	// Test the current implementation (returns nil since it's a placeholder)
	err := service.UpdateRole(ctx, 1, domain.RoleAdmin)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestService_DeactivateUser(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockUserRepository{}
	service := NewService(mockRepo)

	// Test the current implementation (returns nil since it's a placeholder)
	err := service.DeactivateUser(ctx, 1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestService_ActivateUser(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockUserRepository{}
	service := NewService(mockRepo)

	// Test the current implementation (returns nil since it's a placeholder)
	err := service.ActivateUser(ctx, 1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestService_GetProfile(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockUserRepository{}
	service := NewService(mockRepo)

	// Test the current implementation (returns nil, nil since it's a placeholder)
	profile, err := service.GetProfile(ctx, 1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if profile != nil {
		t.Error("Expected nil profile (placeholder implementation), got non-nil profile")
	}
}

func TestService_GetUserStats(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockUserRepository{}
	service := NewService(mockRepo)

	// Test the current implementation
	stats, err := service.GetUserStats(ctx, 1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if stats == nil {
		t.Error("Expected stats to be returned, got nil")
	}
	// Verify stats structure (with mock repo, should return zero counts)
	if stats != nil {
		if stats.PostCount != 0 {
			t.Errorf("Expected PostCount 0, got %d", stats.PostCount)
		}
		if stats.CommentCount != 0 {
			t.Errorf("Expected CommentCount 0, got %d", stats.CommentCount)
		}
	}
}
