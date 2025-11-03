package domain
// Package domain contains the core reaction business logic.
package domain

import "time"

// ReactionType represents the type of reaction.
type ReactionType string

const (
	ReactionLike    ReactionType = "like"
	ReactionDislike ReactionType = "dislike"
)

// TargetType represents what the reaction is for.
type TargetType string

const (
	TargetPost    TargetType = "post"
	TargetComment TargetType = "comment"
)

// Reaction represents a like or dislike on a post or comment.
type Reaction struct {
	ID         string
	UserID     string
	TargetID   string       // Post ID or Comment ID
	TargetType TargetType   // "post" or "comment"
	Type       ReactionType // "like" or "dislike"
	CreatedAt  time.Time
}
