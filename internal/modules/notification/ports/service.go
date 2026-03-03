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
	// userID: recipient internal user ID, actorID: internal user ID that triggered event.
	// targetPublicID: public UUID of related entity.
	CreateNotification(ctx context.Context, userID, actorID int, notifType, message string, targetPublicID string) error

	// GetUserNotifications retrieves all notifications for a user.
	// Uses internal userID from session
	GetUserNotifications(ctx context.Context, userID int) ([]*domain.Notification, error)

	// MarkAsRead marks a notification as read by its public UUID.
	// userID scopes the update to the notification owner for security.
	MarkAsRead(ctx context.Context, userID int, notificationPublicID string) error

	// MarkAllAsRead marks all notifications as read for a user.
	MarkAllAsRead(ctx context.Context, userID int) error

	// CountUnread returns the number of unread notifications for a user.
	CountUnread(ctx context.Context, userID int) (int, error)
}
