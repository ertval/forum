// [OPTIONAL FEATURE: forum-advanced-features]
// Package application implements notification service business logic.
package application

import (
	"context"
	"forum/internal/modules/notification/domain"
	"forum/internal/modules/notification/ports"
	"time"
)

// Service implements the NotificationService interface.
type Service struct {
	notificationRepo ports.NotificationRepository
}

// NewService creates a new notification service.
func NewService(notificationRepo ports.NotificationRepository) *Service {
	return &Service{notificationRepo: notificationRepo}
}

// CreateNotification creates a new notification for a user.
func (s *Service) CreateNotification(ctx context.Context, userID, actorID int, notifType, message string, targetPublicID string) error {
	if userID <= 0 || actorID <= 0 {
		return domain.ErrInvalidUserID
	}

	if targetPublicID == "" {
		return domain.ErrInvalidTarget
	}

	switch notifType {
	case domain.TypeLike, domain.TypeDislike, domain.TypeComment, domain.TypeReply:
	default:
		return domain.ErrInvalidNotificationType
	}

	notification := &domain.Notification{
		UserID:         userID,
		ActorID:        actorID,
		Type:           notifType,
		Message:        message,
		IsRead:         false,
		CreatedAt:      time.Now(),
		PublicTargetID: targetPublicID,
	}

	return s.notificationRepo.Create(ctx, notification)
}

// GetUserNotifications retrieves all notifications for a user.
func (s *Service) GetUserNotifications(ctx context.Context, userID int) ([]*domain.Notification, error) {
	return s.notificationRepo.GetByUserID(ctx, userID)
}

// MarkAsRead marks a notification as read by its public UUID.
// userID scopes the update to the notification owner for security.
func (s *Service) MarkAsRead(ctx context.Context, userID int, notificationPublicID string) error {
	return s.notificationRepo.MarkAsReadByPublicID(ctx, userID, notificationPublicID)
}

// MarkAllAsRead marks all notifications as read for a user.
func (s *Service) MarkAllAsRead(ctx context.Context, userID int) error {
	if userID <= 0 {
		return domain.ErrInvalidUserID
	}
	return s.notificationRepo.MarkAllAsReadByUserID(ctx, userID)
}

// CountUnread returns the number of unread notifications for a user.
func (s *Service) CountUnread(ctx context.Context, userID int) (int, error) {
	if userID <= 0 {
		return 0, domain.ErrInvalidUserID
	}
	return s.notificationRepo.CountUnread(ctx, userID)
}
