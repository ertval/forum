// OUTPUT ADAPTER - SQLite Repository
package adapters
import ("context"; "database/sql"; "forum/internal/modules/notification/domain"; "forum/internal/modules/notification/ports")
type SQLiteNotificationRepository struct { db *sql.DB }
func NewSQLiteNotificationRepository(db *sql.DB) ports.NotificationRepository {
    return &SQLiteNotificationRepository{db: db}
}
func (r *SQLiteNotificationRepository) Create(ctx context.Context, notification *domain.Notification) error { return nil }
func (r *SQLiteNotificationRepository) GetByUserID(ctx context.Context, userID int) ([]*domain.Notification, error) { return nil, nil }
func (r *SQLiteNotificationRepository) MarkAsRead(ctx context.Context, notificationID int) error { return nil }
