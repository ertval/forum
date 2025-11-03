package domain
// Package domain contains the core comment business logic.
package domain

import "time"

// Comment represents a comment entity.
type Comment struct {
	ID        string
	PostID    string
	AuthorID  string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// IsOwner checks if the given user is the comment owner.
func (c *Comment) IsOwner(userID string) bool {
	return c.AuthorID == userID
}
