// OUTPUT PORT - Repository Interface
// [OPTIONAL FEATURE: forum-advanced-features]
// Package ports defines the output ports (data access contracts) for the notification module.
package ports

import (
	"context"
	"forum/internal/modules/notification/domain"
)

// NotificationRepository defines the data access contract for notifications.
type NotificationRepository interface {
	// Create stores a new notification in the repository.
	// Must generate and set PublicID (UUID) before persisting.
	Create(ctx context.Context, notification *domain.Notification) error

	// GetByUserID retrieves all notifications for a user.
	// Uses internal userID for query
	GetByUserID(ctx context.Context, userID int) ([]*domain.Notification, error)

	// MarkAsReadByPublicID marks a notification as read by its public UUID.
	// userID scopes the update to the notification owner for security.
	MarkAsReadByPublicID(ctx context.Context, userID int, notificationPublicID string) error

	// MarkAllAsReadByUserID marks all notifications as read for a user.
	MarkAllAsReadByUserID(ctx context.Context, userID int) error

	// CountUnread returns the number of unread notifications for a user.
	CountUnread(ctx context.Context, userID int) (int, error)
}
