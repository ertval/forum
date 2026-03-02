package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// TestUnauthorizedCommentCreation tests that guests cannot create comments.
// Covers audit.md requirement: "Are you forbidden from creating a comment?" (line 87-89)
func TestUnauthorizedCommentCreation(t *testing.T) {
	app := setupTestApp(t)
	defer app.Cleanup()

	// First, create a post as a logged-in user to get a valid post ID
	sessionToken := registerAndLogin(t, app, "commentpermtest@test.com", "Comment Perm User", "Password123")
	createCategory(t, app, "comment-test")
	postID := createPost(t, app, sessionToken, "Test Post for Comments", "Test content", []string{"comment-test"})

	// Now try to create a comment WITHOUT authentication
	commentData := map[string]string{
		"content": "This is a test comment",
	}

	body, _ := json.Marshal(commentData)
	req := httptest.NewRequest("POST", "/api/comments/posts/"+postID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	// Note: No session cookie added - simulating a guest user

	w := httptest.NewRecorder()
	app.Server.Router().ServeHTTP(w, req)

	// Should return 401 Unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized for guest comment creation, got %d: %s", w.Code, w.Body.String())
	}
}

// TestEmptyCommentValidation tests that empty comments are rejected.
// Covers audit.md requirement: "Were you forbidden from creating the empty comment?" (line 103-105)
func TestEmptyCommentValidation(t *testing.T) {
	app := setupTestApp(t)
	defer app.Cleanup()

	// Register and login
	sessionToken := registerAndLogin(t, app, "emptycommenttest@test.com", "Empty Comment User", "Password123")
	createCategory(t, app, "empty-comment-test")
	postID := createPost(t, app, sessionToken, "Test Post for Empty Comment", "Test content", []string{"empty-comment-test"})

	// Try to create an empty comment
	commentData := map[string]string{
		"content": "", // Empty content
	}

	body, _ := json.Marshal(commentData)
	req := httptest.NewRequest("POST", "/api/comments/posts/"+postID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session_token", Value: sessionToken})

	w := httptest.NewRecorder()
	app.Server.Router().ServeHTTP(w, req)

	// Should return 400 Bad Request
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request for empty comment, got %d: %s", w.Code, w.Body.String())
	}
}

// TestAuthorizedCommentCreation tests that registered users can create comments.
// Covers audit.md requirement: "Were you able to create the comment?" (line 99-101)
func TestAuthorizedCommentCreation(t *testing.T) {
	app := setupTestApp(t)
	defer app.Cleanup()

	// Register and login
	sessionToken := registerAndLogin(t, app, "authcommenttest@test.com", "Auth Comment User", "Password123")
	createCategory(t, app, "auth-comment-test")
	postID := createPost(t, app, sessionToken, "Test Post for Auth Comment", "Test content", []string{"auth-comment-test"})

	// Create a comment as an authenticated user
	commentData := map[string]string{
		"content": "This is a valid comment from an authenticated user",
	}

	body, _ := json.Marshal(commentData)
	req := httptest.NewRequest("POST", "/api/comments/posts/"+postID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session_token", Value: sessionToken})

	w := httptest.NewRecorder()
	app.Server.Router().ServeHTTP(w, req)

	// Should return 201 Created
	if w.Code != http.StatusCreated {
		t.Errorf("Expected 201 Created for authenticated comment creation, got %d: %s", w.Code, w.Body.String())
		return
	}

	// Verify response contains comment ID (UUID format)
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if _, exists := response["id"]; !exists {
		t.Error("Response should contain 'id' field")
	}
}

// TestDeleteCommentReturnsNoContent verifies DELETE comment returns 204 with an empty body.
func TestDeleteCommentReturnsNoContent(t *testing.T) {
	app := setupTestApp(t)
	defer app.Cleanup()

	sessionToken := registerAndLogin(t, app, "deletecommenttest@test.com", "Delete Comment User", "Password123")
	createCategory(t, app, "delete-comment-test")
	postID := createPost(t, app, sessionToken, "Post For Delete Comment", "Test content", []string{"delete-comment-test"})

	commentData := map[string]string{"content": "comment to delete"}
	body, _ := json.Marshal(commentData)
	createReq := httptest.NewRequest("POST", "/api/comments/posts/"+postID, bytes.NewBuffer(body))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.AddCookie(&http.Cookie{Name: "session_token", Value: sessionToken})

	createW := httptest.NewRecorder()
	app.Server.Router().ServeHTTP(createW, createReq)

	if createW.Code != http.StatusCreated {
		t.Fatalf("Expected 201 Created for comment creation, got %d: %s", createW.Code, createW.Body.String())
	}

	var created map[string]interface{}
	if err := json.NewDecoder(createW.Body).Decode(&created); err != nil {
		t.Fatalf("Failed to decode created comment: %v", err)
	}
	commentID, ok := created["id"].(string)
	if !ok || commentID == "" {
		t.Fatalf("Expected created comment id, got: %#v", created["id"])
	}

	deleteReq := httptest.NewRequest("DELETE", "/api/comments/"+commentID, nil)
	deleteReq.AddCookie(&http.Cookie{Name: "session_token", Value: sessionToken})
	deleteW := httptest.NewRecorder()
	app.Server.Router().ServeHTTP(deleteW, deleteReq)

	if deleteW.Code != http.StatusNoContent {
		t.Fatalf("Expected 204 No Content, got %d: %s", deleteW.Code, deleteW.Body.String())
	}
	if deleteW.Body.Len() != 0 {
		t.Fatalf("Expected empty response body for 204, got: %q", deleteW.Body.String())
	}
}

// TestListCommentsEmptyArrayNotNull verifies comments list serializes empty list as [] and not null.
func TestListCommentsEmptyArrayNotNull(t *testing.T) {
	app := setupTestApp(t)
	defer app.Cleanup()

	sessionToken := registerAndLogin(t, app, "emptylistcommenttest@test.com", "Empty List User", "Password123")
	createCategory(t, app, "empty-list-comment-test")
	postID := createPost(t, app, sessionToken, "Post Without Comments", "Test content", []string{"empty-list-comment-test"})

	listReq := httptest.NewRequest("GET", "/api/comments/posts/"+postID, nil)
	listW := httptest.NewRecorder()
	app.Server.Router().ServeHTTP(listW, listReq)

	if listW.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for listing comments, got %d: %s", listW.Code, listW.Body.String())
	}

	var response struct {
		Comments []map[string]interface{} `json:"comments"`
	}
	if err := json.NewDecoder(listW.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode comments list response: %v", err)
	}

	if response.Comments == nil {
		t.Fatalf("Expected comments to be an empty array, got null")
	}
	if len(response.Comments) != 0 {
		t.Fatalf("Expected zero comments, got %d", len(response.Comments))
	}
}
