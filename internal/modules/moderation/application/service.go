// [OPTIONAL FEATURE: forum-moderation]
package application
import ("context"; "forum/internal/modules/moderation/domain"; "forum/internal/modules/moderation/ports")
type Service struct { reportRepo ports.ReportRepository }
func NewService(reportRepo ports.ReportRepository) *Service {
    return &Service{reportRepo: reportRepo}
}
func (s *Service) CreateReport(ctx context.Context, reporterID, targetID int, targetType, reason string) error {
    return nil // TODO
}
func (s *Service) ReviewReport(ctx context.Context, reportID int, decision string) error {
    return nil // TODO
}
func (s *Service) ListReports(ctx context.Context, status string) ([]*domain.Report, error) {
    return s.reportRepo.List(ctx, status)
}
