// OUTPUT ADAPTER - SQLite Repository
// [OPTIONAL FEATURE: forum-moderation]
// Package adapters implements the SQLite repository for reports.
package adapters

import (
	"context"
	"database/sql"
	"errors"
	"forum/internal/modules/moderation/domain"
	"forum/internal/modules/moderation/ports"
)

// SQLiteReportRepository implements the ReportRepository interface using SQLite.
type SQLiteReportRepository struct {
	db *sql.DB
}

// NewSQLiteReportRepository creates a new SQLite report repository.
func NewSQLiteReportRepository(db *sql.DB) ports.ReportRepository {
	return &SQLiteReportRepository{db: db}
}

// Create stores a new report in the database.
func (r *SQLiteReportRepository) Create(ctx context.Context, report *domain.Report) error {
	return errors.New("not implemented")
}

// GetByPublicID retrieves a report by its public UUID.
func (r *SQLiteReportRepository) GetByPublicID(ctx context.Context, reportPublicID string) (*domain.Report, error) {
	return nil, errors.New("not implemented")
}

// List retrieves reports filtered by status.
func (r *SQLiteReportRepository) List(ctx context.Context, status string) ([]*domain.Report, error) {
	return nil, errors.New("not implemented")
}

// Update updates an existing report.
func (r *SQLiteReportRepository) Update(ctx context.Context, report *domain.Report) error {
	return errors.New("not implemented")
}
