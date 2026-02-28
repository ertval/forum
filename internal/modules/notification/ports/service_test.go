package ports

import (
	"context"
	"forum/internal/modules/notification/domain"
	"testing"
)

// This test file verifies that the interfaces are properly defined and can be implemented
func TestNotificationServiceInterface(t *testing.T) {
	var _ NotificationService = (*mockNotificationService)(nil)
}

func TestNotificationRepositoryInterface(t *testing.T) {
	var _ NotificationRepository = (*mockNotificationRepository)(nil)
}

// Mock implementations for interface compatibility testing
type mockNotificationService struct{}

func (m *mockNotificationService) CreateNotification(ctx context.Context, userID, actorID int, notifType, message string, targetPublicID string) error {
	return nil
}

func (m *mockNotificationService) GetUserNotifications(ctx context.Context, userID int) ([]*domain.Notification, error) {
	return nil, nil
}

func (m *mockNotificationService) MarkAsRead(ctx context.Context, notificationPublicID string) error {
	return nil
}

type mockNotificationRepository struct{}

func (m *mockNotificationRepository) GetByUserID(ctx context.Context, userID int) ([]*domain.Notification, error) {
	return nil, nil
}

func (m *mockNotificationRepository) MarkAsReadByPublicID(ctx context.Context, notificationPublicID string) error {
	return nil
}

func (m *mockNotificationRepository) Create(ctx context.Context, notification *domain.Notification) error {
	return nil
}

func TestNotificationServiceInterfaceMethods(t *testing.T) {
	// Create context for testing
	ctx := context.Background()

	// Test that we can call interface methods on a variable of the interface type
	service := &mockNotificationService{}

	// Test each method signature
	// Use public target id string
	err := service.CreateNotification(ctx, 1, 2, "type", "message", "target-1")
	if err != nil {
		// Expected to be not implemented in mock
	}

	_, err = service.GetUserNotifications(ctx, 1)
	if err != nil {
		// Expected to be not implemented in mock
	}

	err = service.MarkAsRead(ctx, "pub-1")
	if err != nil {
		// Expected to be not implemented in mock
	}
}

func TestNotificationRepositoryInterfaceMethods(t *testing.T) {
	// Create context for testing
	ctx := context.Background()

	// Create mock repository
	repo := &mockNotificationRepository{}

	// Test that we can call interface methods on a variable of the interface type
	var notification *domain.Notification
	var notifications []*domain.Notification
	var err error

	// Test GetByUserID method
	notifications, err = repo.GetByUserID(ctx, 1)
	if err != nil {
		// Expected to be not implemented in mock
	}

	// Test MarkAsRead method
	err = repo.MarkAsReadByPublicID(ctx, "pub-1")
	if err != nil {
		// Expected to be not implemented in mock
	}

	// Test Create method
	err = repo.Create(ctx, notification)
	if err != nil {
		// Expected to be not implemented in mock
	}

	_ = notifications // Use the variable to avoid unused variable warning
}
