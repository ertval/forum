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
	Create(ctx context.Context, notification *domain.Notification) error

	// GetByUserID retrieves all notifications for a user.
	GetByUserID(ctx context.Context, userID int) ([]*domain.Notification, error)

	// MarkAsRead marks a notification as read.
	MarkAsRead(ctx context.Context, notificationID int) error
}
