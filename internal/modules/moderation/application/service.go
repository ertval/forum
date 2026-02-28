// [OPTIONAL FEATURE: forum-moderation]
// Package application implements moderation service business logic.
package application

import (
	"context"
	"forum/internal/modules/moderation/domain"
	"forum/internal/modules/moderation/ports"
)

// Service implements the ModerationService interface.
type Service struct {
	reportRepo ports.ReportRepository
}

// NewService creates a new moderation service.
func NewService(reportRepo ports.ReportRepository) *Service {
	return &Service{reportRepo: reportRepo}
}

// CreateReport creates a new moderation report.
// TODO: Implement report creation with validation.
func (s *Service) CreateReport(ctx context.Context, reporterID int, targetPublicID string, targetType, reason string) error {
	// Implementation placeholder
	// 1. Validate target type and reason
	// 2. Resolve targetPublicID to internal target ID
	// 3. Create report entity with "pending" status
	// 4. Save to repository (repo generates PublicID)
	return nil
}

// ReviewReport marks a report as reviewed with a decision.
// TODO: Implement report review logic.
func (s *Service) ReviewReport(ctx context.Context, reportPublicID string, decision string) error {
	// Implementation placeholder
	// 1. Retrieve report by public ID
	// 2. Validate decision (resolved/dismissed)
	// 3. Update report status
	// 4. Save to repository
	return nil
}

// ListReports retrieves reports filtered by status.
func (s *Service) ListReports(ctx context.Context, status string) ([]*domain.Report, error) {
	return s.reportRepo.List(ctx, status)
}
