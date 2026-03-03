package application

import (
	"context"
	"errors"
	"forum/internal/modules/moderation/domain"
	"testing"
	"time"
)

// MockReportRepository implements ReportRepository for testing
type MockReportRepository struct {
	reports                      map[int]*domain.Report
	moderatorRequests            map[int]*domain.ModeratorRequest
	listFn                       func(ctx context.Context, status string) ([]*domain.Report, error)
	createFn                     func(ctx context.Context, report *domain.Report) error
	updateFn                     func(ctx context.Context, report *domain.Report) error
	getFn                        func(ctx context.Context, reportPublicID string) (*domain.Report, error)
	resolveFn                    func(ctx context.Context, targetType, targetPublicID string) (int, error)
	createModeratorRequestFn     func(ctx context.Context, request *domain.ModeratorRequest) error
	getModeratorRequestFn        func(ctx context.Context, requestPublicID string) (*domain.ModeratorRequest, error)
	listModeratorRequestsFn      func(ctx context.Context, status string) ([]*domain.ModeratorRequest, error)
	updateModeratorRequestFn     func(ctx context.Context, request *domain.ModeratorRequest) error
	hasPendingModeratorRequestFn func(ctx context.Context, requesterID int) (bool, error)
}

func (m *MockReportRepository) List(ctx context.Context, status string) ([]*domain.Report, error) {
	if m.listFn != nil {
		return m.listFn(ctx, status)
	}

	var result []*domain.Report
	for _, report := range m.reports {
		if status == "" || report.Status == status {
			result = append(result, report)
		}
	}
	return result, nil
}

func (m *MockReportRepository) Create(ctx context.Context, report *domain.Report) error {
	if m.createFn != nil {
		return m.createFn(ctx, report)
	}

	if m.reports == nil {
		m.reports = make(map[int]*domain.Report)
	}
	if report.ID == 0 {
		report.ID = len(m.reports) + 1
	}
	if report.PublicID == "" {
		report.PublicID = "report-public-id"
	}
	m.reports[report.ID] = report
	return nil
}

func (m *MockReportRepository) Update(ctx context.Context, report *domain.Report) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, report)
	}

	if m.reports == nil {
		m.reports = make(map[int]*domain.Report)
	}
	m.reports[report.ID] = report
	return nil
}

func (m *MockReportRepository) GetByPublicID(ctx context.Context, reportPublicID string) (*domain.Report, error) {
	if m.getFn != nil {
		return m.getFn(ctx, reportPublicID)
	}
	for _, report := range m.reports {
		if report.PublicID == reportPublicID {
			return report, nil
		}
	}
	return nil, domain.ErrReportNotFound
}

func (m *MockReportRepository) ResolveTargetID(ctx context.Context, targetType, targetPublicID string) (int, error) {
	if m.resolveFn != nil {
		return m.resolveFn(ctx, targetType, targetPublicID)
	}
	if targetPublicID == "" {
		return 0, domain.ErrInvalidTarget
	}
	return 100, nil
}

func (m *MockReportRepository) CreateModeratorRequest(ctx context.Context, request *domain.ModeratorRequest) error {
	if m.createModeratorRequestFn != nil {
		return m.createModeratorRequestFn(ctx, request)
	}
	if m.moderatorRequests == nil {
		m.moderatorRequests = make(map[int]*domain.ModeratorRequest)
	}
	if request.ID == 0 {
		request.ID = len(m.moderatorRequests) + 1
	}
	if request.PublicID == "" {
		request.PublicID = "moderator-request-public-id"
	}
	m.moderatorRequests[request.ID] = request
	return nil
}

func (m *MockReportRepository) GetModeratorRequestByPublicID(ctx context.Context, requestPublicID string) (*domain.ModeratorRequest, error) {
	if m.getModeratorRequestFn != nil {
		return m.getModeratorRequestFn(ctx, requestPublicID)
	}
	for _, request := range m.moderatorRequests {
		if request.PublicID == requestPublicID {
			return request, nil
		}
	}
	return nil, domain.ErrModeratorRequestNotFound
}

func (m *MockReportRepository) ListModeratorRequests(ctx context.Context, status string) ([]*domain.ModeratorRequest, error) {
	if m.listModeratorRequestsFn != nil {
		return m.listModeratorRequestsFn(ctx, status)
	}
	var result []*domain.ModeratorRequest
	for _, request := range m.moderatorRequests {
		if status == "" || request.Status == status {
			result = append(result, request)
		}
	}
	return result, nil
}

func (m *MockReportRepository) UpdateModeratorRequest(ctx context.Context, request *domain.ModeratorRequest) error {
	if m.updateModeratorRequestFn != nil {
		return m.updateModeratorRequestFn(ctx, request)
	}
	if m.moderatorRequests == nil {
		m.moderatorRequests = make(map[int]*domain.ModeratorRequest)
	}
	m.moderatorRequests[request.ID] = request
	return nil
}

func (m *MockReportRepository) HasPendingModeratorRequest(ctx context.Context, requesterID int) (bool, error) {
	if m.hasPendingModeratorRequestFn != nil {
		return m.hasPendingModeratorRequestFn(ctx, requesterID)
	}
	for _, request := range m.moderatorRequests {
		if request.RequesterID == requesterID && request.Status == domain.RequestStatusPending {
			return true, nil
		}
	}
	return false, nil
}

func TestService_CreateReport(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockReportRepository{reports: map[int]*domain.Report{}}
	service := NewService(mockRepo)

	t.Run("creates report", func(t *testing.T) {
		report, err := service.CreateReport(ctx, 1, "post-public-id", "post", "Inappropriate content")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if report == nil || report.PublicID == "" {
			t.Fatalf("expected report with public id, got %#v", report)
		}
		if report.Status != domain.StatusPending {
			t.Fatalf("status = %q, want %q", report.Status, domain.StatusPending)
		}
		if report.TargetID != 100 {
			t.Fatalf("target id = %d, want 100", report.TargetID)
		}
	})

	t.Run("invalid target type", func(t *testing.T) {
		_, err := service.CreateReport(ctx, 1, "post-public-id", "invalid", "reason")
		if !errors.Is(err, domain.ErrInvalidTargetType) {
			t.Fatalf("expected ErrInvalidTargetType, got %v", err)
		}
	})

	t.Run("empty reason", func(t *testing.T) {
		_, err := service.CreateReport(ctx, 1, "post-public-id", "post", "   ")
		if !errors.Is(err, domain.ErrInvalidReason) {
			t.Fatalf("expected ErrInvalidReason, got %v", err)
		}
	})

	t.Run("invalid target", func(t *testing.T) {
		mockRepo.resolveFn = func(ctx context.Context, targetType, targetPublicID string) (int, error) {
			return 0, domain.ErrInvalidTarget
		}
		_, err := service.CreateReport(ctx, 1, "unknown", "post", "reason")
		if !errors.Is(err, domain.ErrInvalidTarget) {
			t.Fatalf("expected ErrInvalidTarget, got %v", err)
		}
	})
}

func TestService_ReviewReport(t *testing.T) {
	ctx := context.Background()

	t.Run("reviews report", func(t *testing.T) {
		now := time.Now()
		report := &domain.Report{
			ID:        1,
			PublicID:  "report-public-id",
			Status:    domain.StatusPending,
			CreatedAt: now,
		}
		mockRepo := &MockReportRepository{
			reports: map[int]*domain.Report{1: report},
			getFn: func(ctx context.Context, reportPublicID string) (*domain.Report, error) {
				if reportPublicID != "report-public-id" {
					return nil, domain.ErrReportNotFound
				}
				return report, nil
			},
		}
		service := NewService(mockRepo)

		updated, err := service.ReviewReport(ctx, 42, "report-public-id", domain.StatusResolved, "Handled")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if updated.Status != domain.StatusResolved {
			t.Fatalf("status = %q, want %q", updated.Status, domain.StatusResolved)
		}
		if updated.ModeratorID == nil || *updated.ModeratorID != 42 {
			t.Fatalf("moderator id = %v, want 42", updated.ModeratorID)
		}
		if updated.ReviewedAt == nil {
			t.Fatal("expected reviewed_at to be set")
		}
	})

	t.Run("invalid review status", func(t *testing.T) {
		mockRepo := &MockReportRepository{}
		service := NewService(mockRepo)
		_, err := service.ReviewReport(ctx, 1, "report-public-id", domain.StatusPending, "")
		if !errors.Is(err, domain.ErrInvalidReportStatus) {
			t.Fatalf("expected ErrInvalidReportStatus, got %v", err)
		}
	})

	t.Run("report not found", func(t *testing.T) {
		mockRepo := &MockReportRepository{}
		service := NewService(mockRepo)
		_, err := service.ReviewReport(ctx, 1, "missing", domain.StatusReviewed, "")
		if !errors.Is(err, domain.ErrReportNotFound) {
			t.Fatalf("expected ErrReportNotFound, got %v", err)
		}
	})
}

func TestService_ListReports(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockReportRepository{}
	service := NewService(mockRepo)

	// Add test reports to the mock
	now := time.Now()
	reports := []*domain.Report{
		{ID: 1, ReporterID: 1, TargetID: 10, TargetType: "post", Reason: "Spam", Status: domain.StatusPending, CreatedAt: now},
		{ID: 2, ReporterID: 2, TargetID: 5, TargetType: "comment", Reason: "Inappropriate", Status: domain.StatusReviewed, CreatedAt: now},
		{ID: 3, ReporterID: 3, TargetID: 15, TargetType: "post", Reason: "Off-topic", Status: domain.StatusPending, CreatedAt: now},
	}
	mockRepo.reports = map[int]*domain.Report{}
	for _, report := range reports {
		mockRepo.reports[report.ID] = report
	}

	t.Run("list all reports", func(t *testing.T) {
		result, err := service.ListReports(ctx, "")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(result) != 3 {
			t.Errorf("Expected 3 reports, got %d", len(result))
		}
	})

	t.Run("list pending reports", func(t *testing.T) {
		result, err := service.ListReports(ctx, domain.StatusPending)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(result) != 2 {
			t.Errorf("Expected 2 pending reports, got %d", len(result))
		}

		// Verify all returned reports have the correct status
		for _, report := range result {
			if report.Status != domain.StatusPending {
				t.Errorf("Expected Status %s, got %s", domain.StatusPending, report.Status)
			}
		}
	})

	t.Run("list reviewed reports", func(t *testing.T) {
		result, err := service.ListReports(ctx, domain.StatusReviewed)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(result) != 1 {
			t.Errorf("Expected 1 reviewed report, got %d", len(result))
		}

		// Verify the returned report has the correct status
		if len(result) > 0 && result[0].Status != domain.StatusReviewed {
			t.Errorf("Expected Status %s, got %s", domain.StatusReviewed, result[0].Status)
		}
	})

	t.Run("invalid status filter", func(t *testing.T) {
		_, err := service.ListReports(ctx, "nope")
		if !errors.Is(err, domain.ErrInvalidReportStatus) {
			t.Fatalf("expected ErrInvalidReportStatus, got %v", err)
		}
	})
}

func TestService_RequestModeratorRole(t *testing.T) {
	ctx := context.Background()
	service := NewService(&MockReportRepository{moderatorRequests: map[int]*domain.ModeratorRequest{}})

	request, err := service.RequestModeratorRole(ctx, 12, "I can help moderate")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if request.Status != domain.RequestStatusPending {
		t.Fatalf("status = %q, want %q", request.Status, domain.RequestStatusPending)
	}

	_, err = service.RequestModeratorRole(ctx, 12, "duplicate")
	if !errors.Is(err, domain.ErrModeratorRequestAlreadyPending) {
		t.Fatalf("expected ErrModeratorRequestAlreadyPending, got %v", err)
	}
}

func TestService_ReviewModeratorRequest(t *testing.T) {
	ctx := context.Background()
	req := &domain.ModeratorRequest{ID: 1, PublicID: "req-1", RequesterID: 10, Status: domain.RequestStatusPending, CreatedAt: time.Now()}
	repo := &MockReportRepository{moderatorRequests: map[int]*domain.ModeratorRequest{1: req}}
	service := NewService(repo)

	updated, err := service.ReviewModeratorRequest(ctx, 99, "req-1", domain.RequestStatusApproved, "ok")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Status != domain.RequestStatusApproved {
		t.Fatalf("status = %q, want %q", updated.Status, domain.RequestStatusApproved)
	}
	if updated.ReviewerID == nil || *updated.ReviewerID != 99 {
		t.Fatalf("reviewer id = %v, want 99", updated.ReviewerID)
	}
}
