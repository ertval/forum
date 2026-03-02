// OUTPUT ADAPTER - SQLite Repository
// [OPTIONAL FEATURE: forum-moderation]
// Package adapters implements the SQLite repository for reports.
package adapters

import (
	"context"
	"database/sql"
	"fmt"
	"forum/internal/modules/moderation/domain"
	"forum/internal/modules/moderation/ports"
	"forum/internal/platform/database"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
)

// Report SQL constants.
const reportBaseSelectQuery = `
		SELECT
			r.id,
			r.public_id,
			r.reporter_id,
			r.moderator_id,
			r.target_id,
			r.target_type,
			r.reason,
			r.status,
			COALESCE(r.response, '') AS response,
			r.created_at,
			r.reviewed_at,
			COALESCE(reporter.public_id, '') AS reporter_public_id,
			COALESCE(moderator.public_id, '') AS moderator_public_id,
			CASE
				WHEN r.target_type = 'post' THEN COALESCE(p.public_id, '')
				WHEN r.target_type = 'comment' THEN COALESCE(c.public_id, '')
				ELSE ''
			END AS target_public_id
		FROM reports r
		LEFT JOIN users reporter ON reporter.id = r.reporter_id
		LEFT JOIN users moderator ON moderator.id = r.moderator_id
		LEFT JOIN posts p ON r.target_type = 'post' AND p.id = r.target_id
		LEFT JOIN comments c ON r.target_type = 'comment' AND c.id = r.target_id`

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
	if report.Status == "" {
		report.Status = domain.StatusPending
	}
	if report.CreatedAt.IsZero() {
		report.CreatedAt = time.Now()
	}

	publicID, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("generate report UUID: %w", err)
	}
	report.PublicID = publicID.String()

	result, err := r.db.ExecContext(ctx, `
		INSERT INTO reports (public_id, reporter_id, moderator_id, target_id, target_type, reason, status, response, created_at, reviewed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		report.PublicID,
		report.ReporterID,
		report.ModeratorID,
		report.TargetID,
		report.TargetType,
		report.Reason,
		report.Status,
		nullIfEmpty(report.Response),
		report.CreatedAt,
		report.ReviewedAt,
	)
	if err != nil {
		return fmt.Errorf("insert report: %w", err)
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("read report insert id: %w", err)
	}
	report.ID, err = database.SafeInt64ToInt(lastID)
	if err != nil {
		return fmt.Errorf("last insert id overflow: %w", err)
	}

	return nil
}

// GetByPublicID retrieves a report by its public UUID.
func (r *SQLiteReportRepository) GetByPublicID(ctx context.Context, reportPublicID string) (*domain.Report, error) {
	report, err := r.queryReports(ctx, reportBaseSelectQuery+`
		WHERE r.public_id = ?
	`, reportPublicID)
	if err != nil {
		return nil, err
	}
	if len(report) == 0 {
		return nil, domain.ErrReportNotFound
	}

	return report[0], nil
}

// List retrieves reports filtered by status.
func (r *SQLiteReportRepository) List(ctx context.Context, status string) ([]*domain.Report, error) {
	status = domain.NormalizeStatus(status)

	if status == "" {
		return r.queryReports(ctx, reportBaseSelectQuery+" ORDER BY r.created_at DESC")
	}

	return r.queryReports(ctx, reportBaseSelectQuery+" WHERE r.status = ? ORDER BY r.created_at DESC", status)
}

// Update updates an existing report.
func (r *SQLiteReportRepository) Update(ctx context.Context, report *domain.Report) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE reports
		SET status = ?, response = ?, moderator_id = ?, reviewed_at = ?
		WHERE id = ?
	`, report.Status, nullIfEmpty(report.Response), report.ModeratorID, report.ReviewedAt, report.ID)
	if err != nil {
		return fmt.Errorf("update report: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read updated rows: %w", err)
	}
	if rows == 0 {
		return domain.ErrReportNotFound
	}

	return nil
}

// ResolveTargetID resolves target public UUID to internal INT ID by target type.
func (r *SQLiteReportRepository) ResolveTargetID(ctx context.Context, targetType, targetPublicID string) (int, error) {
	targetType = strings.ToLower(strings.TrimSpace(targetType))
	targetPublicID = strings.TrimSpace(targetPublicID)

	if !domain.IsValidTargetType(targetType) {
		return 0, domain.ErrInvalidTargetType
	}
	if targetPublicID == "" {
		return 0, domain.ErrInvalidTarget
	}

	var query string
	switch targetType {
	case "post":
		query = "SELECT id FROM posts WHERE public_id = ?"
	case "comment":
		query = "SELECT id FROM comments WHERE public_id = ?"
	}

	var targetID int
	err := r.db.QueryRowContext(ctx, query, targetPublicID).Scan(&targetID)
	if err == sql.ErrNoRows {
		return 0, domain.ErrInvalidTarget
	}
	if err != nil {
		return 0, fmt.Errorf("resolve %s target: %w", targetType, err)
	}

	return targetID, nil
}

func (r *SQLiteReportRepository) queryReports(ctx context.Context, query string, args ...any) ([]*domain.Report, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query reports: %w", err)
	}
	defer rows.Close()

	reports := make([]*domain.Report, 0, 16)
	for rows.Next() {
		var report domain.Report
		var moderatorID sql.NullInt64
		var reviewedAt sql.NullTime

		err := rows.Scan(
			&report.ID,
			&report.PublicID,
			&report.ReporterID,
			&moderatorID,
			&report.TargetID,
			&report.TargetType,
			&report.Reason,
			&report.Status,
			&report.Response,
			&report.CreatedAt,
			&reviewedAt,
			&report.PublicReporterID,
			&report.PublicModeratorID,
			&report.PublicTargetID,
		)
		if err != nil {
			return nil, fmt.Errorf("scan report row: %w", err)
		}

		if moderatorID.Valid {
			mid := int(moderatorID.Int64)
			report.ModeratorID = &mid
		}
		if reviewedAt.Valid {
			rt := reviewedAt.Time
			report.ReviewedAt = &rt
		}

		reports = append(reports, &report)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate report rows: %w", err)
	}

	return reports, nil
}

func nullIfEmpty(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
