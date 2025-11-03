// OUTPUT PORT - Repository Interface
package ports
import ("context"; "forum/internal/modules/moderation/domain")
type ReportRepository interface {
    Create(ctx context.Context, report *domain.Report) error
    GetByID(ctx context.Context, reportID int) (*domain.Report, error)
    List(ctx context.Context, status string) ([]*domain.Report, error)
    Update(ctx context.Context, report *domain.Report) error
}
