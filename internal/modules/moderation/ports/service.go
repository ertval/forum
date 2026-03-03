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
	CreateReport(ctx context.Context, reporterID int, targetPublicID string, targetType, reason string) (*domain.Report, error)

	// ReviewReport marks a report as reviewed/resolved by report's public UUID.
	ReviewReport(ctx context.Context, moderatorID int, reportPublicID string, status, response string) (*domain.Report, error)

	// ListReports retrieves reports filtered by status.
	ListReports(ctx context.Context, status string) ([]*domain.Report, error)

	// RequestModeratorRole creates a moderator-role request for a user.
	RequestModeratorRole(ctx context.Context, requesterID int, message string) (*domain.ModeratorRequest, error)

	// ReviewModeratorRequest approves or denies a moderator-role request.
	ReviewModeratorRequest(ctx context.Context, reviewerID int, requestPublicID string, status, response string) (*domain.ModeratorRequest, error)

	// GetModeratorRequestByPublicID retrieves a moderator-role request by UUID.
	GetModeratorRequestByPublicID(ctx context.Context, requestPublicID string) (*domain.ModeratorRequest, error)

	// ListModeratorRequests retrieves moderator-role requests filtered by status.
	ListModeratorRequests(ctx context.Context, status string) ([]*domain.ModeratorRequest, error)
}
