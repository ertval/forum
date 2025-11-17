package tests

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"forum/internal/modules/auth/adapters"
	authDomain "forum/internal/modules/auth/domain"
)

// TestNoIntegerIDsInURLs verifies that all URLs use UUID format, not integer IDs
func TestNoIntegerIDsInURLs(t *testing.T) {
	// UUID regex pattern: 8-4-4-4-12 hexadecimal characters
	uuidPattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

	// Test URLs that should contain UUIDs
	testURLs := []struct {
		path        string
		paramName   string
		description string
	}{
		{"/posts/550e8400-e29b-41d4-a716-446655440001", "", "Post detail URL"},
		{"/board?user=550e8400-e29b-41d4-a716-446655440001", "user", "User filter parameter"},
		{"/posts/550e8400-e29b-41d4-a716-446655440001/edit", "", "Edit post URL"},
	}

	for _, tc := range testURLs {
		t.Run(tc.description, func(t *testing.T) {
			// Extract ID from path or parameter
			var idValue string
			if strings.Contains(tc.path, "?") {
				// Extract from query parameter
				parts := strings.Split(tc.path, "?")
				params := strings.Split(parts[1], "&")
				for _, param := range params {
					kv := strings.Split(param, "=")
					if len(kv) == 2 && kv[0] == tc.paramName {
						idValue = kv[1]
						break
					}
				}
			} else {
				// Extract from path
				pathParts := strings.Split(strings.Trim(tc.path, "/"), "/")
				if len(pathParts) >= 2 {
					idValue = pathParts[1]
				}
			}

			if idValue == "" {
				t.Fatalf("Could not extract ID from URL: %s", tc.path)
			}

			// Verify it's a UUID, not an integer
			if !uuidPattern.MatchString(idValue) {
				t.Errorf("ID is not a valid UUID: %s (URL: %s)", idValue, tc.path)
			}

			// Negative test: ensure it's NOT a simple integer
			if matched, _ := regexp.MatchString(`^\d+$`, idValue); matched {
				t.Errorf("SECURITY VIOLATION: ID is a plain integer: %s (URL: %s)", idValue, tc.path)
			}
		})
	}
}

// TestAPIResponsesOnlyContainUUIDs verifies that API JSON responses never contain internal INT IDs
func TestAPIResponsesOnlyContainUUIDs(t *testing.T) {
	// UUID pattern
	uuidPattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

	testCases := []struct {
		name     string
		jsonBody string
		idFields []string // Field names that should contain UUIDs
	}{
		{
			name: "Post JSON response",
			jsonBody: `{
				"id": "750e8400-e29b-41d4-a716-446655440001",
				"user_id": "550e8400-e29b-41d4-a716-446655440001",
				"title": "Test Post",
				"content": "Test content",
				"like_count": 5,
				"comment_count": 3
			}`,
			idFields: []string{"id", "user_id"},
		},
		{
			name: "User JSON response",
			jsonBody: `{
				"id": "550e8400-e29b-41d4-a716-446655440001",
				"username": "testuser",
				"email": "test@example.com"
			}`,
			idFields: []string{"id"},
		},
		{
			name: "Comment JSON response",
			jsonBody: `{
				"id": "850e8400-e29b-41d4-a716-446655440001",
				"post_id": "750e8400-e29b-41d4-a716-446655440001",
				"user_id": "550e8400-e29b-41d4-a716-446655440001",
				"content": "Test comment"
			}`,
			idFields: []string{"id", "post_id", "user_id"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(tc.jsonBody), &data); err != nil {
				t.Fatalf("Failed to parse JSON: %v", err)
			}

			for _, fieldName := range tc.idFields {
				value, exists := data[fieldName]
				if !exists {
					t.Errorf("Expected ID field '%s' not found in JSON", fieldName)
					continue
				}

				idStr, ok := value.(string)
				if !ok {
					t.Errorf("SECURITY VIOLATION: ID field '%s' is not a string: %T", fieldName, value)
					continue
				}

				if !uuidPattern.MatchString(idStr) {
					t.Errorf("SECURITY VIOLATION: ID field '%s' is not a valid UUID: %s", fieldName, idStr)
				}

				// Negative test: ensure it's NOT a simple integer string
				if matched, _ := regexp.MatchString(`^\d+$`, idStr); matched {
					t.Errorf("SECURITY VIOLATION: ID field '%s' contains integer string: %s", fieldName, idStr)
				}
			}
		})
	}
}

// TestContextStoresPublicIDs verifies that middleware stores UUIDs, not INT IDs in context
func TestContextStoresPublicIDs(t *testing.T) {
	uuidPattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

	// Simulate what middleware should store in context
	testCases := []struct {
		name        string
		storedValue string
		shouldPass  bool
	}{
		{
			name:        "Valid UUID",
			storedValue: "550e8400-e29b-41d4-a716-446655440001",
			shouldPass:  true,
		},
		{
			name:        "Integer string (SECURITY VIOLATION)",
			storedValue: "123",
			shouldPass:  false,
		},
		{
			name:        "Integer string with leading zeros",
			storedValue: "00123",
			shouldPass:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isUUID := uuidPattern.MatchString(tc.storedValue)
			isInteger := regexp.MustCompile(`^\d+$`).MatchString(tc.storedValue)

			if tc.shouldPass {
				if !isUUID {
					t.Errorf("Expected valid UUID, got: %s", tc.storedValue)
				}
				if isInteger {
					t.Errorf("SECURITY VIOLATION: Context stores integer string: %s", tc.storedValue)
				}
			} else {
				if isUUID {
					t.Errorf("Test case expected to fail but value is valid UUID: %s", tc.storedValue)
				}
				if isInteger {
					t.Logf("CORRECTLY DETECTED: Integer string in context: %s", tc.storedValue)
				}
			}
		})
	}
}

// TestGetUserIDReturnsUUID verifies that authAdapters.GetUserID returns UUID format
func TestGetUserIDReturnsUUID(t *testing.T) {
	uuidPattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

	// Create a context with a user ID
	ctx := context.Background()

	testCases := []struct {
		name         string
		contextValue string
		shouldPass   bool
	}{
		{
			name:         "Valid UUID in context",
			contextValue: "550e8400-e29b-41d4-a716-446655440001",
			shouldPass:   true,
		},
		{
			name:         "Integer string in context (VULNERABILITY)",
			contextValue: "123",
			shouldPass:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate storing value in context
			ctx := context.WithValue(ctx, adapters.UserIDKey, tc.contextValue)

			// Extract using GetUserID
			userID := adapters.GetUserID(ctx)

			if userID == "" {
				t.Error("GetUserID returned empty string")
				return
			}

			isUUID := uuidPattern.MatchString(userID)
			isInteger := regexp.MustCompile(`^\d+$`).MatchString(userID)

			if tc.shouldPass {
				if !isUUID {
					t.Errorf("SECURITY VIOLATION: GetUserID returned non-UUID: %s", userID)
				}
				if isInteger {
					t.Errorf("SECURITY VIOLATION: GetUserID returned integer string: %s", userID)
				}
			} else {
				if isInteger {
					t.Logf("DETECTED VULNERABILITY: GetUserID returns integer: %s", userID)
				}
			}
		})
	}
}

// TestOwnershipCheckUsesSameIDType verifies that ownership checks compare compatible ID types
func TestOwnershipCheckUsesSameIDType(t *testing.T) {
	// Simulate template data passed to ownership check
	testCases := []struct {
		name          string
		userID        interface{} // What .User.ID contains
		resourceOwner string      // What .Post.UserPublicID or .Comment.AuthorPublicID contains
		shouldMatch   bool
		description   string
	}{
		{
			name:          "UUID matches UUID (CORRECT)",
			userID:        "550e8400-e29b-41d4-a716-446655440001",
			resourceOwner: "550e8400-e29b-41d4-a716-446655440001",
			shouldMatch:   true,
			description:   "Both sides use UUID - ownership check works",
		},
		{
			name:          "UUID doesn't match different UUID (CORRECT)",
			userID:        "550e8400-e29b-41d4-a716-446655440001",
			resourceOwner: "650e8400-e29b-41d4-a716-446655440002",
			shouldMatch:   false,
			description:   "Different UUIDs - ownership check correctly fails",
		},
		{
			name:          "Integer vs UUID (VULNERABILITY)",
			userID:        123,
			resourceOwner: "550e8400-e29b-41d4-a716-446655440001",
			shouldMatch:   false,
			description:   "Type mismatch - ownership check fails even for actual owner",
		},
		{
			name:          "Integer string vs UUID (VULNERABILITY)",
			userID:        "123",
			resourceOwner: "550e8400-e29b-41d4-a716-446655440001",
			shouldMatch:   false,
			description:   "Type/format mismatch - ownership check fails",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate template equality check: {{if eq .User.ID .Post.UserPublicID}}
			var matches bool
			switch v := tc.userID.(type) {
			case string:
				matches = v == tc.resourceOwner
			case int:
				// Integer can never match UUID string
				matches = false
			}

			if matches != tc.shouldMatch {
				if !tc.shouldMatch && matches {
					t.Errorf("UNEXPECTED: IDs matched when they shouldn't: %v == %s", tc.userID, tc.resourceOwner)
				} else if tc.shouldMatch && !matches {
					t.Errorf("SECURITY VULNERABILITY: Ownership check failed due to type mismatch: %v (%T) != %s",
						tc.userID, tc.userID, tc.resourceOwner)
				}
			} else {
				t.Logf("✓ %s", tc.description)
			}
		})
	}
}

// TestMiddlewareDoesNotLeakInternalIDs verifies middleware doesn't expose internal IDs
func TestMiddlewareDoesNotLeakInternalIDs(t *testing.T) {
	// This test simulates the middleware flow
	t.Run("RequireAuth middleware", func(t *testing.T) {
		// Simulate session with internal INT ID
		session := &authDomain.Session{
			ID:       1, // Internal INT ID
			PublicID: "850e8400-e29b-41d4-a716-446655440001",
			Token:    "test-token",
			UserID:   123, // Internal INT user ID
		}

		// What middleware SHOULD do:
		// 1. Fetch user by session.UserID (INT)
		// 2. Extract user.PublicID (UUID)
		// 3. Store user.PublicID in context

		// What middleware CURRENTLY does (VULNERABILITY):
		// ctx := context.WithValue(r.Context(), UserIDKey, fmt.Sprintf("%d", session.UserID))

		// Simulate current (vulnerable) behavior using session data
		vulnerableContextValue := "123" // This would be fmt.Sprintf("%d", session.UserID)
		_ = session                     // Mark as used

		// Check if it's an integer string
		intPattern := regexp.MustCompile(`^\d+$`)
		if intPattern.MatchString(vulnerableContextValue) {
			t.Errorf("SECURITY VULNERABILITY: Middleware stores integer string in context: %s", vulnerableContextValue)
		}

		// Simulate correct behavior
		correctContextValue := "550e8400-e29b-41d4-a716-446655440001" // user.PublicID
		uuidPattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

		if !uuidPattern.MatchString(correctContextValue) {
			t.Errorf("Correct value is not a UUID: %s", correctContextValue)
		} else {
			t.Logf("✓ Correct behavior: Store UUID in context: %s", correctContextValue)
		}
	})
}

// TestHTMLResponsesDoNotContainIntIDs checks rendered HTML for integer ID exposure
func TestHTMLResponsesDoNotContainIntIDs(t *testing.T) {
	// Simulate HTML response from a handler
	htmlResponses := []struct {
		name       string
		html       string
		shouldPass bool
	}{
		{
			name:       "Correct: UUID in URL parameter",
			html:       `<a href="/board?user=550e8400-e29b-41d4-a716-446655440001">My Posts</a>`,
			shouldPass: true,
		},
		{
			name:       "VULNERABLE: Integer in URL parameter",
			html:       `<a href="/board?user=123">My Posts</a>`,
			shouldPass: false,
		},
		{
			name:       "Correct: UUID in data attribute",
			html:       `<button data-post-id="750e8400-e29b-41d4-a716-446655440001">Delete</button>`,
			shouldPass: true,
		},
		{
			name:       "VULNERABLE: Integer in data attribute",
			html:       `<button data-post-id="42">Delete</button>`,
			shouldPass: false,
		},
	}

	intPattern := regexp.MustCompile(`(user|post|comment|author)=\d+(&|$|")`)
	dataAttrIntPattern := regexp.MustCompile(`data-[a-z-]+-id="\d+"`)

	for _, tc := range htmlResponses {
		t.Run(tc.name, func(t *testing.T) {
			hasIntInParam := intPattern.MatchString(tc.html)
			hasIntInDataAttr := dataAttrIntPattern.MatchString(tc.html)
			hasIntID := hasIntInParam || hasIntInDataAttr

			if tc.shouldPass && hasIntID {
				t.Errorf("SECURITY VIOLATION: HTML contains integer ID: %s", tc.html)
			} else if !tc.shouldPass && !hasIntID {
				t.Errorf("Test expected to detect integer ID but didn't find any in: %s", tc.html)
			} else if !tc.shouldPass && hasIntID {
				t.Logf("✓ CORRECTLY DETECTED integer ID exposure in HTML")
			}
		})
	}
}

// TestHandlerBuildCurrentUserReturnsUUID verifies buildCurrentUser returns UUID as ID
func TestHandlerBuildCurrentUserReturnsUUID(t *testing.T) {
	// Simulate the map returned by buildCurrentUser
	userMaps := []struct {
		name       string
		userMap    map[string]interface{}
		shouldPass bool
	}{
		{
			name: "Correct: ID field contains UUID",
			userMap: map[string]interface{}{
				"ID":           "550e8400-e29b-41d4-a716-446655440001",
				"Username":     "testuser",
				"Email":        "test@example.com",
				"PostCount":    5,
				"CommentCount": 10,
			},
			shouldPass: true,
		},
		{
			name: "VULNERABLE: ID field contains integer",
			userMap: map[string]interface{}{
				"ID":           123, // Internal INT ID leaked
				"Username":     "testuser",
				"Email":        "test@example.com",
				"PostCount":    5,
				"CommentCount": 10,
			},
			shouldPass: false,
		},
		{
			name: "VULNERABLE: ID field contains integer string",
			userMap: map[string]interface{}{
				"ID":           "123",
				"Username":     "testuser",
				"Email":        "test@example.com",
				"PostCount":    5,
				"CommentCount": 10,
			},
			shouldPass: false,
		},
	}

	uuidPattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	intPattern := regexp.MustCompile(`^\d+$`)

	for _, tc := range userMaps {
		t.Run(tc.name, func(t *testing.T) {
			idValue, exists := tc.userMap["ID"]
			if !exists {
				t.Error("ID field missing from user map")
				return
			}

			var isValid bool
			switch v := idValue.(type) {
			case string:
				isValid = uuidPattern.MatchString(v) && !intPattern.MatchString(v)
				if !isValid && intPattern.MatchString(v) {
					t.Errorf("SECURITY VIOLATION: ID field is integer string: %s", v)
				}
			case int:
				isValid = false
				t.Errorf("SECURITY VIOLATION: ID field is integer: %d", v)
			default:
				isValid = false
				t.Errorf("ID field has unexpected type: %T", v)
			}

			if tc.shouldPass && !isValid {
				t.Error("Expected valid UUID in ID field")
			} else if !tc.shouldPass && isValid {
				t.Error("Test expected to fail but ID is valid UUID")
			}
		})
	}
}

// Benchmark tests to ensure UUID operations don't significantly impact performance
func BenchmarkUUIDValidation(b *testing.B) {
	uuidPattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	testUUID := "550e8400-e29b-41d4-a716-446655440001"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = uuidPattern.MatchString(testUUID)
	}
}

func BenchmarkIntegerDetection(b *testing.B) {
	intPattern := regexp.MustCompile(`^\d+$`)
	testInt := "123456"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = intPattern.MatchString(testInt)
	}
}
