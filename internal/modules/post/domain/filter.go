// Package domain contains the core business entities for the post module.
package domain

// PostFilter represents post filtering options.
// This is a domain type used to specify filtering criteria for post queries.
type PostFilter struct {
	UserID        string   // Filter by user's public ID (UUID)
	Categories    []string // Filter by category names
	LikedByUserID string   // Filter posts liked by this user (public ID)
	DateFilter    string   // "today", "week", "month", "all" (default)
	Offset        int      // Pagination offset
	Limit         int      // Pagination limit
}

// FilterParams represents query parameters for filtering.
// This is a domain type used to capture HTTP query parameters
// before converting them to a PostFilter.
type FilterParams struct {
	Category      string // Single category filter
	UserID        string // Filter by specific user's public ID
	MyPosts       bool   // Filter to current user's posts
	LikedPosts    bool   // Filter to current user's liked posts
	DateFilter    string // Date range filter
	Limit         int    // Maximum results
	Offset        int    // Result offset
	CurrentUserID string // Current authenticated user's public ID
}
