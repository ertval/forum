// INPUT PORT - Service Interface
// [OPTIONAL FEATURE: forum-moderation]
// Package ports defines the input ports (use cases) for the moderation module.
package ports

import (
	"context"
	"forum/internal/modules/moderation/domain"
)

// ModerationService defines moderation management use cases.
type ModerationService interface {
	// CreateReport creates a new moderation report.
	// reporterID: internal user ID from session, targetPublicID: public UUID of reported content
	CreateReport(ctx context.Context, reporterID int, targetPublicID string, targetType, reason string) error

	// ReviewReport marks a report as reviewed with a decision by report's public UUID.
	ReviewReport(ctx context.Context, reportPublicID string, decision string) error

	// ListReports retrieves reports filtered by status.
	ListReports(ctx context.Context, status string) ([]*domain.Report, error)
}
