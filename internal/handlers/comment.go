package handlers

// comment.go contains HTTP handlers for comment operations.
// Handles comment creation on posts.

import (
	"net/http"
)

// CreateCommentHandler processes comment creation (POST)
func CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated
	// Parse form data
	// Validate comment content
	// Get post ID from form
	// Create comment in database
	// Redirect to post page
}
