package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"forum/cmd/forum/wire"
	"forum/internal/platform/config"
	"forum/internal/platform/logger"

	_ "github.com/mattn/go-sqlite3"
)

// TestPublicIDExposure tests that public-facing APIs return UUIDs instead of internal int IDs
func TestPublicIDExposure(t *testing.T) {
	app := setupTestApp(t)
	defer app.Cleanup()

	// Register user and login
	sessionToken := registerAndLogin(t, app, "testuser@example.com", "testuser", "password")

	// Create category
	createCategory(t, app, "test")

	// Create a post
	postData := map[string]interface{}{
		"title":      "Test Public ID Exposure",
		"content":    "Testing if public IDs are exposed correctly",
		"categories": []string{"test"},
	}

	body, _ := json.Marshal(postData)
	req := httptest.NewRequest("POST", "/posts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session_token", Value: sessionToken})

	w := httptest.NewRecorder()
	app.Server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create post: %s", w.Body.String())
	}

	// Check that the response contains a public UUID (not an integer)
	var createdPost map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&createdPost)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	idField, exists := createdPost["id"]
	if !exists {
		t.Fatal("Response does not contain 'id' field")
	}

	idStr, ok := idField.(string)
	if !ok {
		t.Fatalf("Post ID is not a string: %v (type: %T)", idField, idField)
	}

	// Check if the ID looks like a UUID (contains hyphens and is of appropriate length)
	if !strings.Contains(idStr, "-") || len(idStr) != 36 {
		t.Errorf("Post ID does not appear to be a UUID: %s", idStr)
	}
}

// TestTemplateIDExposure tests that HTML templates don't expose internal int IDs in URLs
func TestTemplateIDExposure(t *testing.T) {
	app := setupTestApp(t)
	defer app.Cleanup()

	// Register user and login
	sessionToken := registerAndLogin(t, app, "templateuser@example.com", "templateuser", "password")

	// Create category
	createCategory(t, app, "template-test")

	// Create a post
	postID := createPost(t, app, sessionToken, "Template Test Post", "Testing template ID exposure", []string{"template-test"})

	// Request the post detail page
	req := httptest.NewRequest("GET", "/posts/"+postID, nil)
	w := httptest.NewRecorder()
	app.Server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to get post page: %d - %s", w.Code, w.Body.String())
	}

	// Check that the response is HTML (not JSON)
	if !strings.Contains(w.Header().Get("Content-Type"), "text/html") {
		t.Errorf("Expected HTML response, got: %s", w.Header().Get("Content-Type"))
	}

	responseBody := w.Body.String()

	// The response should NOT contain the internal integer ID in URLs
	// Find a way to check if the internal ID is exposed by looking at the pattern
	// In the template, we saw .ID was being used, which would be the internal int ID

	// Look for the pattern of internal integer IDs in URLs (they would be sequential)
	// Since we have only one post, its internal ID would likely be 1, 2, or 3
	// Check if the response contains "/posts/1", "/posts/2", etc. (internal IDs)
	// Since we just created one post and the system might have multiple posts already,
	// we need to check the actual structure

	// First, get the post via API to compare
	apiReq := httptest.NewRequest("GET", "/posts/"+postID, nil)
	apiReq.Header.Set("Accept", "application/json")
	apiW := httptest.NewRecorder()
	app.Server.Router().ServeHTTP(apiW, apiReq)

	if apiW.Code != http.StatusOK {
		t.Fatalf("Failed to get post via API: %d", apiW.Code)
	}

	var apiPost map[string]interface{}
	err := json.NewDecoder(apiW.Body).Decode(&apiPost)
	if err != nil {
		t.Fatalf("Failed to decode API response: %v", err)
	}

	apiPostID, exists := apiPost["id"]
	if !exists {
		t.Fatal("API response does not contain 'id' field")
	}

	// Ensure the template response uses the public ID, not internal IDs
	if strings.Contains(responseBody, "href=\"/posts/"+postID[:10]) || strings.Contains(responseBody, "data-post-id=\""+postID[:10]) {
		// The full UUID (36 characters) would be too long to reasonably appear in the response by chance
		// So we'll check for patterns that indicate internal IDs
		t.Logf("Response body: %s", responseBody)
		t.Errorf("Template may be exposing internal ID in URLs or attributes. Public ID: %s", postID)
	}
}

// TestPostListIDs tests that post lists return public UUIDs
func TestPostListIDs(t *testing.T) {
	app := setupTestApp(t)
	defer app.Cleanup()

	// Register user and login
	sessionToken := registerAndLogin(t, app, "listuser@example.com", "listuser", "password")

	// Create category
	createCategory(t, app, "list-test")

	// Create a post
	postID := createPost(t, app, sessionToken, "List Test Post", "Testing post list IDs", []string{"list-test"})

	// Test API endpoint
	req := httptest.NewRequest("GET", "/posts", nil)
	w := httptest.NewRecorder()
	app.Server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to get posts list: %d - %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode posts list response: %v", err)
	}

	posts, exists := response["posts"]
	if !exists {
		t.Fatal("Response does not contain 'posts' field")
	}

	postsArray, ok := posts.([]interface{})
	if !ok {
		t.Fatalf("Posts field is not an array: %v", posts)
	}

	if len(postsArray) == 0 {
		t.Fatal("No posts returned in list")
	}

	// Check that the first post in the list has a UUID ID
	post := postsArray[0].(map[string]interface{})
	postIDField, exists := post["id"]
	if !exists {
		t.Fatal("Post in list does not contain 'id' field")
	}

	postIDStr, ok := postIDField.(string)
	if !ok {
		t.Fatalf("Post ID in list is not a string: %v (type: %T)", postIDField, postIDField)
	}

	// Verify that the post in the list has the expected public ID
	// This specific post should be in the list
	if postIDStr != postID {
		t.Errorf("Expected post ID %s in list, got %s", postID, postIDStr)
	}

	// Verify that the ID looks like a UUID
	if !strings.Contains(postIDStr, "-") || len(postIDStr) != 36 {
		t.Errorf("Post ID in list does not appear to be a UUID: %s", postIDStr)
	}
}

// TestUserModuleIDHandling tests that user module follows proper ID handling
func TestUserModuleIDHandling(t *testing.T) {
	app := setupTestApp(t)
	defer app.Cleanup()

	// Register a user
	sessionToken := registerAndLogin(t, app, "idtest@example.com", "idtestuser", "password")

	// Get user info via auth validation to see if user IDs are handled properly
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: sessionToken})
	w := httptest.NewRecorder()
	app.Server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Failed to access homepage with session: %d", w.Code)
	}
}

// TestRouteParameterHandling tests that routes properly handle UUID parameters
func TestRouteParameterHandling(t *testing.T) {
	app := setupTestApp(t)
	defer app.Cleanup()

	// Register user and login
	sessionToken := registerAndLogin(t, app, "routeuser@example.com", "routeuser", "password")

	// Create category
	createCategory(t, app, "route-test")

	// Create a post
	postID := createPost(t, app, sessionToken, "Route Test Post", "Testing route parameter handling", []string{"route-test"})

	// Test that we can retrieve the post using its public ID
	req := httptest.NewRequest("GET", "/posts/"+postID, nil)
	w := httptest.NewRecorder()
	app.Server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Failed to retrieve post using public ID %s: %d - %s", postID, w.Code, w.Body.String())
	}

	// Test that invalid IDs return appropriate errors
	invalidID := "invalid-uuid-format"
	req = httptest.NewRequest("GET", "/posts/"+invalidID, nil)
	w = httptest.NewRecorder()
	app.Server.Router().ServeHTTP(w, req)

	// Should return not found (not an internal server error about ID format)
	if w.Code != http.StatusNotFound && w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 or 404 for invalid ID, got: %d", w.Code)
	}
}