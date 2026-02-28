package integration

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	authAdapters "forum/internal/modules/auth/adapters"
	"forum/internal/modules/auth/application"
	authDomain "forum/internal/modules/auth/domain"
	userAdapters "forum/internal/modules/user/adapters"
	userApp "forum/internal/modules/user/application"
	"forum/internal/platform/config"

	_ "github.com/mattn/go-sqlite3" // Import SQLite driver
)

// TestAuthIntegration tests the authentication system end-to-end
// using in-memory components to avoid requiring a running server
func TestAuthIntegration(t *testing.T) {
	// Setup in-memory database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}
	defer db.Close()

	// Initialize config
	cfg := &config.Config{
		Session: config.SessionConfig{
			Duration: 24 * time.Hour, // 24 hours
		},
	}

	// Setup repositories
	sessionRepo := authAdapters.NewSQLiteSessionRepository(db)
	userRepo := userAdapters.NewSQLiteUserRepository(db)
	userService := userApp.NewService(userRepo)

	// Create required tables manually since we're not running migrations in memory
	// This replicates what the migrations do
	_, err = db.Exec(`
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
			reaction_count INTEGER NOT NULL DEFAULT 0,
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
		CREATE INDEX idx_sessions_user_id ON sessions(user_id);
		CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
	`)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	// Setup service
	authService := application.NewService(sessionRepo, userService, cfg.Session.Duration)

	t.Run("Register and Login User", func(t *testing.T) {
		// Test registration
		email := fmt.Sprintf("testuser%d@example.com", time.Now().UnixNano()%1000000)
		username := "Test User"
		password := "password123"

		userID, session, err := authService.Register(context.Background(), email, username, password)
		if err != nil {
			t.Fatalf("Expected no error during registration, got: %v", err)
		}
		if userID <= 0 {
			t.Fatalf("Expected userID > 0, got: %d", userID)
		}
		if session == nil {
			t.Fatal("Expected session to not be nil")
		}
		if session.Token == "" {
			t.Fatal("Expected session token to not be empty")
		}

		// Test login with correct credentials
		loginSession, err := authService.Login(context.Background(), email, password)
		if err != nil {
			t.Fatalf("Expected no error during login, got: %v", err)
		}
		if loginSession == nil {
			t.Fatal("Expected login session to not be nil")
		}
		if loginSession.UserID != userID {
			t.Fatalf("Expected userID %d, got: %d", userID, loginSession.UserID)
		}
		if loginSession.Token == "" {
			t.Fatal("Expected login session token to not be empty")
		}

		// Test login with incorrect password
		_, err = authService.Login(context.Background(), email, "wrongpassword")
		if err == nil {
			t.Fatal("Expected error when logging in with wrong password")
		}
		if err != authDomain.ErrInvalidCredentials {
			t.Fatalf("Expected ErrInvalidCredentials, got: %v", err)
		}

		// Test session validation
		validatedSession, err := authService.ValidateSession(context.Background(), loginSession.Token)
		if err != nil {
			t.Fatalf("Expected no error during session validation, got: %v", err)
		}
		if validatedSession.UserID != userID {
			t.Fatalf("Expected validated session userID %d, got: %d", userID, validatedSession.UserID)
		}

		// Test logout
		err = authService.Logout(context.Background(), loginSession.Token)
		if err != nil {
			t.Fatalf("Expected no error during logout, got: %v", err)
		}

		// Session should be invalid after logout
		_, err = authService.ValidateSession(context.Background(), loginSession.Token)
		if err == nil {
			t.Fatal("Expected error during session validation after logout")
		}
		if err != authDomain.ErrSessionNotFound {
			t.Fatalf("Expected ErrSessionNotFound after logout, got: %v", err)
		}
	})

	t.Run("Registration Validation", func(t *testing.T) {
		// Test duplicate email
		email := fmt.Sprintf("dupetest%d@example.com", time.Now().UnixNano()%1000000)
		username := "Duplicate User"
		password := "password123"

		// First registration should succeed
		_, _, err := authService.Register(context.Background(), email, username, password)
		if err != nil {
			t.Fatalf("Expected no error for first registration, got: %v", err)
		}

		// Second registration with same email should fail
		_, _, err = authService.Register(context.Background(), email, "Other User", password)
		if err == nil {
			t.Fatal("Expected error for duplicate email registration")
		}
		if err != authDomain.ErrEmailAlreadyExists {
			t.Fatalf("Expected ErrEmailAlreadyExists for duplicate email, got: %v", err)
		}

		// Test duplicate username
		email2 := fmt.Sprintf("dupetest2%d@example.com", time.Now().UnixNano()%1000000)
		_, _, err = authService.Register(context.Background(), email2, username, password)
		if err == nil {
			t.Fatal("Expected error for duplicate username registration")
		}
		if err != authDomain.ErrUsernameAlreadyExists {
			t.Fatalf("Expected ErrUsernameAlreadyExists for duplicate username, got: %v", err)
		}
	})

	t.Run("Session Management", func(t *testing.T) {
		email := fmt.Sprintf("sessionuser%d@example.com", time.Now().UnixNano()%1000000)
		username := "Session User"
		password := "password123"

		// Register user
		userID, session, err := authService.Register(context.Background(), email, username, password)
		if err != nil {
			t.Fatalf("Expected no error during registration, got: %v", err)
		}
		if userID <= 0 {
			t.Fatalf("Expected userID > 0, got: %d", userID)
		}
		if session == nil {
			t.Fatal("Expected session to not be nil")
		}

		// Test GetSession
		retrievedSession, err := authService.GetSession(context.Background(), session.Token)
		if err != nil {
			t.Fatalf("Expected no error during GetSession, got: %v", err)
		}
		if retrievedSession.UserID != userID {
			t.Fatalf("Expected retrieved session userID %d, got: %d", userID, retrievedSession.UserID)
		}
		if retrievedSession.Token != session.Token {
			t.Fatalf("Expected retrieved session token %s, got: %s", session.Token, retrievedSession.Token)
		}

		// Test RefreshSession
		originalExpiresAt := retrievedSession.ExpiresAt
		refreshedSession, err := authService.RefreshSession(context.Background(), session.Token)
		if err != nil {
			t.Fatalf("Expected no error during RefreshSession, got: %v", err)
		}
		if refreshedSession.Token != session.Token {
			t.Fatalf("Expected refreshed session token %s, got: %s", session.Token, refreshedSession.Token)
		}
		if !refreshedSession.ExpiresAt.After(originalExpiresAt) {
			t.Fatalf("Expected refreshed session to have later expiry time")
		}
	})
}
