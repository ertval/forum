package application

import (
	"context"
	"forum/internal/modules/notification/domain"
	"testing"
	"time"
)

// MockNotificationRepository implements NotificationRepository for testing
type MockNotificationRepository struct {
	notifications map[string]*domain.Notification
	getByUserFn   func(ctx context.Context, userID int) ([]*domain.Notification, error)
	markAsReadFn  func(ctx context.Context, notificationPublicID string) error
	createFn      func(ctx context.Context, notification *domain.Notification) error
}

func (m *MockNotificationRepository) GetByUserID(ctx context.Context, userID int) ([]*domain.Notification, error) {
	if m.getByUserFn != nil {
		return m.getByUserFn(ctx, userID)
	}

	var result []*domain.Notification
	for _, notification := range m.notifications {
		if notification.UserID == userID {
			result = append(result, notification)
		}
	}
	return result, nil
}

func (m *MockNotificationRepository) MarkAsReadByPublicID(ctx context.Context, notificationPublicID string) error {
	if m.markAsReadFn != nil {
		return m.markAsReadFn(ctx, notificationPublicID)
	}

	if m.notifications != nil {
		if notification, exists := m.notifications[notificationPublicID]; exists {
			notification.MarkAsRead()
		}
	}
	return nil
}

func (m *MockNotificationRepository) MarkAllAsReadByUserID(ctx context.Context, userID int) error {
	for _, n := range m.notifications {
		if n.UserID == userID {
			n.MarkAsRead()
		}
	}
	return nil
}

func (m *MockNotificationRepository) Create(ctx context.Context, notification *domain.Notification) error {
	if m.createFn != nil {
		return m.createFn(ctx, notification)
	}

	// Store by PublicID to match repository contract
	if m.notifications == nil {
		m.notifications = make(map[string]*domain.Notification)
	}
	m.notifications[notification.PublicID] = notification
	return nil
}

func TestService_CreateNotification(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockNotificationRepository{}
	service := NewService(mockRepo)

	err := service.CreateNotification(ctx, 1, 2, domain.TypeLike, "Someone liked your post", "target-10")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	err = service.CreateNotification(ctx, 1, 2, domain.TypeDislike, "Someone disliked your post", "target-10")
	if err != nil {
		t.Errorf("Expected no error for dislike, got %v", err)
	}

	err = service.CreateNotification(ctx, 0, 2, domain.TypeLike, "Someone liked your post", "target-10")
	if err == nil {
		t.Error("Expected error for invalid user id")
	}

	err = service.CreateNotification(ctx, 1, 0, domain.TypeLike, "Someone liked your post", "target-10")
	if err == nil {
		t.Error("Expected error for invalid actor id")
	}

	err = service.CreateNotification(ctx, 1, 2, "invalid", "message", "target-10")
	if err == nil {
		t.Error("Expected error for invalid notification type")
	}
}

func TestService_GetUserNotifications(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockNotificationRepository{}
	service := NewService(mockRepo)

	// Add test notifications to the mock
	now := time.Now()
	notifications := []*domain.Notification{
		{ID: 1, PublicID: "pub-1", UserID: 5, Type: domain.TypeLike, Message: "Someone liked your post", TargetID: 10, PublicTargetID: "target-10", IsRead: false, CreatedAt: now},
		{ID: 2, PublicID: "pub-2", UserID: 5, Type: domain.TypeComment, Message: "Someone commented on your post", TargetID: 10, PublicTargetID: "target-10", IsRead: true, CreatedAt: now},
		{ID: 3, PublicID: "pub-3", UserID: 6, Type: domain.TypeReply, Message: "Someone replied to your comment", TargetID: 15, PublicTargetID: "target-15", IsRead: false, CreatedAt: now}, // Different user
	}
	mockRepo.notifications = map[string]*domain.Notification{}
	for _, notification := range notifications {
		mockRepo.notifications[notification.PublicID] = notification
	}

	t.Run("get notifications for user", func(t *testing.T) {
		result, err := service.GetUserNotifications(ctx, 5)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(result) != 2 {
			t.Errorf("Expected 2 notifications for user 5, got %d", len(result))
		}

		// Verify all returned notifications belong to the correct user
		for _, notification := range result {
			if notification.UserID != 5 {
				t.Errorf("Expected UserID 5, got %d", notification.UserID)
			}
		}
	})

	t.Run("get notifications for user with no notifications", func(t *testing.T) {
		result, err := service.GetUserNotifications(ctx, 999)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(result) != 0 {
			t.Errorf("Expected 0 notifications for non-existent user, got %d", len(result))
		}
	})
}

func TestService_MarkAsRead(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockNotificationRepository{}
	service := NewService(mockRepo)

	// Add a test notification to the mock
	now := time.Now()
	notification := &domain.Notification{
		ID:             1,
		PublicID:       "pub-1",
		UserID:         5,
		ActorID:        6,
		Type:           domain.TypeLike,
		Message:        "Someone liked your post",
		TargetID:       10,
		PublicTargetID: "target-10",
		IsRead:         false,
		CreatedAt:      now,
	}
	mockRepo.notifications = map[string]*domain.Notification{
		"pub-1": notification,
	}

	// Verify it starts as unread
	if notification.IsRead {
		t.Error("Expected notification to be unread initially")
	}

	err := service.MarkAsRead(ctx, "pub-1")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify it was marked as read
	if !notification.IsRead {
		t.Error("Expected notification to be marked as read")
	}
}
