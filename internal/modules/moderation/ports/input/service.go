package input
// Package input defines the inbound ports for the moderation module.
package input

import (
	"context"

	"forum/internal/modules/moderation/domain"
)

// ModerationService defines the moderation use cases.
type ModerationService interface {
	// CreateReport creates a new report.
	CreateReport(ctx context.Context, report *domain.Report) error

	// GetReport retrieves a report by ID.
	GetReport(ctx context.Context, reportID string) (*domain.Report, error)

	// ListReports lists all reports (for moderators/admins).
	ListReports(ctx context.Context, status domain.ReportStatus, limit, offset int) ([]*domain.Report, error)

	// ReviewReport reviews a report and takes action.
	ReviewReport(ctx context.Context, reportID, moderatorID, response string, accept bool) error

	// DeleteContent deletes reported content (moderator action).
	DeleteContent(ctx context.Context, moderatorID, targetID string, targetType domain.ReportTargetType) error
}
