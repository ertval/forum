// [OPTIONAL FEATURE: forum-moderation]
// Package domain contains the core business entities for the moderation module.
package domain

import (
	"strings"
	"time"
)

// Report represents a moderation report for inappropriate content.
type Report struct {
	ID          int        `json:"-"`                     // Internal unique identifier (INT PRIMARY KEY)
	PublicID    string     `json:"id"`                    // Public UUID identifier (exposed in API)
	ReporterID  int        `json:"-"`                     // Internal ID of the user who created the report
	ModeratorID *int       `json:"-"`                     // Internal ID of moderator who reviewed the report
	TargetID    int        `json:"-"`                     // Internal ID of the reported content (post or comment)
	TargetType  string     `json:"target_type"`           // Type of target: "post" or "comment"
	Reason      string     `json:"reason"`                // Reason for the report
	Status      string     `json:"status"`                // Report status: "pending", "reviewed", "resolved"
	Response    string     `json:"response,omitempty"`    // Moderator response to report
	CreatedAt   time.Time  `json:"created_at"`            // Report creation timestamp
	ReviewedAt  *time.Time `json:"reviewed_at,omitempty"` // Report review timestamp
	// For API responses - public UUIDs of related entities
	PublicReporterID  string `json:"reporter_id,omitempty"`  // Public UUID of reporter
	PublicModeratorID string `json:"moderator_id,omitempty"` // Public UUID of reviewer
	PublicTargetID    string `json:"target_id,omitempty"`    // Public UUID of reported content
}

// ReportStatus constants
const (
	StatusPending  = "pending"  // Report is pending review
	StatusReviewed = "reviewed" // Report has been reviewed
	StatusResolved = "resolved" // Report has been resolved
)

// NormalizeStatus normalizes status values for comparison and storage.
func NormalizeStatus(status string) string {
	return strings.ToLower(strings.TrimSpace(status))
}

// IsValidStatus checks if status is valid for reports.
func IsValidStatus(status string) bool {
	switch NormalizeStatus(status) {
	case StatusPending, StatusReviewed, StatusResolved:
		return true
	default:
		return false
	}
}

// IsValidTargetType checks if target type is supported.
func IsValidTargetType(targetType string) bool {
	switch strings.ToLower(strings.TrimSpace(targetType)) {
	case "post", "comment":
		return true
	default:
		return false
	}
}

// Validate checks if the report is valid and returns an error describing the failure.
func (r *Report) Validate() error {
	if !IsValidTargetType(r.TargetType) {
		return ErrInvalidTargetType
	}
	if !IsValidStatus(r.Status) {
		return ErrInvalidReportStatus
	}
	if strings.TrimSpace(r.Reason) == "" {
		return ErrInvalidReason
	}
	return nil
}
