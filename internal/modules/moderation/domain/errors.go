// [OPTIONAL FEATURE: forum-moderation]
// Package domain contains error definitions for the moderation module.
package domain

import "errors"

// Domain errors for the moderation module.
var (
	// ErrReportNotFound is returned when a report doesn't exist.
	ErrReportNotFound = errors.New("report not found")

	// ErrInvalidReportStatus is returned when the report status is invalid.
	ErrInvalidReportStatus = errors.New("invalid report status")

	// ErrInvalidTargetType is returned when the report target type is invalid.
	ErrInvalidTargetType = errors.New("invalid target type")

	// ErrInvalidReason is returned when the report reason is invalid.
	ErrInvalidReason = errors.New("invalid report reason")

	// ErrInvalidTarget is returned when target content cannot be found.
	ErrInvalidTarget = errors.New("invalid report target")

	// ErrInsufficientPermissions is returned when a user lacks moderation permissions.
	ErrInsufficientPermissions = errors.New("insufficient permissions to perform moderation action")
)
