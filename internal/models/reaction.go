package models

// reaction.go defines the Reaction model and related database operations.
// It handles likes and dislikes for posts and comments.

import (
	"time"
)

// Reaction represents a like or dislike on a post or comment
type Reaction struct {
	ID           int
	UserID       int
	TargetType   string // "post" or "comment"
	TargetID     int
	ReactionType int // 1 for like, -1 for dislike
	CreatedAt    time.Time
}

// CreateOrUpdateReaction adds or updates a reaction (like/dislike)
// If a reaction already exists, it updates the reaction type
// If the same reaction type is provided, it removes the reaction (toggle)
func CreateOrUpdateReaction(userID int, targetType string, targetID, reactionType int) error {
	// Check if reaction exists
	// If exists with same type, delete it (toggle off)
	// If exists with different type, update it
	// If doesn't exist, create new reaction
	return nil
}

// GetReaction retrieves a specific user's reaction on a target
func GetReaction(userID int, targetType string, targetID int) (*Reaction, error) {
	// Query reaction from database
	return nil, nil
}

// GetReactionCounts returns the count of likes and dislikes for a target
func GetReactionCounts(targetType string, targetID int) (likes, dislikes int, err error) {
	// Query and count reactions by type
	// Return separate counts for likes (1) and dislikes (-1)
	return 0, 0, nil
}

// DeleteReaction removes a reaction
func DeleteReaction(userID int, targetType string, targetID int) error {
	// Delete reaction from database
	return nil
}
