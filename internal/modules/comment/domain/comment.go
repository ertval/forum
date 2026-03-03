// Package domain contains the core business entities for the comment module.
package domain

import (
	"strings"
	"time"
)

// Comment represents a forum comment on a post.
type Comment struct {
	ID        int       `json:"-"`          // Internal unique identifier (INT PRIMARY KEY)
	PublicID  string    `json:"id"`         // Public UUID identifier (exposed in API)
	PostID    int       `json:"-"`          // Internal ID of the post this comment belongs to
	UserID    int       `json:"-"`          // Internal ID of the user who created the comment (mapped to author_id in DB)
	Content   string    `json:"content"`    // Comment text content
	CreatedAt time.Time `json:"created_at"` // Comment creation timestamp
	UpdatedAt time.Time `json:"updated_at"` // Last update timestamp
	// For API responses - public UUIDs of related entities
	PublicPostID   string `json:"post_id,omitempty"`        // Public UUID of the post
	PublicUserID   string `json:"user_id,omitempty"`        // Public UUID of the author
	AuthorUsername string `json:"author_username,omitempty"` // Author's username (populated by JOIN)
}

// Validate checks if the comment is valid.
func (c *Comment) Validate() error {
	// Check content is not empty or just whitespace
	if strings.TrimSpace(c.Content) == "" {
		return ErrEmptyContent
	}

	// Check content length limits (max 5000 characters)
	if len([]rune(c.Content)) > 5000 {
		return ErrContentTooLong
	}

	return nil
}
