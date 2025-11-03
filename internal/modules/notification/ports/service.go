// INPUT PORT - Service Interface
// [OPTIONAL FEATURE: forum-advanced-features]
// Package ports defines the input ports (use cases) for the notification module.
package ports

import (
	"context"
	"forum/internal/modules/notification/domain"
)

// NotificationService defines notification management use cases.
type NotificationService interface {
	// CreateNotification creates a new notification for a user.
	CreateNotification(ctx context.Context, userID int, notifType, message string, targetID int) error

	// GetUserNotifications retrieves all notifications for a user.
	GetUserNotifications(ctx context.Context, userID int) ([]*domain.Notification, error)

	// MarkAsRead marks a notification as read.
	MarkAsRead(ctx context.Context, notificationID int) error
}
