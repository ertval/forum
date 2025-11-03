package domain
// Package domain contains the core moderation business logic.
package domain

import "time"

// ReportStatus represents the status of a report.
type ReportStatus string

const (
	ReportPending  ReportStatus = "pending"
	ReportReviewed ReportStatus = "reviewed"
	ReportRejected ReportStatus = "rejected"
	ReportAccepted ReportStatus = "accepted"
)

// ReportTargetType represents what is being reported.
type ReportTargetType string

const (
	ReportTargetPost    ReportTargetType = "post"
	ReportTargetComment ReportTargetType = "comment"
)

// Report represents a moderation report.
type Report struct {
	ID           string
	ReporterID   string
	ModeratorID  string           // Moderator who reviewed it
	TargetID     string           // Post or Comment ID
	TargetType   ReportTargetType // "post" or "comment"
	Reason       string
	Status       ReportStatus
	Response     string    // Admin/Moderator response
	CreatedAt    time.Time
	ReviewedAt   *time.Time
}

// IsReviewed checks if the report has been reviewed.
func (r *Report) IsReviewed() bool {
	return r.Status != ReportPending
}
