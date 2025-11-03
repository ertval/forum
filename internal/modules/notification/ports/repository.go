// OUTPUT PORT - Repository Interface
package ports
import ("context"; "forum/internal/modules/notification/domain")
type NotificationRepository interface {
    Create(ctx context.Context, notification *domain.Notification) error
    GetByUserID(ctx context.Context, userID int) ([]*domain.Notification, error)
    MarkAsRead(ctx context.Context, notificationID int) error
}
