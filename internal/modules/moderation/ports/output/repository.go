package output
// Package output defines the outbound ports for the moderation module.
package output

import (
	"context"

	"forum/internal/modules/moderation/domain"
)

// ReportRepository defines the interface for report persistence.
type ReportRepository interface {
	Create(ctx context.Context, report *domain.Report) error
	GetByID(ctx context.Context, reportID string) (*domain.Report, error)
	List(ctx context.Context, status domain.ReportStatus, limit, offset int) ([]*domain.Report, error)
	Update(ctx context.Context, report *domain.Report) error
	Delete(ctx context.Context, reportID string) error
}
