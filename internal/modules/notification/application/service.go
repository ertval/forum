// [OPTIONAL FEATURE: forum-advanced-features]
package application
import ("context"; "forum/internal/modules/notification/domain"; "forum/internal/modules/notification/ports")
type Service struct { notificationRepo ports.NotificationRepository }
func NewService(notificationRepo ports.NotificationRepository) *Service {
    return &Service{notificationRepo: notificationRepo}
}
func (s *Service) CreateNotification(ctx context.Context, userID int, notifType, message string, targetID int) error {
    return nil // TODO
}
func (s *Service) GetUserNotifications(ctx context.Context, userID int) ([]*domain.Notification, error) {
    return s.notificationRepo.GetByUserID(ctx, userID)
}
func (s *Service) MarkAsRead(ctx context.Context, notificationID int) error {
    return s.notificationRepo.MarkAsRead(ctx, notificationID)
}
