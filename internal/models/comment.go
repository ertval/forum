package models

// comment.go defines the Comment model and related database operations.
// It handles comment creation and retrieval.

import (
	"time"
)

// Comment represents a comment on a post
type Comment struct {
	ID           int
	PostID       int
	UserID       int
	Username     string // Joined from users table
	Content      string
	LikeCount    int
	DislikeCount int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// CreateComment inserts a new comment into the database
func CreateComment(postID, userID int, content string) (*Comment, error) {
	// Insert comment into database
	// Return created comment with ID
	return nil, nil
}

// GetCommentByID retrieves a comment by its ID
func GetCommentByID(id int) (*Comment, error) {
	// Query comment from database by ID
	return nil, nil
}

// GetCommentsByPostID retrieves all comments for a specific post
func GetCommentsByPostID(postID int) ([]*Comment, error) {
	// Query comments from database by post ID
	// Join with users table for username
	// Get reaction counts
	return nil, nil
}
