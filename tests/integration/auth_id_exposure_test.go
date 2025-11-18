package integration

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"testing"
	"time"

	authAdapters "forum/internal/modules/auth/adapters"
	authApp "forum/internal/modules/auth/application"
	userAdapters "forum/internal/modules/user/adapters"
	userApp "forum/internal/modules/user/application"

	_ "github.com/mattn/go-sqlite3"
)

// TestAuthModule_NoInternalIDExposure verifies that the auth module
// properly uses PublicIDs instead of internal IDs for external-facing operations.
// This test addresses the security audit findings from:
// - docs/id_schema/auth_ID_EXPOSURE_SECURITY_AUDIT.md
// - docs/id_schema/AUTH_ID_EXPOSURE_AUDIT.md
func TestAuthModule_NoInternalIDExposure(t *testing.T) {
	// Setup in-memory database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}
	defer db.Close()

	// Create tables
	createTablesSQL := `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			public_id TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user',
			oauth_provider TEXT,
			oauth_provider_id TEXT,
			post_count INTEGER NOT NULL DEFAULT 0,
			comment_count INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			is_active INTEGER NOT NULL DEFAULT 1
		);
		CREATE TABLE sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			public_id TEXT UNIQUE NOT NULL,
			user_id INTEGER NOT NULL,
			token TEXT UNIQUE NOT NULL,
			expires_at DATETIME NOT NULL,
			created_at DATETIME NOT NULL,
			ip_address TEXT,
			user_agent TEXT,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);
		CREATE INDEX idx_users_email ON users(email);
		CREATE INDEX idx_users_username ON users(username);
		CREATE INDEX idx_sessions_token ON sessions(token);
	`

	if _, err = db.Exec(createTablesSQL); err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	// Setup repositories and services
	sessionRepo := authAdapters.NewSQLiteSessionRepository(db)
	userRepo := userAdapters.NewSQLiteUserRepository(db)

	authService := authApp.NewService(sessionRepo, userRepo, 24*time.Hour)
	userService := userApp.NewService(userRepo)

	// UUID regex pattern (standard UUID v4 format)
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

	t.Run("Register creates user with PublicID", func(t *testing.T) {
		ctx := context.Background()
		email := fmt.Sprintf("test%d@example.com", time.Now().UnixNano())
		username := fmt.Sprintf("user%d", time.Now().UnixNano()%1000000)
		password := "password123"

		// Register user
		userID, session, err := authService.Register(ctx, email, username, password)
		if err != nil {
			t.Fatalf("Failed to register: %v", err)
		}

		// Verify internal ID is an integer
		if userID <= 0 {
			t.Errorf("Expected positive internal user ID, got %d", userID)
		}

		// Fetch user by internal ID to get PublicID
		user, err := userService.GetByID(ctx, userID)
		if err != nil {
			t.Fatalf("Failed to get user by ID: %v", err)
		}

		// CRITICAL: Verify PublicID is a valid UUID (not an integer string)
		if !uuidRegex.MatchString(user.PublicID) {
			t.Errorf("User PublicID is not a valid UUID: %s", user.PublicID)
		}

		// CRITICAL: Ensure PublicID is not a stringified integer
		if matched, _ := regexp.MatchString(`^\d+$`, user.PublicID); matched {
			t.Errorf("User PublicID appears to be an integer string: %s (SECURITY ISSUE)", user.PublicID)
		}

		// Verify session was created
		if session == nil {
			t.Fatal("Session should not be nil")
		}
		if session.UserID != userID {
			t.Errorf("Session UserID mismatch: expected %d, got %d", userID, session.UserID)
		}
	})

	t.Run("Login and ValidateSession work correctly", func(t *testing.T) {
		ctx := context.Background()
		email := fmt.Sprintf("login%d@example.com", time.Now().UnixNano())
		username := fmt.Sprintf("loginuser%d", time.Now().UnixNano()%1000000)
		password := "password123"

		// Register user first
		userID, _, err := authService.Register(ctx, email, username, password)
		if err != nil {
			t.Fatalf("Failed to register: %v", err)
		}

		// Login
		session, err := authService.Login(ctx, email, password)
		if err != nil {
			t.Fatalf("Failed to login: %v", err)
		}

		// Verify session contains internal ID (this is correct for internal use)
		if session.UserID != userID {
			t.Errorf("Session UserID mismatch: expected %d, got %d", userID, session.UserID)
		}

		// Fetch user to verify PublicID
		user, err := userService.GetByID(ctx, session.UserID)
		if err != nil {
			t.Fatalf("Failed to get user: %v", err)
		}

		// CRITICAL: PublicID must be UUID format
		if !uuidRegex.MatchString(user.PublicID) {
			t.Errorf("User PublicID is not a valid UUID: %s", user.PublicID)
		}

		// Validate session
		validatedSession, err := authService.ValidateSession(ctx, session.Token)
		if err != nil {
			t.Fatalf("Failed to validate session: %v", err)
		}

		if validatedSession.UserID != userID {
			t.Errorf("Validated session UserID mismatch")
		}
	})

	t.Run("Handlers must use PublicID for API responses", func(t *testing.T) {
		// This test documents the expected behavior:
		// - HTTP handlers should fetch user by internal ID (from session)
		// - Then return user.PublicID in JSON responses
		// - Never return strconv.Itoa(userID) or similar

		ctx := context.Background()
		email := fmt.Sprintf("api%d@example.com", time.Now().UnixNano())
		username := fmt.Sprintf("apiuser%d", time.Now().UnixNano()%1000000)
		password := "password123"

		userID, _, err := authService.Register(ctx, email, username, password)
		if err != nil {
			t.Fatalf("Failed to register: %v", err)
		}

		// Simulate what a handler should do:
		// 1. Get internal ID from session context
		// 2. Fetch user by internal ID
		user, err := userService.GetByID(ctx, userID)
		if err != nil {
			t.Fatalf("Failed to get user: %v", err)
		}

		// 3. Return user.PublicID in response (NOT strconv.Itoa(userID))
		responseID := user.PublicID

		// Verify response ID is a UUID
		if !uuidRegex.MatchString(responseID) {
			t.Errorf("Response ID should be a UUID, got: %s", responseID)
		}

		// Verify it's NOT an integer string
		if matched, _ := regexp.MatchString(`^\d+$`, responseID); matched {
			t.Errorf("Response ID is an integer string (SECURITY VIOLATION): %s", responseID)
		}

		t.Logf("✅ Auth module correctly uses PublicID: %s (internal ID: %d)", responseID, userID)
	})
}
