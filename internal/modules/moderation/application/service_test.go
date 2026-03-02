package application

import (
	"context"
	"errors"
	"fmt"
	"forum/internal/modules/moderation/domain"
	"testing"
	"time"
)

// MockReportRepository implements ReportRepository for testing
type MockReportRepository struct {
	reports  map[int]*domain.Report
	listFn   func(ctx context.Context, status string) ([]*domain.Report, error)
	createFn func(ctx context.Context, report *domain.Report) error
	updateFn func(ctx context.Context, report *domain.Report) error
	getFn    func(ctx context.Context, reportPublicID string) (*domain.Report, error)
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

func (m *MockReportRepository) GetByID(ctx context.Context, reportID int) (*domain.Report, error) {
	// adapt to new repository method: GetByPublicID
	if m.getFn != nil {
		// The test harness will call GetByPublicID directly; keep a shim for compatibility
		return nil, nil
	}

	return nil, domain.ErrReportNotFound
}

// Implement the new interface method expected by ports.ReportRepository
func (m *MockReportRepository) GetByPublicID(ctx context.Context, reportPublicID string) (*domain.Report, error) {
	if m.getFn != nil {
		return m.getFn(ctx, reportPublicID)
	}

	// Try to parse public ID as an int suffix for our simple mock storage
	// expected format in tests: "pub-<id>"
	var id int
	n, _ := fmt.Sscanf(reportPublicID, "pub-%d", &id)
	if n == 1 {
		if report, exists := m.reports[id]; exists {
			return report, nil
		}
	}
	return nil, domain.ErrReportNotFound
}

func TestService_CreateReport(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockReportRepository{}
	service := NewService(mockRepo)

	// Test the current implementation (returns not implemented error)
	err := service.CreateReport(ctx, 1, "pub-10", "post", "Inappropriate content")
	if err == nil {
		t.Error("Expected 'not implemented' error, got nil")
	}
	if !errors.Is(err, domain.ErrNotImplemented) {
		t.Errorf("Expected ErrNotImplemented, got %v", err)
	}
}

func TestService_ReviewReport(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockReportRepository{}
	service := NewService(mockRepo)

	// Test the current implementation (returns not implemented error)
	err := service.ReviewReport(ctx, "pub-1", "resolved")
	if err == nil {
		t.Error("Expected 'not implemented' error, got nil")
	}
	if !errors.Is(err, domain.ErrNotImplemented) {
		t.Errorf("Expected ErrNotImplemented, got %v", err)
	}
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
}
