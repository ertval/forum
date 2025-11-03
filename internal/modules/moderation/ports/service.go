// INPUT PORT - Service Interface
// [OPTIONAL FEATURE: forum-moderation]
package ports
import ("context"; "forum/internal/modules/moderation/domain")
type ModerationService interface {
    CreateReport(ctx context.Context, reporterID, targetID int, targetType, reason string) error
    ReviewReport(ctx context.Context, reportID int, decision string) error
    ListReports(ctx context.Context, status string) ([]*domain.Report, error)
}
