// [OPTIONAL FEATURE: forum-moderation]
// Package domain contains the core business entities for the moderation module.
package domain

import (
	"strings"
	"time"
)

// ModeratorRequest represents a request from a user to become a moderator.
type ModeratorRequest struct {
	ID          int        `json:"-"`
	PublicID    string     `json:"id"`
	RequesterID int        `json:"-"`
	ReviewerID  *int       `json:"-"`
	Status      string     `json:"status"`
	Message     string     `json:"message,omitempty"`
	Response    string     `json:"response,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	ReviewedAt  *time.Time `json:"reviewed_at,omitempty"`

	PublicRequesterID string `json:"requester_id,omitempty"`
	PublicReviewerID  string `json:"reviewer_id,omitempty"`
}

const (
	RequestStatusPending  = "pending"
	RequestStatusApproved = "approved"
	RequestStatusDenied   = "denied"
)

// NormalizeRequestStatus normalizes moderator-request status values.
func NormalizeRequestStatus(status string) string {
	return strings.ToLower(strings.TrimSpace(status))
}

// IsValidRequestStatus checks if a status is valid for moderator requests.
func IsValidRequestStatus(status string) bool {
	switch NormalizeRequestStatus(status) {
	case RequestStatusPending, RequestStatusApproved, RequestStatusDenied:
		return true
	default:
		return false
	}
}

// Validate checks if the moderator request is valid.
func (r *ModeratorRequest) Validate() error {
	if r.RequesterID <= 0 {
		return ErrInvalidRequester
	}
	if !IsValidRequestStatus(r.Status) {
		return ErrInvalidRequestStatus
	}
	return nil
}
