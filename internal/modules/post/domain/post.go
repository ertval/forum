// Package domain contains the core business entities for the post module.
package domain

import "time"

// Post represents a forum post.
type Post struct {
	ID             string    // Unique post identifier (UUID)
	UserID         string    // ID of the user who created the post (UUID)
	AuthorUsername string    // Username of the post author (for display)
	Title          string    // Post title
	Content        string    // Post content (body text)
	ImageURL       string    // Optional image URL/path
	Categories     []string  // List of category names associated with the post
	LikeCount      int       // Number of likes
	DislikeCount   int       // Number of dislikes
	CommentCount   int       // Number of comments
	CreatedAt      time.Time // Post creation timestamp
	UpdatedAt      time.Time // Last update timestamp
}

// Validate checks if the post is valid.
func (p *Post) Validate() error {
	if p.Title == "" {
		return ErrEmptyTitle
	}
	if len(p.Title) > 300 {
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
