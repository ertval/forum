// [OPTIONAL FEATURE: forum-moderation]
// Package domain contains core entities for the moderation module.
package domain

import "time"

// Report represents a moderation report for inappropriate content.
type Report struct {
	ID         int       // Unique report identifier
	ReporterID int       // ID of the user who created the report
	TargetID   int       // ID of the reported content (post or comment)
	TargetType string    // Type of target: "post" or "comment"
	Reason     string    // Reason for the report
	Status     string    // Report status: "pending", "reviewed", "resolved"
	CreatedAt  time.Time // Report creation timestamp
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
