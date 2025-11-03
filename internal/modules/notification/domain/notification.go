// [OPTIONAL FEATURE: forum-advanced-features]
// Package domain contains core entities for the notification module.
package domain

import "time"

// Notification represents a user notification.
type Notification struct {
	ID        int       // Unique notification identifier
	UserID    int       // ID of the user receiving the notification
	Type      string    // Notification type: "like", "comment", "reply", etc.
	Message   string    // Notification message content
	TargetID  int       // ID of the related entity (post, comment, etc.)
	IsRead    bool      // Whether the notification has been read
	CreatedAt time.Time // Notification creation timestamp
}

// NotificationType constants
const (
	TypeLike    = "like"    // Someone liked user's content
	TypeComment = "comment" // Someone commented on user's post
	TypeReply   = "reply"   // Someone replied to user's comment
)

// MarkAsRead marks the notification as read.
func (n *Notification) MarkAsRead() {
	n.IsRead = true
}
