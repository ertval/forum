// OUTPUT ADAPTER - SQLite Repository
// [OPTIONAL FEATURE: forum-moderation]
// Package adapters implements the SQLite repository for reports.
package adapters

import (
	"context"
	"database/sql"
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
// TODO: Implement report creation with UUID generation.
func (r *SQLiteReportRepository) Create(ctx context.Context, report *domain.Report) error {
	// Implementation placeholder
	// 1. Generate UUID for PublicID
	// 2. INSERT INTO reports (public_id, reporter_id, target_id, target_type, reason, status, created_at)
	//    VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	return nil
}

// GetByPublicID retrieves a report by its public UUID.
// TODO: Implement report retrieval by public UUID.
func (r *SQLiteReportRepository) GetByPublicID(ctx context.Context, reportPublicID string) (*domain.Report, error) {
	// Implementation placeholder
	// SELECT id, public_id, reporter_id, target_id, target_type, reason, status, created_at
	// FROM reports WHERE public_id = ?
	return nil, nil
}

// List retrieves reports filtered by status.
// TODO: Implement report listing with optional status filter.
func (r *SQLiteReportRepository) List(ctx context.Context, status string) ([]*domain.Report, error) {
	// Implementation placeholder
	// SELECT id, public_id, reporter_id, target_id, target_type, reason, status, created_at
	// FROM reports WHERE status = ? OR ? = '' ORDER BY created_at DESC
	return nil, nil
}

// Update updates an existing report.
// TODO: Implement report update.
func (r *SQLiteReportRepository) Update(ctx context.Context, report *domain.Report) error {
	// Implementation placeholder
	// UPDATE reports SET status = ? WHERE id = ?
	return nil
}
