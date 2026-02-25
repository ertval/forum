package domain

import (
	"testing"
	"time"
)

func TestNotification_MarkAsRead(t *testing.T) {
	notification := &Notification{
		ID:        1,
		UserID:    5,
		Type:      TypeLike,
		Message:   "Someone liked your post",
		TargetID:  10,
		IsRead:    false,
		CreatedAt: time.Now(),
	}

	if notification.IsRead {
		t.Error("Expected notification to be unread initially")
	}

	notification.MarkAsRead()

	if !notification.IsRead {
		t.Error("Expected notification to be marked as read")
	}
}

func TestNotification_StructFields(t *testing.T) {
	now := time.Now()
	notification := &Notification{
		ID:        1,
		UserID:    5,
		Type:      TypeComment,
		Message:   "Someone commented on your post",
		TargetID:  10,
		IsRead:    false,
		CreatedAt: now,
	}

	if notification.ID != 1 {
		t.Errorf("Expected ID 1, got %d", notification.ID)
	}
	if notification.UserID != 5 {
		t.Errorf("Expected UserID 5, got %d", notification.UserID)
	}
	if notification.Type != TypeComment {
		t.Errorf("Expected Type '%s', got '%s'", TypeComment, notification.Type)
	}
	if notification.Message != "Someone commented on your post" {
		t.Errorf("Expected Message 'Someone commented on your post', got '%s'", notification.Message)
	}
	if notification.TargetID != 10 {
		t.Errorf("Expected TargetID 10, got %d", notification.TargetID)
	}
	if notification.IsRead {
		t.Error("Expected IsRead to be false")
	}
	if !notification.CreatedAt.Equal(now) {
		t.Errorf("Expected CreatedAt %v, got %v", now, notification.CreatedAt)
	}
}

func TestNotificationTypeConstants(t *testing.T) {
	if TypeLike != "like" {
		t.Errorf("Expected TypeLike to be 'like', got '%s'", TypeLike)
	}
	if TypeComment != "comment" {
		t.Errorf("Expected TypeComment to be 'comment', got '%s'", TypeComment)
	}
	if TypeReply != "reply" {
		t.Errorf("Expected TypeReply to be 'reply', got '%s'", TypeReply)
	}
}
