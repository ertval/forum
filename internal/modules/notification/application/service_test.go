package application

import (
	"context"
	"forum/internal/modules/notification/domain"
	"forum/internal/modules/notification/ports"
	"testing"
	"time"
)

// MockNotificationRepository implements NotificationRepository for testing
type MockNotificationRepository struct {
	notifications     map[int]*domain.Notification
	getByUserFn       func(ctx context.Context, userID int) ([]*domain.Notification, error)
	markAsReadFn      func(ctx context.Context, notificationID int) error
	createFn          func(ctx context.Context, notification *domain.Notification) error
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

func (m *MockNotificationRepository) MarkAsRead(ctx context.Context, notificationID int) error {
	if m.markAsReadFn != nil {
		return m.markAsReadFn(ctx, notificationID)
	}
	
	if m.notifications != nil {
		if notification, exists := m.notifications[notificationID]; exists {
			notification.MarkAsRead()
		}
	}
	return nil
}

func (m *MockNotificationRepository) Create(ctx context.Context, notification *domain.Notification) error {
	if m.createFn != nil {
		return m.createFn(ctx, notification)
	}
	
	if m.notifications == nil {
		m.notifications = make(map[int]*domain.Notification)
	}
	m.notifications[notification.ID] = notification
	return nil
}

func TestService_CreateNotification(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockNotificationRepository{}
	service := NewService(mockRepo)

	// Test the current implementation (returns nil since it's a placeholder)
	err := service.CreateNotification(ctx, 1, domain.TypeLike, "Someone liked your post", 10)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestService_GetUserNotifications(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockNotificationRepository{}
	service := NewService(mockRepo)

	// Add test notifications to the mock
	now := time.Now()
	notifications := []*domain.Notification{
		{ID: 1, UserID: 5, Type: domain.TypeLike, Message: "Someone liked your post", TargetID: 10, IsRead: false, CreatedAt: now},
		{ID: 2, UserID: 5, Type: domain.TypeComment, Message: "Someone commented on your post", TargetID: 10, IsRead: true, CreatedAt: now},
		{ID: 3, UserID: 6, Type: domain.TypeReply, Message: "Someone replied to your comment", TargetID: 15, IsRead: false, CreatedAt: now}, // Different user
	}
	mockRepo.notifications = map[int]*domain.Notification{}
	for _, notification := range notifications {
		mockRepo.notifications[notification.ID] = notification
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
		ID:        1,
		UserID:    5,
		Type:      domain.TypeLike,
		Message:   "Someone liked your post",
		TargetID:  10,
		IsRead:    false,
		CreatedAt: now,
	}
	mockRepo.notifications = map[int]*domain.Notification{
		1: notification,
	}

	// Verify it starts as unread
	if notification.IsRead {
		t.Error("Expected notification to be unread initially")
	}

	err := service.MarkAsRead(ctx, 1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify it was marked as read
	if !notification.IsRead {
		t.Error("Expected notification to be marked as read")
	}
}