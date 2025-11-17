// Package domain contains the core business entities for the post module.
package domain

import "time"

// Post represents a forum post.
type Post struct {
	ID             int       `json:"-"`                         // Internal unique identifier (INT PRIMARY KEY)
	PublicID       string    `json:"id"`                        // Public UUID identifier (exposed in API)
	UserID         int       `json:"-"`                         // Internal ID of the user who created the post
	UserPublicID   string    `json:"user_id,omitempty"`         // Public UUID of the user (for API)
	AuthorUsername string    `json:"author_username,omitempty"` // Username of the post author (for display)
	Author         string    `json:"author,omitempty"`          // Alias for AuthorUsername (for compatibility)
	Title          string    `json:"title"`                     // Post title
	Content        string    `json:"content"`                   // Post content (body text)
	ImageURL       string    `json:"image_url,omitempty"`       // Optional image URL/path
	Categories     []string  `json:"categories"`                // List of category names associated with the post
	LikeCount      int       `json:"like_count"`                // Number of likes
	DislikeCount   int       `json:"dislike_count"`             // Number of dislikes
	CommentCount   int       `json:"comment_count"`             // Number of comments
	CreatedAt      time.Time `json:"created_at"`                // Post creation timestamp
	UpdatedAt      time.Time `json:"updated_at"`                // Last update timestamp
}

// Validate checks if the post is valid.
func (p *Post) Validate() error {
	if p.Title == "" {
		return ErrEmptyTitle
	}
	if len(p.Title) > 255 {
		return ErrTitleTooLong
	}
	if p.Content == "" {
		return ErrEmptyContent
	}
	if len(p.Content) > 50000 {
		return ErrContentTooLong
	}
	if len(p.Categories) == 0 {
		return ErrNoCategories
	}
	return nil
}

// HasImage returns true if the post has an associated image.
func (p *Post) HasImage() bool {
	return p.ImageURL != ""
}
