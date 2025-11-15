package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"forum/cmd/forum/wire"
	"forum/internal/platform/config"
	"forum/internal/platform/logger"

	_ "github.com/mattn/go-sqlite3"
)

// TestPostCreationAndRetrieval tests basic post creation and retrieval.
// Covers audit requirements: posts visible to all, registered users can create posts.
func TestPostCreationAndRetrieval(t *testing.T) {
	app := setupTestApp(t)
	defer app.Cleanup()

	// Register user and login
	sessionToken := registerAndLogin(t, app, "user@test.com", "testuser", "pass123")

	// Create category
	createCategory(t, app, "general")

	// Create post
	postID := createPost(t, app, sessionToken, "Test Post", "Test Content", []string{"general"})

	// Get post (public access)
	req := httptest.NewRequest("GET", "/posts/"+postID, nil)
	w := httptest.NewRecorder()
	app.Server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// TestUnauthorizedPostCreation tests that guests cannot create posts.
// Covers audit requirement: non-registered users cannot create posts.
func TestUnauthorizedPostCreation(t *testing.T) {
	app := setupTestApp(t)
	defer app.Cleanup()

	postData := map[string]interface{}{
		"title":      "Test",
		"content":    "Content",
		"categories": []string{"general"},
	}

	body, _ := json.Marshal(postData)
	req := httptest.NewRequest("POST", "/posts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	app.Server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", w.Code)
	}
}

// TestEmptyPostValidation tests validation for empty posts.
// Covers audit requirement: cannot create empty posts.
func TestEmptyPostValidation(t *testing.T) {
	app := setupTestApp(t)
	defer app.Cleanup()

	sessionToken := registerAndLogin(t, app, "user2@test.com", "user2", "pass123")

	// Empty title
	postData := map[string]interface{}{
		"title":      "",
		"content":    "Content",
		"categories": []string{"general"},
	}

	body, _ := json.Marshal(postData)
	req := httptest.NewRequest("POST", "/posts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session_token", Value: sessionToken})

	w := httptest.NewRecorder()
	app.Server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for empty title, got %d", w.Code)
	}
}

// Helper functions

func setupTestApp(t *testing.T) *wire.App {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Path:          ":memory:",
			MigrationsDir: "../../migrations",
		},
		Server:   config.ServerConfig{Port: 8080},
		Session:  config.SessionConfig{Duration: 24 * time.Hour},
		Security: config.SecurityConfig{RateLimitRequests: 100, RateLimitWindow: time.Minute},
	}

	lgr := logger.New(logger.InfoLevel, os.Stdout)
	app, err := wire.InitializeApp(cfg, lgr)
	if err != nil {
		t.Fatalf("Failed to init app: %v", err)
	}

	return app
}

func registerAndLogin(t *testing.T, app *wire.App, email, username, password string) string {
	regData := map[string]string{"email": email, "username": username, "password": password}
	body, _ := json.Marshal(regData)
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	app.Server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to register: %s", w.Body.String())
	}

	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == "session_token" {
			return cookie.Value
		}
	}

	t.Fatal("No session token")
	return ""
}

func createCategory(t *testing.T, app *wire.App, name string) {
	query := `INSERT OR IGNORE INTO categories (id, name, description, created_at) VALUES (?, ?, ?, ?)`
	_, err := app.Database.DB().Exec(query, name+"-id", name, "", time.Now())
	if err != nil {
		t.Logf("Category error: %v", err)
	}
}

func createPost(t *testing.T, app *wire.App, sessionToken, title, content string, categories []string) string {
	postData := map[string]interface{}{"title": title, "content": content, "categories": categories}
	body, _ := json.Marshal(postData)
	req := httptest.NewRequest("POST", "/posts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session_token", Value: sessionToken})

	w := httptest.NewRecorder()
	app.Server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create post: %s", w.Body.String())
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)
	return response["ID"].(string)
}
