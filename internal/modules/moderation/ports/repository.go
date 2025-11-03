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
	Create(ctx context.Context, report *domain.Report) error

	// GetByID retrieves a report by its ID.
	GetByID(ctx context.Context, reportID int) (*domain.Report, error)

	// List retrieves reports filtered by status.
	List(ctx context.Context, status string) ([]*domain.Report, error)

	// Update updates an existing report in the repository.
	Update(ctx context.Context, report *domain.Report) error
}
