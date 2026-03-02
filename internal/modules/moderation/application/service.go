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
	return domain.ErrNotImplemented
}

// ReviewReport marks a report as reviewed with a decision.
// TODO: Implement report review logic.
func (s *Service) ReviewReport(ctx context.Context, reportPublicID string, decision string) error {
	return domain.ErrNotImplemented
}

// ListReports retrieves reports filtered by status.
func (s *Service) ListReports(ctx context.Context, status string) ([]*domain.Report, error) {
	return s.reportRepo.List(ctx, status)
}
