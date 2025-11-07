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
// TODO: Implement post validation (title, content, categories).
func (p *Post) Validate() error {
	// Check title is not empty and within length limits
	// Check content is not empty
	// Check categories list has at least one category
	return nil
}

// HasImage returns true if the post has an associated image.
func (p *Post) HasImage() bool {
	return p.ImageURL != ""
}
