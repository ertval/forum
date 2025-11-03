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
func (s *Service) CreateNotification(ctx context.Context, userID int, notifType, message string, targetID int) error {
	// Implementation placeholder
	// 1. Validate notification type
	// 2. Create notification entity
	// 3. Save to repository
	return nil
}

// GetUserNotifications retrieves all notifications for a user.
func (s *Service) GetUserNotifications(ctx context.Context, userID int) ([]*domain.Notification, error) {
	return s.notificationRepo.GetByUserID(ctx, userID)
}

// MarkAsRead marks a notification as read.
func (s *Service) MarkAsRead(ctx context.Context, notificationID int) error {
	return s.notificationRepo.MarkAsRead(ctx, notificationID)
}
