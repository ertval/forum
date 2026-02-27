// OUTPUT PORT - Repository Interface
// [OPTIONAL FEATURE: forum-moderation]
// Package ports defines the output ports (data access contracts) for the moderation module.
package ports

import (
	"context"
	"forum/internal/modules/moderation/domain"
)

// ReportRepository defines the data access contract for reports.
type ReportRepository interface {
	// Create stores a new report in the repository.
	// Must generate and set PublicID (UUID) before persisting.
	Create(ctx context.Context, report *domain.Report) error

	// GetByPublicID retrieves a report by its public UUID.
	GetByPublicID(ctx context.Context, reportPublicID string) (*domain.Report, error)

	// List retrieves reports filtered by status.
	List(ctx context.Context, status string) ([]*domain.Report, error)

	// Update updates an existing report in the repository.
	// Uses internal ID from the report entity.
	Update(ctx context.Context, report *domain.Report) error
}
