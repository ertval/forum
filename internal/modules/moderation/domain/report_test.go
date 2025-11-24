package domain

import (
	"testing"
	"time"
)

func TestReport_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		report   *Report
		expected bool
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
			expected: true,
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
			expected: true,
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
			expected: false,
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
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.report.IsValid()
			if result != tt.expected {
				t.Errorf("IsValid() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestReport_StructFields(t *testing.T) {
	now := time.Now()
	report := &Report{
		ID:         1,
		ReporterID: 10,
		TargetID:   5,
		TargetType: "post",
		Reason:     "Inappropriate content",
		Status:     StatusPending,
		CreatedAt:  now,
	}

	if report.ID != 1 {
		t.Errorf("Expected ID 1, got %d", report.ID)
	}
	if report.ReporterID != 10 {
		t.Errorf("Expected ReporterID 10, got %d", report.ReporterID)
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