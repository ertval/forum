package application

import (
	"context"
	"errors"
	"forum/internal/modules/user/domain"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
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

// IncrementPostCount increments the user's post count.
func (m *MockUserRepository) IncrementPostCount(ctx context.Context, userID int) error {
	if user, exists := m.users[userID]; exists {
		user.PostCount++
	}
	return nil
}

// DecrementPostCount decrements the user's post count.
func (m *MockUserRepository) DecrementPostCount(ctx context.Context, userID int) error {
	if user, exists := m.users[userID]; exists && user.PostCount > 0 {
		user.PostCount--
	}
	return nil
}

// IncrementCommentCount increments the user's comment count.
func (m *MockUserRepository) IncrementCommentCount(ctx context.Context, userID int) error {
	if user, exists := m.users[userID]; exists {
		user.CommentCount++
	}
	return nil
}

// DecrementCommentCount decrements the user's comment count.
func (m *MockUserRepository) DecrementCommentCount(ctx context.Context, userID int) error {
	if user, exists := m.users[userID]; exists && user.CommentCount > 0 {
		user.CommentCount--
	}
	return nil
}

// IncrementReactionCount increments the user's reaction count.
func (m *MockUserRepository) IncrementReactionCount(ctx context.Context, userID int) error {
	if user, exists := m.users[userID]; exists {
		user.ReactionCount++
	}
	return nil
}

// DecrementReactionCount decrements the user's reaction count.
func (m *MockUserRepository) DecrementReactionCount(ctx context.Context, userID int) error {
	if user, exists := m.users[userID]; exists && user.ReactionCount > 0 {
		user.ReactionCount--
	}
	return nil
}

func TestService_UpdateSettings_Success(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	initialHash, _ := bcrypt.GenerateFromPassword([]byte("OldPassword123"), bcrypt.DefaultCost)

	mockRepo := &MockUserRepository{
		users: map[int]*domain.User{
			1: {
				ID:           1,
				PublicID:     "test-public-id",
				Email:        "old@example.com",
				Username:     "Old Name",
				PasswordHash: string(initialHash),
				Role:         domain.RoleUser,
				CreatedAt:    now,
				UpdatedAt:    now,
				IsActive:     true,
			},
		},
	}
	service := NewService(mockRepo)

	updated, err := service.UpdateSettings(ctx, "test-public-id", "Alice Smith", "alice@example.com", "NewStrongPass123", "avatar-new.png")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated == nil {
		t.Fatalf("expected updated user")
	}
	if updated.Username != "Alice Smith" {
		t.Fatalf("expected updated username, got %s", updated.Username)
	}
	if updated.Email != "alice@example.com" {
		t.Fatalf("expected updated email, got %s", updated.Email)
	}
	if updated.AvatarPath != "avatar-new.png" {
		t.Fatalf("expected avatar path to be updated")
	}
	if updated.AvatarURL != "/static/uploads/avatar-new.png" {
		t.Fatalf("expected avatar URL to be derived")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(updated.PasswordHash), []byte("NewStrongPass123")); err != nil {
		t.Fatalf("expected password hash to match new password")
	}
}

func TestService_UpdateSettings_InvalidEmail(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	mockRepo := &MockUserRepository{
		users: map[int]*domain.User{
			1: {
				ID:           1,
				PublicID:     "test-public-id",
				Email:        "old@example.com",
				Username:     "Old Name",
				PasswordHash: "hash",
				Role:         domain.RoleUser,
				CreatedAt:    now,
				UpdatedAt:    now,
				IsActive:     true,
			},
		},
	}
	service := NewService(mockRepo)

	_, err := service.UpdateSettings(ctx, "test-public-id", "Alice Smith", "not-an-email", "", "")
	if !errors.Is(err, domain.ErrInvalidEmail) {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestService_GetByPublicID(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockUserRepository{}
	service := NewService(mockRepo)

	// Add a test user to the mock
	now := time.Now()
	user := &domain.User{
		ID:           1,
		PublicID:     "test-public-id",
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

	t.Run("successful get user by public ID", func(t *testing.T) {
		result, err := service.GetByPublicID(ctx, "test-public-id")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result == nil {
			t.Fatal("Expected user, got nil")
		}
		if result.PublicID != "test-public-id" {
			t.Errorf("Expected PublicID 'test-public-id', got '%s'", result.PublicID)
		}
		if result.Email != "test@example.com" {
			t.Errorf("Expected Email 'test@example.com', got '%s'", result.Email)
		}
	})

	t.Run("user not found by public ID", func(t *testing.T) {
		_, err := service.GetByPublicID(ctx, "nonexistent")
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
	now := time.Now()

	t.Run("successfully update role", func(t *testing.T) {
		mockRepo := &MockUserRepository{
			users: map[int]*domain.User{
				1: {
					ID:        1,
					Email:     "test@example.com",
					Username:  "testuser",
					Role:      domain.RoleUser,
					IsActive:  true,
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
		}
		service := NewService(mockRepo)

		err := service.UpdateRole(ctx, 1, domain.RoleModerator)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify the role was updated
		if mockRepo.users[1].Role != domain.RoleModerator {
			t.Errorf("Expected role to be moderator, got %s", mockRepo.users[1].Role)
		}
	})

	t.Run("invalid role returns error", func(t *testing.T) {
		mockRepo := &MockUserRepository{
			users: map[int]*domain.User{
				1: {ID: 1, Email: "test@example.com", Role: domain.RoleUser, IsActive: true},
			},
		}
		service := NewService(mockRepo)

		err := service.UpdateRole(ctx, 1, domain.Role("invalid"))
		if err != domain.ErrInvalidRole {
			t.Errorf("Expected ErrInvalidRole, got %v", err)
		}
	})

	t.Run("user not found returns error", func(t *testing.T) {
		mockRepo := &MockUserRepository{
			users: map[int]*domain.User{},
			getByIDFn: func(ctx context.Context, id int) (*domain.User, error) {
				return nil, domain.ErrUserNotFound
			},
		}
		service := NewService(mockRepo)

		err := service.UpdateRole(ctx, 999, domain.RoleAdmin)
		if !errors.Is(err, domain.ErrUserNotFound) {
			t.Errorf("Expected ErrUserNotFound, got %v", err)
		}
	})
}

func TestService_DeactivateUser(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	t.Run("successfully deactivate user", func(t *testing.T) {
		mockRepo := &MockUserRepository{
			users: map[int]*domain.User{
				1: {
					ID:        1,
					Email:     "test@example.com",
					IsActive:  true,
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
		}
		service := NewService(mockRepo)

		err := service.DeactivateUser(ctx, 1)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify user was deactivated
		if mockRepo.users[1].IsActive {
			t.Error("Expected user to be deactivated")
		}
	})

	t.Run("deactivate already inactive user succeeds", func(t *testing.T) {
		mockRepo := &MockUserRepository{
			users: map[int]*domain.User{
				1: {ID: 1, Email: "test@example.com", IsActive: false},
			},
		}
		service := NewService(mockRepo)

		err := service.DeactivateUser(ctx, 1)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("user not found returns error", func(t *testing.T) {
		mockRepo := &MockUserRepository{
			users: map[int]*domain.User{},
			getByIDFn: func(ctx context.Context, id int) (*domain.User, error) {
				return nil, domain.ErrUserNotFound
			},
		}
		service := NewService(mockRepo)

		err := service.DeactivateUser(ctx, 999)
		if !errors.Is(err, domain.ErrUserNotFound) {
			t.Errorf("Expected ErrUserNotFound, got %v", err)
		}
	})
}

func TestService_ActivateUser(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	t.Run("successfully activate user", func(t *testing.T) {
		mockRepo := &MockUserRepository{
			users: map[int]*domain.User{
				1: {
					ID:        1,
					Email:     "test@example.com",
					IsActive:  false,
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
		}
		service := NewService(mockRepo)

		err := service.ActivateUser(ctx, 1)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify user was activated
		if !mockRepo.users[1].IsActive {
			t.Error("Expected user to be activated")
		}
	})

	t.Run("activate already active user succeeds", func(t *testing.T) {
		mockRepo := &MockUserRepository{
			users: map[int]*domain.User{
				1: {ID: 1, Email: "test@example.com", IsActive: true},
			},
		}
		service := NewService(mockRepo)

		err := service.ActivateUser(ctx, 1)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("user not found returns error", func(t *testing.T) {
		mockRepo := &MockUserRepository{
			users: map[int]*domain.User{},
			getByIDFn: func(ctx context.Context, id int) (*domain.User, error) {
				return nil, domain.ErrUserNotFound
			},
		}
		service := NewService(mockRepo)

		err := service.ActivateUser(ctx, 999)
		if !errors.Is(err, domain.ErrUserNotFound) {
			t.Errorf("Expected ErrUserNotFound, got %v", err)
		}
	})
}

func TestService_IncrementPostCount(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockUserRepository{}
	service := NewService(mockRepo)

	// Add a test user
	now := time.Now()
	user := &domain.User{
		ID:        1,
		Email:     "test@example.com",
		Username:  "testuser",
		PostCount: 5,
		CreatedAt: now,
		UpdatedAt: now,
		IsActive:  true,
	}
	mockRepo.users = map[int]*domain.User{1: user}

	err := service.IncrementPostCount(ctx, 1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify count was incremented
	if user.PostCount != 6 {
		t.Errorf("Expected PostCount 6, got %d", user.PostCount)
	}
}

func TestService_DecrementPostCount(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockUserRepository{}
	service := NewService(mockRepo)

	// Add a test user
	now := time.Now()
	user := &domain.User{
		ID:        1,
		Email:     "test@example.com",
		Username:  "testuser",
		PostCount: 5,
		CreatedAt: now,
		UpdatedAt: now,
		IsActive:  true,
	}
	mockRepo.users = map[int]*domain.User{1: user}

	err := service.DecrementPostCount(ctx, 1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify count was decremented
	if user.PostCount != 4 {
		t.Errorf("Expected PostCount 4, got %d", user.PostCount)
	}
}

func TestService_IncrementCommentCount(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockUserRepository{}
	service := NewService(mockRepo)

	// Add a test user
	now := time.Now()
	user := &domain.User{
		ID:           1,
		Email:        "test@example.com",
		Username:     "testuser",
		CommentCount: 3,
		CreatedAt:    now,
		UpdatedAt:    now,
		IsActive:     true,
	}
	mockRepo.users = map[int]*domain.User{1: user}

	err := service.IncrementCommentCount(ctx, 1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify count was incremented
	if user.CommentCount != 4 {
		t.Errorf("Expected CommentCount 4, got %d", user.CommentCount)
	}
}

func TestService_DecrementCommentCount(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockUserRepository{}
	service := NewService(mockRepo)

	// Add a test user
	now := time.Now()
	user := &domain.User{
		ID:           1,
		Email:        "test@example.com",
		Username:     "testuser",
		CommentCount: 3,
		CreatedAt:    now,
		UpdatedAt:    now,
		IsActive:     true,
	}
	mockRepo.users = map[int]*domain.User{1: user}

	err := service.DecrementCommentCount(ctx, 1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify count was decremented
	if user.CommentCount != 2 {
		t.Errorf("Expected CommentCount 2, got %d", user.CommentCount)
	}
}

func TestService_CreateUser(t *testing.T) {
	mockRepo := &MockUserRepository{}
	service := NewService(mockRepo)
	ctx := context.Background()

	// Test successful user creation
	userID, err := service.CreateUser(ctx, "test@example.com", "testuser", "hashedpassword")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	// User ID is set by the mock implementation to 0 (default), which is fine
	_ = userID
}

func TestService_CreateUser_WithError(t *testing.T) {
	mockRepo := &MockUserRepository{
		createFn: func(ctx context.Context, user *domain.User) error {
			return domain.ErrUserNotFound
		},
	}
	service := NewService(mockRepo)
	ctx := context.Background()

	_, err := service.CreateUser(ctx, "test@example.com", "testuser", "hashedpassword")
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestService_GetByID(t *testing.T) {
	now := time.Now()
	mockRepo := &MockUserRepository{
		users: map[int]*domain.User{
			1: {
				ID:        1,
				PublicID:  "test-public-id",
				Email:     "test@example.com",
				Username:  "testuser",
				Role:      domain.RoleUser,
				CreatedAt: now,
				UpdatedAt: now,
				IsActive:  true,
			},
		},
	}
	service := NewService(mockRepo)
	ctx := context.Background()

	user, err := service.GetByID(ctx, 1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if user == nil {
		t.Error("Expected user, got nil")
	}
	if user != nil && user.ID != 1 {
		t.Errorf("Expected user ID 1, got %d", user.ID)
	}
}

func TestService_GetByID_NotFound(t *testing.T) {
	mockRepo := &MockUserRepository{
		users: map[int]*domain.User{},
	}
	service := NewService(mockRepo)
	ctx := context.Background()

	user, err := service.GetByID(ctx, 999)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if user != nil {
		t.Errorf("Expected nil user, got %+v", user)
	}
}

func TestService_ExistsByEmail(t *testing.T) {
	mockRepo := &MockUserRepository{
		existsByEmailFn: func(ctx context.Context, email string) (bool, error) {
			return email == "exists@example.com", nil
		},
	}
	service := NewService(mockRepo)
	ctx := context.Background()

	// Test existing email
	exists, err := service.ExistsByEmail(ctx, "exists@example.com")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !exists {
		t.Error("Expected email to exist")
	}

	// Test non-existing email
	exists, err = service.ExistsByEmail(ctx, "notexists@example.com")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if exists {
		t.Error("Expected email to not exist")
	}
}

func TestService_ExistsByUsername(t *testing.T) {
	mockRepo := &MockUserRepository{
		existsByUsernameFn: func(ctx context.Context, username string) (bool, error) {
			return username == "existinguser", nil
		},
	}
	service := NewService(mockRepo)
	ctx := context.Background()

	// Test existing username
	exists, err := service.ExistsByUsername(ctx, "existinguser")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !exists {
		t.Error("Expected username to exist")
	}

	// Test non-existing username
	exists, err = service.ExistsByUsername(ctx, "nonexistinguser")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if exists {
		t.Error("Expected username to not exist")
	}
}
