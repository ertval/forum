package adapters

import (
	"context"
	"encoding/json"
	"forum/internal/modules/notification/domain"
	userDomain "forum/internal/modules/user/domain"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authPorts "forum/internal/modules/auth/ports"
)

type mockNotificationService struct {
	notifications []*domain.Notification
	markErr       error
}

func (m *mockNotificationService) CreateNotification(ctx context.Context, userID, actorID int, notifType, message string, targetPublicID string) error {
	return nil
}
func (m *mockNotificationService) GetUserNotifications(ctx context.Context, userID int) ([]*domain.Notification, error) {
	return m.notifications, nil
}
func (m *mockNotificationService) MarkAsRead(ctx context.Context, notificationPublicID string) error {
	return m.markErr
}

type mockUserService struct{}

func (m *mockUserService) CreateUser(ctx context.Context, email, username, passwordHash string) (userID int, err error) {
	return 0, nil
}
func (m *mockUserService) GetByID(ctx context.Context, userID int) (*userDomain.User, error) {
	return nil, nil
}
func (m *mockUserService) GetByPublicID(ctx context.Context, publicID string) (*userDomain.User, error) {
	return &userDomain.User{ID: 7, PublicID: publicID}, nil
}
func (m *mockUserService) GetByUsername(ctx context.Context, username string) (*userDomain.User, error) {
	return nil, nil
}
func (m *mockUserService) GetByEmail(ctx context.Context, email string) (*userDomain.User, error) {
	return nil, nil
}
func (m *mockUserService) UpdateRole(ctx context.Context, userID int, newRole userDomain.Role) error {
	return nil
}
func (m *mockUserService) DeactivateUser(ctx context.Context, userID int) error { return nil }
func (m *mockUserService) ActivateUser(ctx context.Context, userID int) error   { return nil }
func (m *mockUserService) ListUsers(ctx context.Context, offset, limit int) ([]*userDomain.User, error) {
	return nil, nil
}
func (m *mockUserService) IncrementPostCount(ctx context.Context, userID int) error    { return nil }
func (m *mockUserService) DecrementPostCount(ctx context.Context, userID int) error    { return nil }
func (m *mockUserService) IncrementCommentCount(ctx context.Context, userID int) error { return nil }
func (m *mockUserService) DecrementCommentCount(ctx context.Context, userID int) error { return nil }
func (m *mockUserService) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return false, nil
}
func (m *mockUserService) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return false, nil
}
func (m *mockUserService) IncrementReactionCount(ctx context.Context, userID int) error { return nil }
func (m *mockUserService) DecrementReactionCount(ctx context.Context, userID int) error { return nil }

func TestGetNotificationsAPI(t *testing.T) {
	h := &HTTPHandler{
		notificationService: &mockNotificationService{notifications: []*domain.Notification{{
			PublicID:  "notif-1",
			Type:      domain.TypeComment,
			Message:   "Someone commented",
			CreatedAt: time.Now(),
		}}},
		userService: &mockUserService{},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/notifications", nil)
	req = req.WithContext(context.WithValue(req.Context(), authPorts.UserIDKey, "user-public-id"))
	rr := httptest.NewRecorder()

	h.GetNotificationsAPI(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if payload["count"].(float64) != 1 {
		t.Fatalf("expected count 1, got %v", payload["count"])
	}
}

func TestMarkAsReadAPI_NotFound(t *testing.T) {
	h := &HTTPHandler{
		notificationService: &mockNotificationService{markErr: domain.ErrNotificationNotFound},
		userService:         &mockUserService{},
	}

	req := httptest.NewRequest(http.MethodPut, "/api/notifications/notif-404/read", nil)
	req.SetPathValue("id", "notif-404")
	rr := httptest.NewRecorder()

	h.MarkAsReadAPI(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}
