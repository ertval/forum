// Package domain contains core entities for the reaction module.
package domain

import "time"

// ReactionType represents the type of reaction (like or dislike).
type ReactionType string

const (
	// ReactionLike represents a positive reaction.
	ReactionLike ReactionType = "like"

	// ReactionDislike represents a negative reaction.
	ReactionDislike ReactionType = "dislike"
)

// Reaction represents a user's reaction to a post or comment.
type Reaction struct {
	ID         int          // Unique reaction identifier
	UserID     int          // ID of the user who reacted
	TargetID   int          // ID of the target (post or comment)
	TargetType string       // Type of target: "post" or "comment"
	Type       ReactionType // Type of reaction: like or dislike
	CreatedAt  time.Time    // Reaction creation timestamp
}

// IsValid checks if the reaction is valid.
// TODO: Implement reaction validation.
func (r *Reaction) IsValid() bool {
	// Check target type is "post" or "comment"
	// Check reaction type is "like" or "dislike"
	return r.TargetType == "post" || r.TargetType == "comment"
}
