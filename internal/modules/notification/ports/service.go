// INPUT PORT - Service Interface
// [OPTIONAL FEATURE: forum-advanced-features]
package ports
import ("context"; "forum/internal/modules/notification/domain")
type NotificationService interface {
    CreateNotification(ctx context.Context, userID int, notifType, message string, targetID int) error
    GetUserNotifications(ctx context.Context, userID int) ([]*domain.Notification, error)
    MarkAsRead(ctx context.Context, notificationID int) error
}
