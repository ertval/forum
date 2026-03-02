package id_audit

import (
	"encoding/json"
	authDomain "forum/internal/modules/auth/domain"
	authPorts "forum/internal/modules/auth/ports"
	postDomain "forum/internal/modules/post/domain"
	userDomain "forum/internal/modules/user/domain"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

var strictUUIDRegex = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)

// TestDomainIDHandling verifies that domain entities properly implement the UUID/public ID pattern
func TestDomainIDHandling(t *testing.T) {
	t.Run("Auth Session ID Pattern", func(t *testing.T) {
		session := &authDomain.Session{
			ID:       1,                                      // Internal int ID
			PublicID: "550e8400-e29b-41d4-a716-446655440000", // Public UUID
			UserID:   100,                                    // Internal int ID for foreign key
		}

		// Check that internal ID is an integer
		if session.ID <= 0 {
			t.Error("Session internal ID should be a positive integer")
		}

		// Check that public ID follows UUID format
		if !strictUUIDRegex.MatchString(session.PublicID) {
			t.Fatalf("Session PublicID should be a UUID format, got: %s", session.PublicID)
		}

		// Check that UserID is an internal integer
		if session.UserID <= 0 {
			t.Error("Session UserID should be a positive integer")
		}
	})

	t.Run("User Entity ID Pattern", func(t *testing.T) {
		user := &userDomain.User{
			ID:       1,                                      // Internal int ID
			PublicID: "550e8400-e29b-41d4-a716-446655440001", // Public UUID
			Email:    "test@example.com",
		}

		// Check that internal ID is an integer
		if user.ID <= 0 {
			t.Error("User internal ID should be a positive integer")
		}

		if !strictUUIDRegex.MatchString(user.PublicID) {
			t.Fatalf("User PublicID should be a UUID format, got: %s", user.PublicID)
		}
	})

	t.Run("Post Entity ID Pattern", func(t *testing.T) {
		post := &postDomain.Post{
			ID:           1,                                      // Internal int ID
			PublicID:     "550e8400-e29b-41d4-a716-446655440002", // Public UUID
			UserID:       100,                                    // Internal int ID for foreign key
			UserPublicID: "550e8400-e29b-41d4-a716-446655440001", // User's public UUID
			Title:        "Test Post",
			Content:      "Test content",
		}

		// Check that internal ID is an integer
		if post.ID <= 0 {
			t.Error("Post internal ID should be a positive integer")
		}

		if !strictUUIDRegex.MatchString(post.PublicID) {
			t.Fatalf("Post PublicID should be a UUID format, got: %s", post.PublicID)
		}
		if !strictUUIDRegex.MatchString(post.UserPublicID) {
			t.Fatalf("Post UserPublicID should be a UUID format, got: %s", post.UserPublicID)
		}

		// Check that UserID is an internal integer
		if post.UserID <= 0 {
			t.Error("Post UserID should be a positive integer")
		}
	})
}

// TestAPIResponseIDPattern verifies that API responses follow the expected pattern
func TestAPIResponseIDPattern(t *testing.T) {
	t.Run("Post API Response Structure", func(t *testing.T) {
		post := &postDomain.Post{
			ID:           1,
			PublicID:     "750e8400-1234-5678-9012-123456789012",
			UserID:       100,
			UserPublicID: "550e8400-e29b-41d4-a716-446655440001",
			Title:        "title",
			Content:      "content",
		}

		jsonData, err := json.Marshal(post)
		if err != nil {
			t.Fatalf("failed to marshal post: %v", err)
		}

		var payload map[string]any
		if err := json.Unmarshal(jsonData, &payload); err != nil {
			t.Fatalf("failed to unmarshal post payload: %v", err)
		}

		if _, exists := payload["ID"]; exists {
			t.Fatal("internal ID must not be exposed in JSON")
		}
		if _, exists := payload["UserID"]; exists {
			t.Fatal("internal UserID must not be exposed in JSON")
		}
		if payload["id"] != post.PublicID {
			t.Fatalf("expected id=%s, got %#v", post.PublicID, payload["id"])
		}
		if payload["user_id"] != post.UserPublicID {
			t.Fatalf("expected user_id=%s, got %#v", post.UserPublicID, payload["user_id"])
		}
	})

	t.Run("User API Response Structure", func(t *testing.T) {
		user := &userDomain.User{
			ID:       2,
			PublicID: "550e8400-e29b-41d4-a716-446655440010",
			Email:    "test@example.com",
			Username: "tester",
		}

		jsonData, err := json.Marshal(user)
		if err != nil {
			t.Fatalf("failed to marshal user: %v", err)
		}

		var payload map[string]any
		if err := json.Unmarshal(jsonData, &payload); err != nil {
			t.Fatalf("failed to unmarshal user payload: %v", err)
		}

		if _, exists := payload["ID"]; exists {
			t.Fatal("internal ID must not be exposed in user JSON")
		}
		if payload["id"] != user.PublicID {
			t.Fatalf("expected id=%s, got %#v", user.PublicID, payload["id"])
		}
	})
}

func TestTemplateURLPattern(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	repoRoot := wd
	for {
		if _, err := os.Stat(filepath.Join(repoRoot, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(repoRoot)
		if parent == repoRoot {
			t.Fatal("failed to locate repository root containing go.mod")
		}
		repoRoot = parent
	}

	templateFiles, err := filepath.Glob(filepath.Join(repoRoot, "templates", "*.html"))
	if err != nil {
		t.Fatalf("failed to list templates: %v", err)
	}
	if len(templateFiles) == 0 {
		t.Fatal("no template files found")
	}

	forbidden := []string{"/posts/{{.ID}}", "?user={{.ID}}", "/users/{{.ID}}"}
	foundPublicIDUsage := false

	for _, file := range templateFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("failed to read template %s: %v", file, err)
		}
		text := string(content)

		if strings.Contains(text, "PublicID") || strings.Contains(text, "PostPublicID") || strings.Contains(text, "UserPublicID") {
			foundPublicIDUsage = true
		}

		for _, pattern := range forbidden {
			if strings.Contains(text, pattern) {
				t.Fatalf("forbidden internal ID template URL pattern found in %s: %s", file, pattern)
			}
		}
	}

	if !foundPublicIDUsage {
		t.Fatal("expected templates to contain PublicID-based URL usage")
	}
}

// TestServiceMethodSignatures verifies that service methods use UUIDs where appropriate
func TestServiceMethodSignatures(t *testing.T) {
	filterType := reflect.TypeOf(postDomain.PostFilter{})
	for _, fieldName := range []string{"UserID", "CommenterID", "ReactedByUserID", "LikedByUserID", "DislikedByUserID"} {
		field, ok := filterType.FieldByName(fieldName)
		if !ok {
			t.Fatalf("expected PostFilter.%s field to exist", fieldName)
		}
		if field.Type.Kind() != reflect.String {
			t.Fatalf("expected PostFilter.%s to be string for UUID IDs, got %s", fieldName, field.Type.Kind())
		}
	}

	getUserIDType := reflect.TypeOf(authPorts.GetUserID)
	if getUserIDType.NumIn() != 1 || getUserIDType.NumOut() != 1 {
		t.Fatalf("unexpected GetUserID signature: %s", getUserIDType.String())
	}
	if getUserIDType.Out(0).Kind() != reflect.String {
		t.Fatalf("expected GetUserID to return UUID string, got %s", getUserIDType.Out(0).Kind())
	}
}
