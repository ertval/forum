package input
// Package input defines the inbound ports for the notification module.
package input

import (
	"context"

	"forum/internal/modules/notification/domain"
)

// NotificationService defines the notification management use cases.
type NotificationService interface {
	// Create creates a new notification.
	Create(ctx context.Context, notification *domain.Notification) error

	// GetByUserID retrieves all notifications for a user.
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.Notification, error)

	// GetUnreadCount gets the count of unread notifications for a user.
	GetUnreadCount(ctx context.Context, userID string) (int, error)

	// MarkAsRead marks a notification as read.
	MarkAsRead(ctx context.Context, notificationID, userID string) error

	// MarkAllAsRead marks all notifications as read for a user.
	MarkAllAsRead(ctx context.Context, userID string) error

	// Delete deletes a notification.
	Delete(ctx context.Context, notificationID, userID string) error
}
