// [OPTIONAL FEATURE: forum-advanced-features]
// Package domain contains error definitions for the notification module.
package domain

import "errors"

// Domain errors for the notification module.
var (
	// ErrNotificationNotFound is returned when a notification doesn't exist.
	ErrNotificationNotFound = errors.New("notification not found")

	// ErrInvalidNotificationType is returned when the notification type is invalid.
	ErrInvalidNotificationType = errors.New("invalid notification type")
)
