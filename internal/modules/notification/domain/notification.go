package domain
// Package domain contains the core notification business logic.
package domain

import "time"

// NotificationType represents the type of notification.
type NotificationType string

const (
	NotificationPostLiked      NotificationType = "post_liked"
	NotificationPostDisliked   NotificationType = "post_disliked"
	NotificationCommentAdded   NotificationType = "comment_added"
	NotificationCommentLiked   NotificationType = "comment_liked"
	NotificationCommentDisliked NotificationType = "comment_disliked"
)

// Notification represents a user notification.
type Notification struct {
	ID        string
	UserID    string           // User who receives the notification
	ActorID   string           // User who triggered the notification
	TargetID  string           // Post or Comment ID
	Type      NotificationType
	Message   string
	Read      bool
	CreatedAt time.Time
}

// MarkAsRead marks the notification as read.
func (n *Notification) MarkAsRead() {
	n.Read = true
}
