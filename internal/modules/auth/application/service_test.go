package application

import (
	"context"
	"errors"
	"fmt"
	"forum/internal/modules/auth/domain"
	userDomain "forum/internal/modules/user/domain"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"golang.org/x/crypto/bcrypt"
)

// MockUserService implements user ports UserService for testing
type MockUserService struct {
	users           map[string]*userDomain.User
	emailExists     map[string]bool
	usernameExists  map[string]bool
	createError     error
	getByEmailError error
	nextID          int // Track next ID for Create
}

func (m *MockUserService) CreateUser(ctx context.Context, email, username, passwordHash string) (userID int, err error) {
	if m.createError != nil {
		return 0, m.createError
	}
	if m.users == nil {
		m.users = make(map[string]*userDomain.User)
		m.nextID = 1
	}
	m.nextID++
	user := &userDomain.User{
		ID:           m.nextID,
		PublicID:     fmt.Sprintf("user-public-id-%d", m.nextID),
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
		Role:         userDomain.RoleUser,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}
	m.users[email] = user
	return user.ID, nil
}

func (m *MockUserService) GetByID(ctx context.Context, userID int) (*userDomain.User, error) {
	for _, user := range m.users {
		if user.ID == userID {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *MockUserService) GetByPublicID(ctx context.Context, publicID string) (*userDomain.User, error) {
	for _, user := range m.users {
		if user.PublicID == publicID {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *MockUserService) GetByUsername(ctx context.Context, username string) (*userDomain.User, error) {
	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *MockUserService) GetByEmail(ctx context.Context, email string) (*userDomain.User, error) {
	if m.getByEmailError != nil {
		return nil, m.getByEmailError
	}
	if user, exists := m.users[email]; exists {
		return user, nil
	}
	return nil, errors.New("user not found")
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

func (m *MockUserService) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	if m.emailExists != nil {
		return m.emailExists[email], nil
	}
	_, exists := m.users[email]
	return exists, nil
}

func (m *MockUserService) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	if m.usernameExists != nil {
		return m.usernameExists[username], nil
	}
	for _, user := range m.users {
		if user.Username == username {
			return true, nil
		}
	}
	return false, nil
}

// MockSessionRepository implements auth ports SessionRepository for testing
type MockSessionRepository struct {
	sessions        map[string]*domain.Session
	createError     error
	getByTokenError error
	updateError     error
	deleteError     error
}

func (m *MockSessionRepository) Create(ctx context.Context, session *domain.Session) error {
	if m.createError != nil {
		return m.createError
	}
	if m.sessions == nil {
		m.sessions = make(map[string]*domain.Session)
	}
	m.sessions[session.Token] = session
	return nil
}

func (m *MockSessionRepository) GetByToken(ctx context.Context, token string) (*domain.Session, error) {
	if m.getByTokenError != nil {
		return nil, m.getByTokenError
	}
	if session, exists := m.sessions[token]; exists {
		return session, nil
	}
	return nil, domain.ErrSessionNotFound
}

func (m *MockSessionRepository) GetByUserID(ctx context.Context, userID int) ([]*domain.Session, error) {
	var userSessions []*domain.Session
	for _, session := range m.sessions {
		if session.UserID == userID {
			userSessions = append(userSessions, session)
		}
	}
	return userSessions, nil
}

func (m *MockSessionRepository) Update(ctx context.Context, session *domain.Session) error {
	if m.updateError != nil {
		return m.updateError
	}
	if m.sessions == nil {
		m.sessions = make(map[string]*domain.Session)
	}
	m.sessions[session.Token] = session
	return nil
}

func (m *MockSessionRepository) Delete(ctx context.Context, token string) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	delete(m.sessions, token)
	return nil
}

func (m *MockSessionRepository) DeleteByUserID(ctx context.Context, userID int) error {
	for token, session := range m.sessions {
		if session.UserID == userID {
			delete(m.sessions, token)
		}
	}
	return nil
}

func (m *MockSessionRepository) DeleteExpired(ctx context.Context) error {
	// Not implemented for this test
	return nil
}

func TestService_Register(t *testing.T) {
	ctx := context.Background()
	mockUserService := &MockUserService{}
	mockSessionRepo := &MockSessionRepository{}
	service := NewService(mockSessionRepo, mockUserService, 24*time.Hour)

	tests := []struct {
		name          string
		email         string
		username      string
		password      string
		expectedError error
		setup         func()
	}{
		{
			name:          "successful registration",
			email:         "test@example.com",
			username:      "testuser",
			password:      "password123",
			expectedError: nil,
		},
		{
			name:          "invalid email format",
			email:         "invalid-email",
			username:      "testuser",
			password:      "password123",
			expectedError: domain.ErrInvalidEmail,
		},
		{
			name:          "empty email",
			email:         "",
			username:      "testuser",
			password:      "password123",
			expectedError: domain.ErrInvalidEmail,
		},
		{
			name:          "empty password",
			email:         "test@example.com",
			username:      "testuser",
			password:      "",
			expectedError: domain.ErrWeakPassword,
		},
		{
			name:          "weak password",
			email:         "test@example.com",
			username:      "testuser",
			password:      "123", // Too short
			expectedError: domain.ErrWeakPassword,
		},
		{
			name:          "invalid username",
			email:         "test@example.com",
			username:      "invalid@username", // Contains @
			password:      "password123",
			expectedError: domain.ErrInvalidUsername,
		},
		{
			name:          "email already exists",
			email:         "existing@example.com",
			username:      "testuser",
			password:      "password123",
			expectedError: domain.ErrUserAlreadyExists,
			setup: func() {
				mockUserService.emailExists = map[string]bool{
					"existing@example.com": true,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockUserService = &MockUserService{}
			mockSessionRepo = &MockSessionRepository{}
			service = NewService(mockSessionRepo, mockUserService, 24*time.Hour)

			if tt.setup != nil {
				tt.setup()
			}

			userID, session, err := service.Register(ctx, tt.email, tt.username, tt.password)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("Expected error %v, but got nil", tt.expectedError)
				} else if !errors.Is(err, tt.expectedError) {
					t.Errorf("Expected error %v, but got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got %v", err)
				}
				if userID == 0 {
					t.Error("Expected a valid user ID")
				}
				if session != nil {
					if session.Token == "" {
						t.Error("Expected a valid session token")
					}
					if session.UserID != userID {
						t.Error("Session user ID doesn't match returned user ID")
					}
				} else {
					t.Error("Expected a valid session")
				}
			}
		})
	}
}

func TestService_Login(t *testing.T) {
	ctx := context.Background()
	mockUserService := &MockUserService{}
	mockSessionRepo := &MockSessionRepository{}
	service := NewService(mockSessionRepo, mockUserService, 24*time.Hour)

	// Create a user for testing
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	testUser := &userDomain.User{
		ID:           1,
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: string(passwordHash),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}
	mockUserService.users = map[string]*userDomain.User{
		"test@example.com": testUser,
	}

	tests := []struct {
		name          string
		email         string
		password      string
		expectedError error
	}{
		{
			name:          "successful login",
			email:         "test@example.com",
			password:      "password123",
			expectedError: nil,
		},
		{
			name:          "invalid email",
			email:         "nonexistent@example.com",
			password:      "password123",
			expectedError: domain.ErrInvalidCredentials,
		},
		{
			name:          "invalid password",
			email:         "test@example.com",
			password:      "wrongpassword",
			expectedError: domain.ErrInvalidCredentials,
		},
		{
			name:          "invalid email format",
			email:         "invalid-email",
			password:      "password123",
			expectedError: domain.ErrInvalidCredentials,
		},
		{
			name:          "empty email",
			email:         "",
			password:      "password123",
			expectedError: domain.ErrInvalidCredentials,
		},
		{
			name:          "empty password",
			email:         "test@example.com",
			password:      "",
			expectedError: domain.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := service.Login(ctx, tt.email, tt.password)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("Expected error %v, but got nil", tt.expectedError)
				} else if !errors.Is(err, tt.expectedError) {
					t.Errorf("Expected error %v, but got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got %v", err)
				}
				if session != nil {
					if session.Token == "" {
						t.Error("Expected a valid session token")
					}
					if session.UserID != testUser.ID {
						t.Error("Session user ID doesn't match user ID")
					}
				} else {
					t.Error("Expected a valid session")
				}
			}
		})
	}
}

func TestService_Logout(t *testing.T) {
	ctx := context.Background()
	mockUserService := &MockUserService{}
	mockSessionRepo := &MockSessionRepository{}
	service := NewService(mockSessionRepo, mockUserService, 24*time.Hour)

	// Create a session for testing
	testSession := &domain.Session{
		ID:     1,
		Token:  "test-session-token",
		UserID: 1,
	}
	mockSessionRepo.sessions = map[string]*domain.Session{
		"test-session-token": testSession,
	}

	err := service.Logout(ctx, "test-session-token")
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	// Check that the session was deleted
	_, err = mockSessionRepo.GetByToken(ctx, "test-session-token")
	if err == nil {
		t.Error("Expected session to be deleted")
	}
}

func TestService_ValidateSession(t *testing.T) {
	ctx := context.Background()
	mockUserService := &MockUserService{}
	mockSessionRepo := &MockSessionRepository{}
	service := NewService(mockSessionRepo, mockUserService, 24*time.Hour)

	// Create a valid session
	validSession := &domain.Session{
		ID:        1,
		Token:     "valid-session-token",
		UserID:    1,
		ExpiresAt: time.Now().Add(1 * time.Hour), // Not expired
	}
	mockSessionRepo.sessions = map[string]*domain.Session{
		"valid-session-token": validSession,
	}

	// Test valid session
	session, err := service.ValidateSession(ctx, "valid-session-token")
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	if session == nil {
		t.Error("Expected a valid session")
	}

	// Create an expired session
	expiredSession := &domain.Session{
		ID:        2,
		Token:     "expired-session-token",
		UserID:    1,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
	}
	mockSessionRepo.sessions["expired-session-token"] = expiredSession

	// Test expired session
	_, err = service.ValidateSession(ctx, "expired-session-token")
	if err == nil {
		t.Error("Expected an error for expired session")
	} else if !errors.Is(err, domain.ErrSessionExpired) {
		t.Errorf("Expected ErrSessionExpired, but got %v", err)
	}

	// Test non-existent session
	_, err = service.ValidateSession(ctx, "non-existent-token")
	if err == nil {
		t.Error("Expected an error for non-existent session")
	}
}

func TestService_RefreshSession(t *testing.T) {
	ctx := context.Background()
	mockUserService := &MockUserService{}
	mockSessionRepo := &MockSessionRepository{}
	service := NewService(mockSessionRepo, mockUserService, 24*time.Hour)

	// Create a valid session
	originalTime := time.Now().Add(1 * time.Hour)
	testSession := &domain.Session{
		ID:        1,
		Token:     "refresh-session-token",
		UserID:    1,
		ExpiresAt: originalTime,
	}
	mockSessionRepo.sessions = map[string]*domain.Session{
		"refresh-session-token": testSession,
	}

	// Test refreshing a valid session
	session, err := service.RefreshSession(ctx, "refresh-session-token")
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	if session != nil {
		if session.ExpiresAt.Before(originalTime) {
			t.Error("Expected session to have extended expiration time")
		}
	} else {
		t.Error("Expected a valid session")
	}

	// Test refreshing an expired session
	expiredSession := &domain.Session{
		ID:        2,
		Token:     "expired-refresh-token",
		UserID:    1,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
	}
	mockSessionRepo.sessions["expired-refresh-token"] = expiredSession

	_, err = service.RefreshSession(ctx, "expired-refresh-token")
	if err == nil {
		t.Error("Expected an error for expired session")
	} else if !errors.Is(err, domain.ErrSessionExpired) {
		t.Errorf("Expected ErrSessionExpired, but got %v", err)
	}
}

func TestValidateCredentials(t *testing.T) {
	tests := []struct {
		name     string
		creds    *domain.Credentials
		expected error
	}{
		{
			name: "valid credentials",
			creds: &domain.Credentials{
				Email:    "test@example.com",
				Password: "password123",
			},
			expected: nil,
		},
		{
			name: "invalid email",
			creds: &domain.Credentials{
				Email:    "invalid-email",
				Password: "password123",
			},
			expected: domain.ErrInvalidEmail,
		},
		{
			name: "empty email",
			creds: &domain.Credentials{
				Email:    "",
				Password: "password123",
			},
			expected: domain.ErrInvalidEmail,
		},
		{
			name: "empty password",
			creds: &domain.Credentials{
				Email:    "test@example.com",
				Password: "",
			},
			expected: domain.ErrWeakPassword,
		},
		{
			name: "weak password",
			creds: &domain.Credentials{
				Email:    "test@example.com",
				Password: "123", // Too short
			},
			expected: domain.ErrWeakPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCredentials(tt.creds)
			if tt.expected != nil {
				if err == nil {
					t.Errorf("Expected error %v, but got nil", tt.expected)
				} else if !errors.Is(err, tt.expected) {
					t.Errorf("Expected error %v, but got %v", tt.expected, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got %v", err)
				}
			}
		})
	}
}

func TestService_hashPassword(t *testing.T) {
	service := &Service{}

	password := "testpassword"
	hash, err := service.hashPassword(password)
	if err != nil {
		t.Fatalf("hashPassword returned error: %v", err)
	}

	if hash == password {
		t.Error("Hashed password should be different from original password")
	}

	// Verify that the hash is valid
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		t.Errorf("Hash is not valid for original password: %v", err)
	}
}

func TestService_generateSessionToken(t *testing.T) {
	service := &Service{}

	token, err := service.generateSessionToken()
	if err != nil {
		t.Fatalf("generateSessionToken returned error: %v", err)
	}

	if token == "" {
		t.Error("Generated session token should not be empty")
	}

	// Try to parse the token as a UUID to verify format
	_, err = uuid.FromString(token)
	if err != nil {
		t.Errorf("Generated token is not a valid UUID: %v", err)
	}
}
