// Package domain contains the core business entities for the post module.
package domain

// PostFilter represents post filtering options.
// This is a domain type used to specify filtering criteria for post queries.
type PostFilter struct {
	UserID               string   // Filter by user's public ID (UUID)
	Categories           []string // Filter by category names
	CommenterID          string   // Filter posts commented on by this user (public ID)
	ReactedByUserID      string   // Filter posts reacted to by this user (any reaction type)
	LikedByUserID        string   // Filter posts liked by this user (public ID)
	DislikedByUserID     string   // Filter posts disliked by this user (public ID)
	ReceivedReactionType string   // Filter posts that have this reaction type received ("like" or "dislike")
	ReceivedReactionScope string  // Scope for ReceivedReactionType: "post" (default), "comment", "post_or_comment"
	RequireCommentedPost bool     // Filter posts that have at least one comment
	RequireReactedPost   bool     // Filter posts that have at least one reaction on post or any comment under post
	DateFilter           string   // "today", "week", "month", "all" (default)
	Offset               int      // Pagination offset
	Limit                int      // Pagination limit
}
