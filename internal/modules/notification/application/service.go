// [OPTIONAL FEATURE: forum-advanced-features]
// Package application implements notification service business logic.
package application

import (
	"context"
	"forum/internal/modules/notification/domain"
	"forum/internal/modules/notification/ports"
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
// TODO: Implement notification creation with validation.
func (s *Service) CreateNotification(ctx context.Context, userID int, notifType, message string, targetPublicID string) error {
	// Implementation placeholder
	// 1. Validate notification type
	// 2. Resolve targetPublicID to internal target ID if needed
	// 3. Create notification entity
	// 4. Save to repository (repo generates PublicID)
	return nil
}

// GetUserNotifications retrieves all notifications for a user.
func (s *Service) GetUserNotifications(ctx context.Context, userID int) ([]*domain.Notification, error) {
	return s.notificationRepo.GetByUserID(ctx, userID)
}

// MarkAsRead marks a notification as read by its public UUID.
func (s *Service) MarkAsRead(ctx context.Context, notificationPublicID string) error {
	return s.notificationRepo.MarkAsReadByPublicID(ctx, notificationPublicID)
}
