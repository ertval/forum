// INPUT ADAPTER TEST - HTTP API Handler Tests
package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"forum/internal/modules/user/domain"
)

// MockUserService implements the UserService interface for testing
type MockUserService struct {
	getByIDFn               func(ctx context.Context, userID int) (*domain.User, error)
	getByPublicIDFn         func(ctx context.Context, publicID string) (*domain.User, error)
	getByUsernameFn         func(ctx context.Context, username string) (*domain.User, error)
	getByEmailFn            func(ctx context.Context, email string) (*domain.User, error)
	createUserFn            func(ctx context.Context, email, username, passwordHash string) (int, error)
	updateRoleFn            func(ctx context.Context, userID int, newRole domain.Role) error
	deactivateUserFn        func(ctx context.Context, userID int) error
	activateUserFn          func(ctx context.Context, userID int) error
	listUsersFn             func(ctx context.Context, offset, limit int) ([]*domain.User, error)
	incrementPostCountFn    func(ctx context.Context, userID int) error
	decrementPostCountFn    func(ctx context.Context, userID int) error
	incrementCommentCountFn func(ctx context.Context, userID int) error
	decrementCommentCountFn func(ctx context.Context, userID int) error
	existsByEmailFn         func(ctx context.Context, email string) (bool, error)
	existsByUsernameFn      func(ctx context.Context, username string) (bool, error)
	incrementReactionCountFn func(ctx context.Context, userID int) error
	decrementReactionCountFn func(ctx context.Context, userID int) error
}

func (m *MockUserService) GetByID(ctx context.Context, userID int) (*domain.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, userID)
	}
	return nil, nil
}

func (m *MockUserService) GetByPublicID(ctx context.Context, publicID string) (*domain.User, error) {
	if m.getByPublicIDFn != nil {
		return m.getByPublicIDFn(ctx, publicID)
	}
	return nil, nil
}

func (m *MockUserService) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	if m.getByUsernameFn != nil {
		return m.getByUsernameFn(ctx, username)
	}
	return nil, nil
}

func (m *MockUserService) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.getByEmailFn != nil {
		return m.getByEmailFn(ctx, email)
	}
	return nil, nil
}

func (m *MockUserService) CreateUser(ctx context.Context, email, username, passwordHash string) (int, error) {
	if m.createUserFn != nil {
		return m.createUserFn(ctx, email, username, passwordHash)
	}
	return 1, nil
}

func (m *MockUserService) UpdateRole(ctx context.Context, userID int, newRole domain.Role) error {
	if m.updateRoleFn != nil {
		return m.updateRoleFn(ctx, userID, newRole)
	}
	return nil
}

func (m *MockUserService) DeactivateUser(ctx context.Context, userID int) error {
	if m.deactivateUserFn != nil {
		return m.deactivateUserFn(ctx, userID)
	}
	return nil
}

func (m *MockUserService) ActivateUser(ctx context.Context, userID int) error {
	if m.activateUserFn != nil {
		return m.activateUserFn(ctx, userID)
	}
	return nil
}

func (m *MockUserService) ListUsers(ctx context.Context, offset, limit int) ([]*domain.User, error) {
	if m.listUsersFn != nil {
		return m.listUsersFn(ctx, offset, limit)
	}
	return []*domain.User{}, nil
}

func (m *MockUserService) IncrementPostCount(ctx context.Context, userID int) error {
	if m.incrementPostCountFn != nil {
		return m.incrementPostCountFn(ctx, userID)
	}
	return nil
}

func (m *MockUserService) DecrementPostCount(ctx context.Context, userID int) error {
	if m.decrementPostCountFn != nil {
		return m.decrementPostCountFn(ctx, userID)
	}
	return nil
}

func (m *MockUserService) IncrementCommentCount(ctx context.Context, userID int) error {
	if m.incrementCommentCountFn != nil {
		return m.incrementCommentCountFn(ctx, userID)
	}
	return nil
}

func (m *MockUserService) DecrementCommentCount(ctx context.Context, userID int) error {
	if m.decrementCommentCountFn != nil {
		return m.decrementCommentCountFn(ctx, userID)
	}
	return nil
}

func (m *MockUserService) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	if m.existsByEmailFn != nil {
		return m.existsByEmailFn(ctx, email)
	}
	return false, nil
}

func (m *MockUserService) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	if m.existsByUsernameFn != nil {
		return m.existsByUsernameFn(ctx, username)
	}
	return false, nil
}

func (m *MockUserService) IncrementReactionCount(ctx context.Context, userID int) error {
	if m.incrementReactionCountFn != nil {
		return m.incrementReactionCountFn(ctx, userID)
	}
	return nil
}

func (m *MockUserService) DecrementReactionCount(ctx context.Context, userID int) error {
	if m.decrementReactionCountFn != nil {
		return m.decrementReactionCountFn(ctx, userID)
	}
	return nil
}

// mockServiceContainer implements ServiceContainer for testing
type mockServiceContainer struct {
	userService *MockUserService
}

func (m *mockServiceContainer) User() interface {
	GetByID(ctx context.Context, userID int) (*domain.User, error)
	GetByPublicID(ctx context.Context, publicID string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	CreateUser(ctx context.Context, email, username, passwordHash string) (int, error)
	UpdateRole(ctx context.Context, userID int, newRole domain.Role) error
	DeactivateUser(ctx context.Context, userID int) error
	ActivateUser(ctx context.Context, userID int) error
	ListUsers(ctx context.Context, offset, limit int) ([]*domain.User, error)
	IncrementPostCount(ctx context.Context, userID int) error
	DecrementPostCount(ctx context.Context, userID int) error
	IncrementCommentCount(ctx context.Context, userID int) error
	DecrementCommentCount(ctx context.Context, userID int) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
} {
	return m.userService
}

func TestGetUserAPI_Success(t *testing.T) {
	mockService := &MockUserService{
		getByPublicIDFn: func(ctx context.Context, publicID string) (*domain.User, error) {
			if publicID == "test-uuid" {
				return &domain.User{
					ID:       1,
					PublicID: "test-uuid",
					Email:    "test@example.com",
					Username: "testuser",
					Role:     domain.RoleUser,
					IsActive: true,
				}, nil
			}
			return nil, nil
		},
	}

	handler := &HTTPHandler{userService: mockService}

	req := httptest.NewRequest(http.MethodGet, "/api/users/test-uuid", nil)
	req.SetPathValue("id", "test-uuid")
	rec := httptest.NewRecorder()

	handler.GetUserAPI(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if response["username"] != "testuser" {
		t.Errorf("Expected username 'testuser', got %v", response["username"])
	}
}

func TestGetUserAPI_NotFound(t *testing.T) {
	mockService := &MockUserService{
		getByPublicIDFn: func(ctx context.Context, publicID string) (*domain.User, error) {
			return nil, nil
		},
	}

	handler := &HTTPHandler{userService: mockService}

	req := httptest.NewRequest(http.MethodGet, "/api/users/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")
	rec := httptest.NewRecorder()

	handler.GetUserAPI(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rec.Code)
	}
}

func TestGetUserAPI_MissingID(t *testing.T) {
	mockService := &MockUserService{}
	handler := &HTTPHandler{userService: mockService}

	req := httptest.NewRequest(http.MethodGet, "/api/users/", nil)
	req.SetPathValue("id", "")
	rec := httptest.NewRecorder()

	handler.GetUserAPI(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestListUsersAPI_Success(t *testing.T) {
	mockService := &MockUserService{
		listUsersFn: func(ctx context.Context, offset, limit int) ([]*domain.User, error) {
			return []*domain.User{
				{ID: 1, PublicID: "uuid-1", Username: "user1", Email: "user1@example.com"},
				{ID: 2, PublicID: "uuid-2", Username: "user2", Email: "user2@example.com"},
			}, nil
		},
	}

	handler := &HTTPHandler{userService: mockService}

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rec := httptest.NewRecorder()

	handler.ListUsersAPI(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if response["count"].(float64) != 2 {
		t.Errorf("Expected count 2, got %v", response["count"])
	}
}

func TestUpdateRoleAPI_Success(t *testing.T) {
	mockService := &MockUserService{
		getByPublicIDFn: func(ctx context.Context, publicID string) (*domain.User, error) {
			return &domain.User{ID: 1, PublicID: publicID, Role: domain.RoleUser}, nil
		},
		updateRoleFn: func(ctx context.Context, userID int, newRole domain.Role) error {
			return nil
		},
	}

	handler := &HTTPHandler{userService: mockService}

	body := bytes.NewBufferString(`{"role": "moderator"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/users/test-uuid/role", body)
	req.SetPathValue("id", "test-uuid")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateRoleAPI(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestUpdateRoleAPI_InvalidRole(t *testing.T) {
	mockService := &MockUserService{
		getByPublicIDFn: func(ctx context.Context, publicID string) (*domain.User, error) {
			return &domain.User{ID: 1, PublicID: publicID}, nil
		},
	}

	handler := &HTTPHandler{userService: mockService}

	body := bytes.NewBufferString(`{"role": "invalid_role"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/users/test-uuid/role", body)
	req.SetPathValue("id", "test-uuid")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateRoleAPI(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestUpdateRoleAPI_UserNotFound(t *testing.T) {
	mockService := &MockUserService{
		getByPublicIDFn: func(ctx context.Context, publicID string) (*domain.User, error) {
			return nil, nil
		},
	}

	handler := &HTTPHandler{userService: mockService}

	body := bytes.NewBufferString(`{"role": "moderator"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/users/nonexistent/role", body)
	req.SetPathValue("id", "nonexistent")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateRoleAPI(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rec.Code)
	}
}

func TestDeactivateUserAPI_Success(t *testing.T) {
	mockService := &MockUserService{
		getByPublicIDFn: func(ctx context.Context, publicID string) (*domain.User, error) {
			return &domain.User{ID: 1, PublicID: publicID, IsActive: true}, nil
		},
		deactivateUserFn: func(ctx context.Context, userID int) error {
			return nil
		},
	}

	handler := &HTTPHandler{userService: mockService}

	req := httptest.NewRequest(http.MethodPut, "/api/users/test-uuid/deactivate", nil)
	req.SetPathValue("id", "test-uuid")
	rec := httptest.NewRecorder()

	handler.DeactivateUserAPI(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestDeactivateUserAPI_UserNotFound(t *testing.T) {
	mockService := &MockUserService{
		getByPublicIDFn: func(ctx context.Context, publicID string) (*domain.User, error) {
			return nil, nil
		},
	}

	handler := &HTTPHandler{userService: mockService}

	req := httptest.NewRequest(http.MethodPut, "/api/users/nonexistent/deactivate", nil)
	req.SetPathValue("id", "nonexistent")
	rec := httptest.NewRecorder()

	handler.DeactivateUserAPI(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rec.Code)
	}
}

func TestActivateUserAPI_Success(t *testing.T) {
	mockService := &MockUserService{
		getByPublicIDFn: func(ctx context.Context, publicID string) (*domain.User, error) {
			return &domain.User{ID: 1, PublicID: publicID, IsActive: false}, nil
		},
		activateUserFn: func(ctx context.Context, userID int) error {
			return nil
		},
	}

	handler := &HTTPHandler{userService: mockService}

	req := httptest.NewRequest(http.MethodPut, "/api/users/test-uuid/activate", nil)
	req.SetPathValue("id", "test-uuid")
	rec := httptest.NewRecorder()

	handler.ActivateUserAPI(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestActivateUserAPI_UserNotFound(t *testing.T) {
	mockService := &MockUserService{
		getByPublicIDFn: func(ctx context.Context, publicID string) (*domain.User, error) {
			return nil, nil
		},
	}

	handler := &HTTPHandler{userService: mockService}

	req := httptest.NewRequest(http.MethodPut, "/api/users/nonexistent/activate", nil)
	req.SetPathValue("id", "nonexistent")
	rec := httptest.NewRecorder()

	handler.ActivateUserAPI(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rec.Code)
	}
}
