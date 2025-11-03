package domain
// Package domain contains the core post and category business logic.
package domain

import "time"

// Post represents a forum post entity.
type Post struct {
	ID         string
	Title      string
	Content    string
	AuthorID   string
	ImagePath  string    // Path to uploaded image (optional)
	Categories []string  // Category IDs
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Category represents a post category.
type Category struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
}

// IsOwner checks if the given user is the post owner.
func (p *Post) IsOwner(userID string) bool {
	return p.AuthorID == userID
}
