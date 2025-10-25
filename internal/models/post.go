package models

// post.go defines the Post model and related database operations.
// It handles post creation, retrieval, and filtering.

import (
	"time"
)

// Post represents a forum post
type Post struct {
	ID         int
	UserID     int
	Username   string // Joined from users table
	Title      string
	Content    string
	Categories []string
	LikeCount  int
	DislikeCount int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// CreatePost inserts a new post into the database
func CreatePost(userID int, title, content string, categoryIDs []int) (*Post, error) {
	// Insert post into database
	// Associate post with categories
	// Return created post with ID
	return nil, nil
}

// GetPostByID retrieves a post by its ID with associated data
func GetPostByID(id int) (*Post, error) {
	// Query post from database by ID
	// Join with users table for username
	// Get categories
	// Get reaction counts
	return nil, nil
}

// GetAllPosts retrieves all posts with pagination
func GetAllPosts(limit, offset int) ([]*Post, error) {
	// Query all posts from database
	// Join with users table
	// Get categories and reaction counts
	return nil, nil
}

// GetPostsByUserID retrieves all posts created by a specific user
func GetPostsByUserID(userID int) ([]*Post, error) {
	// Query posts by user ID
	return nil, nil
}

// GetPostsByCategory retrieves posts filtered by category
func GetPostsByCategory(categoryID int) ([]*Post, error) {
	// Query posts by category ID
	return nil, nil
}

// GetLikedPostsByUser retrieves posts liked by a specific user
func GetLikedPostsByUser(userID int) ([]*Post, error) {
	// Query posts liked by user
	return nil, nil
}
