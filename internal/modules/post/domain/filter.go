// Package domain contains the core business entities for the post module.
package domain

// PostFilter represents post filtering options.
// This is a domain type used to specify filtering criteria for post queries.
type PostFilter struct {
	UserID           string   // Filter by user's public ID (UUID)
	Categories       []string // Filter by category names
	CommenterID      string   // Filter posts commented on by this user (public ID)
	ReactedByUserID  string   // Filter posts reacted to by this user (any reaction type)
	LikedByUserID    string   // Filter posts liked by this user (public ID)
	DislikedByUserID string   // Filter posts disliked by this user (public ID)
	DateFilter       string   // "today", "week", "month", "all" (default)
	Offset           int      // Pagination offset
	Limit            int      // Pagination limit
}

// FilterParams represents query parameters for filtering.
// This is a domain type used to capture HTTP query parameters
// before converting them to a PostFilter.
type FilterParams struct {
	Category       string // Single category filter
	UserID         string // Filter by specific user's public ID
	ActivityType   string // Activity type filter for board: all, my_posts, reactions, commented_posts
	ReactionType   string // Reaction type filter: all, like, dislike
	MyPosts        bool   // Filter to current user's posts
	LikedPosts     bool   // Filter to current user's liked posts
	DislikedPosts  bool   // Filter to current user's disliked posts
	CommentedPosts bool   // Filter to posts current user has commented on
	DateFilter     string // Date range filter
	Limit          int    // Maximum results
	Offset         int    // Result offset
	CurrentUserID  string // Current authenticated user's public ID
	Commenter      string // Filter by posts commented on by this user (public ID
}
