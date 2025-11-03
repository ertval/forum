package output
// Package output defines the outbound ports for the notification module.
package output

import (
	"context"

	"forum/internal/modules/notification/domain"
)

// NotificationRepository defines the interface for notification persistence.
type NotificationRepository interface {
	Create(ctx context.Context, notification *domain.Notification) error
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.Notification, error)
	GetUnreadCount(ctx context.Context, userID string) (int, error)
	Update(ctx context.Context, notification *domain.Notification) error
	MarkAllAsRead(ctx context.Context, userID string) error
	Delete(ctx context.Context, notificationID string) error
}
