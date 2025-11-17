// [OPTIONAL FEATURE: forum-moderation]
// Package domain contains core entities for the moderation module.
package domain

import "time"

// Report represents a moderation report for inappropriate content.
type Report struct {
	ID         int       `json:"-"`           // Internal unique identifier (INT PRIMARY KEY)
	PublicID   string    `json:"id"`          // Public UUID identifier (exposed in API)
	ReporterID int       `json:"-"`           // Internal ID of the user who created the report
	TargetID   int       `json:"-"`           // Internal ID of the reported content (post or comment)
	TargetType string    `json:"target_type"` // Type of target: "post" or "comment"
	Reason     string    `json:"reason"`      // Reason for the report
	Status     string    `json:"status"`      // Report status: "pending", "reviewed", "resolved"
	CreatedAt  time.Time `json:"created_at"`  // Report creation timestamp
	// For API responses - public UUIDs of related entities
	PublicReporterID string `json:"reporter_id,omitempty"` // Public UUID of reporter
	PublicTargetID   string `json:"target_id,omitempty"`   // Public UUID of reported content
}

// ReportStatus constants
const (
	StatusPending  = "pending"  // Report is pending review
	StatusReviewed = "reviewed" // Report has been reviewed
	StatusResolved = "resolved" // Report has been resolved
)

// IsValid checks if the report is valid.
// TODO: Implement report validation.
func (r *Report) IsValid() bool {
	// Check target type is "post" or "comment"
	// Check status is valid
	// Check reason is not empty
	return r.TargetType == "post" || r.TargetType == "comment"
}
