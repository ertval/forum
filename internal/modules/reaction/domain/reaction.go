// Package domain contains the core business entities for the reaction module.
package domain

import "time"

// ReactionType represents the type of reaction (like or dislike).
type ReactionType string

// ToggleAction represents the result of an atomic reaction toggle operation.
type ToggleAction string

const (
	// ReactionLike represents a positive reaction.
	ReactionLike ReactionType = "like"

	// ReactionDislike represents a negative reaction.
	ReactionDislike ReactionType = "dislike"

	// ToggleActionCreated indicates a new reaction was created.
	ToggleActionCreated ToggleAction = "created"
	// ToggleActionUpdated indicates an existing reaction type was switched.
	ToggleActionUpdated ToggleAction = "updated"
	// ToggleActionRemoved indicates an existing same-type reaction was removed.
	ToggleActionRemoved ToggleAction = "removed"
)

// Reaction represents a user's reaction to a post or comment.
type Reaction struct {
	ID         int          `json:"-"`           // Internal unique identifier (INT PRIMARY KEY)
	PublicID   string       `json:"id"`          // Public UUID identifier (exposed in API)
	UserID     int          `json:"-"`           // Internal ID of the user who reacted
	TargetID   int          `json:"-"`           // Internal ID of the target (post or comment)
	TargetType string       `json:"target_type"` // Type of target: "post" or "comment"
	Type       ReactionType `json:"type"`        // Type of reaction: like or dislike
	CreatedAt  time.Time    `json:"created_at"`  // Reaction creation timestamp
	// For API responses - public UUIDs of related entities
	PublicUserID   string `json:"user_id,omitempty"`   // Public UUID of the user
	PublicTargetID string `json:"target_id,omitempty"` // Public UUID of target
}

// Validate checks if the reaction is valid.
func (r *Reaction) Validate() error {
	if r.TargetType != "post" && r.TargetType != "comment" {
		return ErrInvalidTarget
	}

	if r.Type != ReactionLike && r.Type != ReactionDislike {
		return ErrInvalidReactionType
	}

	if r.UserID <= 0 {
		return ErrInvalidUserID
	}

	if r.PublicTargetID == "" {
		return ErrInvalidTargetID
	}

	return nil
}
