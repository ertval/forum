// OUTPUT ADAPTER - SQLite Repository
// [OPTIONAL FEATURE: forum-advanced-features]
// Package adapters implements the SQLite repository for notifications.
package adapters

import (
	"context"
	"database/sql"
	"fmt"
	"forum/internal/modules/notification/domain"
	"forum/internal/modules/notification/ports"

	"github.com/gofrs/uuid/v5"
)

// SQLiteNotificationRepository implements the NotificationRepository interface using SQLite.
type SQLiteNotificationRepository struct {
	db *sql.DB
}

// NewSQLiteNotificationRepository creates a new SQLite notification repository.
func NewSQLiteNotificationRepository(db *sql.DB) ports.NotificationRepository {
	return &SQLiteNotificationRepository{db: db}
}

// Create stores a new notification in the database.
func (r *SQLiteNotificationRepository) Create(ctx context.Context, notification *domain.Notification) error {
	publicID, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("generate notification UUID: %w", err)
	}
	notification.PublicID = publicID.String()

	var targetID int
	err = r.db.QueryRowContext(ctx, "SELECT id FROM posts WHERE public_id = ?", notification.PublicTargetID).Scan(&targetID)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.ErrInvalidTarget
		}
		return fmt.Errorf("resolve notification target: %w", err)
	}
	notification.TargetID = targetID

	createdAt := notification.CreatedAt
	if createdAt.IsZero() {
		// Keep DB-driven timestamp behavior if not set by service.
		createdAt = sql.NullTime{}.Time
	}

	if createdAt.IsZero() {
		_, err = r.db.ExecContext(ctx, `
			INSERT INTO notifications (public_id, user_id, actor_id, target_id, type, message, read, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		`, notification.PublicID, notification.UserID, notification.ActorID, notification.TargetID, notification.Type, notification.Message, notification.IsRead)
		if err != nil {
			return fmt.Errorf("insert notification: %w", err)
		}
		return nil
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO notifications (public_id, user_id, actor_id, target_id, type, message, read, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, notification.PublicID, notification.UserID, notification.ActorID, notification.TargetID, notification.Type, notification.Message, notification.IsRead, createdAt)
	if err != nil {
		return fmt.Errorf("insert notification: %w", err)
	}

	return nil
}

// GetByUserID retrieves all notifications for a user.
func (r *SQLiteNotificationRepository) GetByUserID(ctx context.Context, userID int) ([]*domain.Notification, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT n.id, n.public_id, n.user_id, n.actor_id, n.type, n.message, n.target_id, n.read, n.created_at,
		       p.public_id as target_public_id, u.public_id as actor_public_id
		FROM notifications n
		LEFT JOIN posts p ON n.target_id = p.id
		LEFT JOIN users u ON n.actor_id = u.id
		WHERE n.user_id = ?
		ORDER BY n.created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("query notifications: %w", err)
	}
	defer rows.Close()

	notifications := make([]*domain.Notification, 0)
	for rows.Next() {
		var notification domain.Notification
		var targetPublicID sql.NullString
		var actorPublicID sql.NullString

		if err := rows.Scan(
			&notification.ID,
			&notification.PublicID,
			&notification.UserID,
			&notification.ActorID,
			&notification.Type,
			&notification.Message,
			&notification.TargetID,
			&notification.IsRead,
			&notification.CreatedAt,
			&targetPublicID,
			&actorPublicID,
		); err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}

		if targetPublicID.Valid {
			notification.PublicTargetID = targetPublicID.String
		}
		if actorPublicID.Valid {
			notification.PublicActorID = actorPublicID.String
		}

		notifications = append(notifications, &notification)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notifications: %w", err)
	}

	return notifications, nil
}

// MarkAsReadByPublicID marks a notification as read by its public UUID.
func (r *SQLiteNotificationRepository) MarkAsReadByPublicID(ctx context.Context, notificationPublicID string) error {
	result, err := r.db.ExecContext(ctx, `UPDATE notifications SET read = 1 WHERE public_id = ?`, notificationPublicID)
	if err != nil {
		return fmt.Errorf("mark notification as read: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected for mark as read: %w", err)
	}
	if rows == 0 {
		return domain.ErrNotificationNotFound
	}

	return nil
}

// MarkAllAsReadByUserID marks all notifications as read for a user.
func (r *SQLiteNotificationRepository) MarkAllAsReadByUserID(ctx context.Context, userID int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE notifications SET read = 1 WHERE user_id = ? AND read = 0`, userID)
	if err != nil {
		return fmt.Errorf("mark all notifications as read: %w", err)
	}
	return nil
}
