// [OPTIONAL FEATURE: forum-advanced-features]
// Package domain contains the core business entities for the notification module.
package domain

import "time"

// Notification represents a user notification.
type Notification struct {
	ID        int       `json:"-"`          // Internal unique identifier (INT PRIMARY KEY)
	PublicID  string    `json:"id"`         // Public UUID identifier (exposed in API)
	UserID    int       `json:"-"`          // Internal ID of the user receiving the notification
	ActorID   int       `json:"-"`          // Internal ID of the user triggering the notification
	Type      string    `json:"type"`       // Notification type: "like", "comment", "reply", etc.
	Message   string    `json:"message"`    // Notification message content
	TargetID  int       `json:"-"`          // Internal ID of the related entity (post, comment, etc.)
	IsRead    bool      `json:"is_read"`    // Whether the notification has been read
	CreatedAt time.Time `json:"created_at"` // Notification creation timestamp
	// For API responses - public UUID of related entity
	PublicTargetID string `json:"target_id,omitempty"` // Public UUID of related entity
	PublicActorID  string `json:"actor_id,omitempty"`  // Public UUID of actor user
}

// NotificationType constants
const (
	TypeLike    = "like"    // Someone liked user's content
	TypeDislike = "dislike" // Someone disliked user's content
	TypeComment = "comment" // Someone commented on user's post
	TypeReply   = "reply"   // Someone replied to user's comment
)

// Validate checks that the notification has the required fields set.
func (n *Notification) Validate() error {
	if n.PublicID == "" {
		return ErrInvalidPublicID
	}
	if n.UserID <= 0 {
		return ErrInvalidUserID
	}
	if n.Message == "" {
		return ErrInvalidMessage
	}
	if n.Type != TypeLike && n.Type != TypeDislike && n.Type != TypeComment && n.Type != TypeReply {
		return ErrInvalidNotificationType
	}
	return nil
}

// MarkAsRead marks the notification as read.
func (n *Notification) MarkAsRead() {
	n.IsRead = true
}
