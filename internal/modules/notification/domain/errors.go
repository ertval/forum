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

	// ErrInvalidUserID is returned when the provided user ID is invalid.
	ErrInvalidUserID = errors.New("invalid user id")

	// ErrInvalidTarget is returned when target data is invalid.
	ErrInvalidTarget = errors.New("invalid target")

	// ErrInvalidMessage is returned when the notification message is empty.
	ErrInvalidMessage = errors.New("invalid message")

	// ErrInvalidPublicID is returned when the public identifier is empty or invalid.
	ErrInvalidPublicID = errors.New("invalid public id")
)
