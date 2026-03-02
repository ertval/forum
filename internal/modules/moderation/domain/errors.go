// [OPTIONAL FEATURE: forum-moderation]
// Package domain contains error definitions for the moderation module.
package domain

import "errors"

// Domain errors for the moderation module.
var (
	// ErrNotImplemented is returned for optional moderation operations that are not implemented.
	ErrNotImplemented = errors.New("not implemented")

	// ErrReportNotFound is returned when a report doesn't exist.
	ErrReportNotFound = errors.New("report not found")

	// ErrInvalidReportStatus is returned when the report status is invalid.
	ErrInvalidReportStatus = errors.New("invalid report status")

	// ErrInsufficientPermissions is returned when a user lacks moderation permissions.
	ErrInsufficientPermissions = errors.New("insufficient permissions to perform moderation action")
)
