// Package domain contains core entities for the comment module.
package domain

import "time"

// Comment represents a forum comment on a post.
type Comment struct {
	ID        int       // Internal unique identifier (INT PRIMARY KEY)
	PublicID  string    // Public UUID identifier (exposed in API)
	PostID    int       // Internal ID of the post this comment belongs to
	UserID    int       // Internal ID of the user who created the comment
	Content   string    // Comment text content
	CreatedAt time.Time // Comment creation timestamp
	UpdatedAt time.Time // Last update timestamp
}

// Validate checks if the comment is valid.
// TODO: Implement comment validation (non-empty content, length limits).
func (c *Comment) Validate() error {
	// Check content is not empty
	// Check content length limits
	return nil
}
