// [OPTIONAL FEATURE: forum-moderation]
// Package application implements the moderation service business logic.
package application

import (
	"context"
	"forum/internal/modules/moderation/domain"
	"forum/internal/modules/moderation/ports"
	"strings"
	"time"
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
func (s *Service) CreateReport(ctx context.Context, reporterID int, targetPublicID string, targetType, reason string) (*domain.Report, error) {
	targetType = strings.ToLower(strings.TrimSpace(targetType))
	reason = strings.TrimSpace(reason)
	targetPublicID = strings.TrimSpace(targetPublicID)

	if reporterID <= 0 {
		return nil, domain.ErrInsufficientPermissions
	}
	if !domain.IsValidTargetType(targetType) {
		return nil, domain.ErrInvalidTargetType
	}
	if reason == "" {
		return nil, domain.ErrInvalidReason
	}
	if targetPublicID == "" {
		return nil, domain.ErrInvalidTarget
	}

	targetID, err := s.reportRepo.ResolveTargetID(ctx, targetType, targetPublicID)
	if err != nil {
		return nil, err
	}

	report := &domain.Report{
		ReporterID:     reporterID,
		TargetID:       targetID,
		TargetType:     targetType,
		Reason:         reason,
		Status:         domain.StatusPending,
		CreatedAt:      time.Now(),
		PublicTargetID: targetPublicID,
	}

	if err := report.Validate(); err != nil {
		return nil, err
	}

	if err := s.reportRepo.Create(ctx, report); err != nil {
		return nil, err
	}

	return report, nil
}

// ReviewReport marks a report as reviewed with a decision.
func (s *Service) ReviewReport(ctx context.Context, moderatorID int, reportPublicID string, status, response string) (*domain.Report, error) {
	if moderatorID <= 0 {
		return nil, domain.ErrInsufficientPermissions
	}

	reportPublicID = strings.TrimSpace(reportPublicID)
	status = domain.NormalizeStatus(status)
	response = strings.TrimSpace(response)

	if reportPublicID == "" {
		return nil, domain.ErrReportNotFound
	}
	if status != domain.StatusReviewed && status != domain.StatusResolved {
		return nil, domain.ErrInvalidReportStatus
	}

	report, err := s.reportRepo.GetByPublicID(ctx, reportPublicID)
	if err != nil {
		return nil, err
	}

	report.Status = status
	report.Response = response
	report.ModeratorID = &moderatorID
	now := time.Now()
	report.ReviewedAt = &now

	if err := s.reportRepo.Update(ctx, report); err != nil {
		return nil, err
	}

	return report, nil
}

// ListReports retrieves reports filtered by status.
func (s *Service) ListReports(ctx context.Context, status string) ([]*domain.Report, error) {
	status = strings.TrimSpace(status)
	if status != "" && !domain.IsValidStatus(status) {
		return nil, domain.ErrInvalidReportStatus
	}
	return s.reportRepo.List(ctx, status)
}

// RequestModeratorRole creates a new moderator-role request for a user.
func (s *Service) RequestModeratorRole(ctx context.Context, requesterID int, message string) (*domain.ModeratorRequest, error) {
	if requesterID <= 0 {
		return nil, domain.ErrInvalidRequester
	}

	pending, err := s.reportRepo.HasPendingModeratorRequest(ctx, requesterID)
	if err != nil {
		return nil, err
	}
	if pending {
		return nil, domain.ErrModeratorRequestAlreadyPending
	}

	request := &domain.ModeratorRequest{
		RequesterID: requesterID,
		Status:      domain.RequestStatusPending,
		Message:     strings.TrimSpace(message),
		CreatedAt:   time.Now(),
	}

	if err := request.Validate(); err != nil {
		return nil, err
	}

	if err := s.reportRepo.CreateModeratorRequest(ctx, request); err != nil {
		return nil, err
	}

	return request, nil
}

// ReviewModeratorRequest approves or denies a moderator-role request.
func (s *Service) ReviewModeratorRequest(ctx context.Context, reviewerID int, requestPublicID string, status, response string) (*domain.ModeratorRequest, error) {
	if reviewerID <= 0 {
		return nil, domain.ErrInsufficientPermissions
	}

	requestPublicID = strings.TrimSpace(requestPublicID)
	status = domain.NormalizeRequestStatus(status)
	response = strings.TrimSpace(response)

	if requestPublicID == "" {
		return nil, domain.ErrModeratorRequestNotFound
	}
	if status != domain.RequestStatusApproved && status != domain.RequestStatusDenied {
		return nil, domain.ErrInvalidRequestStatus
	}

	request, err := s.reportRepo.GetModeratorRequestByPublicID(ctx, requestPublicID)
	if err != nil {
		return nil, err
	}

	request.Status = status
	request.Response = response
	request.ReviewerID = &reviewerID
	now := time.Now()
	request.ReviewedAt = &now

	if err := s.reportRepo.UpdateModeratorRequest(ctx, request); err != nil {
		return nil, err
	}

	return request, nil
}

// GetModeratorRequestByPublicID retrieves a moderator request by UUID.
func (s *Service) GetModeratorRequestByPublicID(ctx context.Context, requestPublicID string) (*domain.ModeratorRequest, error) {
	requestPublicID = strings.TrimSpace(requestPublicID)
	if requestPublicID == "" {
		return nil, domain.ErrModeratorRequestNotFound
	}
	return s.reportRepo.GetModeratorRequestByPublicID(ctx, requestPublicID)
}

// ListModeratorRequests retrieves moderator requests filtered by status.
func (s *Service) ListModeratorRequests(ctx context.Context, status string) ([]*domain.ModeratorRequest, error) {
	status = strings.TrimSpace(status)
	if status != "" && !domain.IsValidRequestStatus(status) {
		return nil, domain.ErrInvalidRequestStatus
	}
	return s.reportRepo.ListModeratorRequests(ctx, status)
}
