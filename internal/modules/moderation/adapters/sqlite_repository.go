// OUTPUT ADAPTER - SQLite Repository
// [OPTIONAL FEATURE: forum-moderation]
// Package adapters implements the SQLite repository for reports.
package adapters

import (
	"context"
	"database/sql"
	"errors"
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

const moderatorRequestBaseSelectQuery = `
		SELECT
			mr.id,
			mr.public_id,
			mr.requester_id,
			mr.reviewer_id,
			mr.status,
			COALESCE(mr.message, '') AS message,
			COALESCE(mr.response, '') AS response,
			mr.created_at,
			mr.reviewed_at,
			COALESCE(requester.public_id, '') AS requester_public_id,
			COALESCE(reviewer.public_id, '') AS reviewer_public_id
		FROM moderator_requests mr
		LEFT JOIN users requester ON requester.id = mr.requester_id
		LEFT JOIN users reviewer ON reviewer.id = mr.reviewer_id`

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

// CreateModeratorRequest stores a new moderator request in the database.
func (r *SQLiteReportRepository) CreateModeratorRequest(ctx context.Context, request *domain.ModeratorRequest) error {
	if request.Status == "" {
		request.Status = domain.RequestStatusPending
	}
	if request.CreatedAt.IsZero() {
		request.CreatedAt = time.Now()
	}

	publicID, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("generate moderator request UUID: %w", err)
	}
	request.PublicID = publicID.String()

	result, err := r.db.ExecContext(ctx, `
		INSERT INTO moderator_requests (public_id, requester_id, reviewer_id, status, message, response, created_at, reviewed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		request.PublicID,
		request.RequesterID,
		request.ReviewerID,
		request.Status,
		nullIfEmpty(request.Message),
		nullIfEmpty(request.Response),
		request.CreatedAt,
		request.ReviewedAt,
	)
	if err != nil {
		return fmt.Errorf("insert moderator request: %w", err)
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("read moderator request insert id: %w", err)
	}
	request.ID, err = database.SafeInt64ToInt(lastID)
	if err != nil {
		return fmt.Errorf("last insert id overflow: %w", err)
	}

	return nil
}

// GetModeratorRequestByPublicID retrieves a moderator request by UUID.
func (r *SQLiteReportRepository) GetModeratorRequestByPublicID(ctx context.Context, requestPublicID string) (*domain.ModeratorRequest, error) {
	requests, err := r.queryModeratorRequests(ctx, moderatorRequestBaseSelectQuery+`
		WHERE mr.public_id = ?
	`, requestPublicID)
	if err != nil {
		return nil, err
	}
	if len(requests) == 0 {
		return nil, domain.ErrModeratorRequestNotFound
	}

	return requests[0], nil
}

// ListModeratorRequests retrieves moderator requests filtered by status.
func (r *SQLiteReportRepository) ListModeratorRequests(ctx context.Context, status string) ([]*domain.ModeratorRequest, error) {
	status = domain.NormalizeRequestStatus(status)

	if status == "" {
		return r.queryModeratorRequests(ctx, moderatorRequestBaseSelectQuery+" ORDER BY mr.created_at DESC")
	}

	return r.queryModeratorRequests(ctx, moderatorRequestBaseSelectQuery+" WHERE mr.status = ? ORDER BY mr.created_at DESC", status)
}

// UpdateModeratorRequest updates an existing moderator request.
func (r *SQLiteReportRepository) UpdateModeratorRequest(ctx context.Context, request *domain.ModeratorRequest) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE moderator_requests
		SET status = ?, response = ?, reviewer_id = ?, reviewed_at = ?
		WHERE id = ?
	`, request.Status, nullIfEmpty(request.Response), request.ReviewerID, request.ReviewedAt, request.ID)
	if err != nil {
		return fmt.Errorf("update moderator request: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read updated rows: %w", err)
	}
	if rows == 0 {
		return domain.ErrModeratorRequestNotFound
	}

	return nil
}

// HasPendingModeratorRequest checks if requester has an existing pending moderator request.
func (r *SQLiteReportRepository) HasPendingModeratorRequest(ctx context.Context, requesterID int) (bool, error) {
	var exists int
	err := r.db.QueryRowContext(ctx, `
		SELECT 1
		FROM moderator_requests
		WHERE requester_id = ? AND status = ?
		LIMIT 1
	`, requesterID, domain.RequestStatusPending).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check pending moderator request: %w", err)
	}
	return true, nil
}

func (r *SQLiteReportRepository) queryModeratorRequests(ctx context.Context, query string, args ...any) ([]*domain.ModeratorRequest, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query moderator requests: %w", err)
	}
	defer rows.Close()

	requests := make([]*domain.ModeratorRequest, 0, 16)
	for rows.Next() {
		var request domain.ModeratorRequest
		var reviewerID sql.NullInt64
		var reviewedAt sql.NullTime

		err := rows.Scan(
			&request.ID,
			&request.PublicID,
			&request.RequesterID,
			&reviewerID,
			&request.Status,
			&request.Message,
			&request.Response,
			&request.CreatedAt,
			&reviewedAt,
			&request.PublicRequesterID,
			&request.PublicReviewerID,
		)
		if err != nil {
			return nil, fmt.Errorf("scan moderator request row: %w", err)
		}

		if reviewerID.Valid {
			rid := int(reviewerID.Int64)
			request.ReviewerID = &rid
		}
		if reviewedAt.Valid {
			rt := reviewedAt.Time
			request.ReviewedAt = &rt
		}

		requests = append(requests, &request)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate moderator request rows: %w", err)
	}

	return requests, nil
}

func nullIfEmpty(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
