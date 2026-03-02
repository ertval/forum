package domain

import (
	"testing"
	"time"
)

func TestReport_Validate(t *testing.T) {
	tests := []struct {
		name      string
		report    *Report
		expectErr bool
	}{
		{
			name: "valid post report",
			report: &Report{
				ID:         1,
				ReporterID: 1,
				TargetID:   10,
				TargetType: "post",
				Reason:     "Inappropriate content",
				Status:     StatusPending,
				CreatedAt:  time.Now(),
			},
			expectErr: false,
		},
		{
			name: "valid comment report",
			report: &Report{
				ID:         1,
				ReporterID: 1,
				TargetID:   5,
				TargetType: "comment",
				Reason:     "Spam",
				Status:     StatusPending,
				CreatedAt:  time.Now(),
			},
			expectErr: false,
		},
		{
			name: "invalid target type",
			report: &Report{
				ID:         1,
				ReporterID: 1,
				TargetID:   10,
				TargetType: "invalid",
				Reason:     "Inappropriate content",
				Status:     StatusPending,
				CreatedAt:  time.Now(),
			},
			expectErr: true,
		},
		{
			name: "empty target type",
			report: &Report{
				ID:         1,
				ReporterID: 1,
				TargetID:   10,
				TargetType: "",
				Reason:     "Inappropriate content",
				Status:     StatusPending,
				CreatedAt:  time.Now(),
			},
			expectErr: true,
		},
		{
			name: "empty reason",
			report: &Report{
				ID:         1,
				ReporterID: 1,
				TargetID:   10,
				TargetType: "post",
				Reason:     "   ",
				Status:     StatusPending,
				CreatedAt:  time.Now(),
			},
			expectErr: true,
		},
		{
			name: "invalid status",
			report: &Report{
				ID:         1,
				ReporterID: 1,
				TargetID:   10,
				TargetType: "post",
				Reason:     "Inappropriate content",
				Status:     "unknown",
				CreatedAt:  time.Now(),
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.report.Validate()
			if (err != nil) != tt.expectErr {
				t.Errorf("Validate() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestReport_StructFields(t *testing.T) {
	now := time.Now()
	report := &Report{
		ID:         1,
		ReporterID: 10,
		PublicID:   "report-public-id",
		TargetID:   5,
		TargetType: "post",
		Reason:     "Inappropriate content",
		Status:     StatusPending,
		Response:   "",
		CreatedAt:  now,
	}

	if report.ID != 1 {
		t.Errorf("Expected ID 1, got %d", report.ID)
	}
	if report.ReporterID != 10 {
		t.Errorf("Expected ReporterID 10, got %d", report.ReporterID)
	}
	if report.PublicID != "report-public-id" {
		t.Errorf("Expected PublicID 'report-public-id', got '%s'", report.PublicID)
	}
	if report.TargetID != 5 {
		t.Errorf("Expected TargetID 5, got %d", report.TargetID)
	}
	if report.TargetType != "post" {
		t.Errorf("Expected TargetType 'post', got '%s'", report.TargetType)
	}
	if report.Reason != "Inappropriate content" {
		t.Errorf("Expected Reason 'Inappropriate content', got '%s'", report.Reason)
	}
	if report.Status != StatusPending {
		t.Errorf("Expected Status '%s', got '%s'", StatusPending, report.Status)
	}
	if report.Response != "" {
		t.Errorf("Expected Response '', got '%s'", report.Response)
	}
	if !report.CreatedAt.Equal(now) {
		t.Errorf("Expected CreatedAt %v, got %v", now, report.CreatedAt)
	}
}

func TestReportStatusConstants(t *testing.T) {
	if StatusPending != "pending" {
		t.Errorf("Expected StatusPending to be 'pending', got '%s'", StatusPending)
	}
	if StatusReviewed != "reviewed" {
		t.Errorf("Expected StatusReviewed to be 'reviewed', got '%s'", StatusReviewed)
	}
	if StatusResolved != "resolved" {
		t.Errorf("Expected StatusResolved to be 'resolved', got '%s'", StatusResolved)
	}
}

func TestHelpers(t *testing.T) {
	if !IsValidStatus(StatusPending) {
		t.Fatal("expected pending status to be valid")
	}
	if IsValidStatus("unknown") {
		t.Fatal("expected unknown status to be invalid")
	}
	if !IsValidTargetType("post") || !IsValidTargetType("comment") {
		t.Fatal("expected post/comment target types to be valid")
	}
	if IsValidTargetType("user") {
		t.Fatal("expected user target type to be invalid")
	}
	if NormalizeStatus("  ReSolVed ") != StatusResolved {
		t.Fatalf("expected normalized status to be %q", StatusResolved)
	}
}
