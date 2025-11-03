// OUTPUT ADAPTER - SQLite Repository
package adapters
import ("context"; "database/sql"; "forum/internal/modules/moderation/domain"; "forum/internal/modules/moderation/ports")
type SQLiteReportRepository struct { db *sql.DB }
func NewSQLiteReportRepository(db *sql.DB) ports.ReportRepository {
    return &SQLiteReportRepository{db: db}
}
func (r *SQLiteReportRepository) Create(ctx context.Context, report *domain.Report) error { return nil }
func (r *SQLiteReportRepository) GetByID(ctx context.Context, reportID int) (*domain.Report, error) { return nil, nil }
func (r *SQLiteReportRepository) List(ctx context.Context, status string) ([]*domain.Report, error) { return nil, nil }
func (r *SQLiteReportRepository) Update(ctx context.Context, report *domain.Report) error { return nil }
