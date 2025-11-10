package integration

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"forum/internal/modules/auth/application"
	authDomain "forum/internal/modules/auth/domain"
	authAdapters "forum/internal/modules/auth/adapters"
	userAdapters "forum/internal/modules/user/adapters"
	"forum/internal/platform/config"
	_ "github.com/mattn/go-sqlite3" // Import SQLite driver
)

// TestAuthIntegration tests the authentication system end-to-end
// using in-memory components to avoid requiring a running server
func TestAuthIntegration(t *testing.T) {
	// Setup in-memory database
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
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

	// Create required tables manually since we're not running migrations in memory
	// This replicates what the migrations do
	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user',
			oauth_provider TEXT,
			oauth_provider_id TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			is_active INTEGER NOT NULL DEFAULT 1
		);
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
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
	require.NoError(t, err)

	// Setup service
	authService := application.NewService(sessionRepo, userRepo, cfg.Session.Duration)

	t.Run("Register and Login User", func(t *testing.T) {
		// Test registration
		email := fmt.Sprintf("testuser%d@example.com", time.Now().UnixNano()%1000000)
		username := "testuser"
		password := "password123"

		userID, session, err := authService.Register(context.Background(), email, username, password)
		require.NoError(t, err)
		assert.Greater(t, userID, 0)
		assert.NotNil(t, session)
		assert.NotEmpty(t, session.Token)

		// Test login with correct credentials
		loginSession, err := authService.Login(context.Background(), email, password)
		require.NoError(t, err)
		assert.NotNil(t, loginSession)
		assert.Equal(t, userID, loginSession.UserID)
		assert.NotEmpty(t, loginSession.Token)

		// Test login with incorrect password
		_, err = authService.Login(context.Background(), email, "wrongpassword")
		assert.Error(t, err)
		assert.Equal(t, authDomain.ErrInvalidCredentials, err)

		// Test session validation
		validatedSession, err := authService.ValidateSession(context.Background(), loginSession.Token)
		require.NoError(t, err)
		assert.Equal(t, userID, validatedSession.UserID)

		// Test logout
		err = authService.Logout(context.Background(), loginSession.Token)
		require.NoError(t, err)

		// Session should be invalid after logout
		_, err = authService.ValidateSession(context.Background(), loginSession.Token)
		assert.Error(t, err)
		assert.Equal(t, authDomain.ErrSessionNotFound, err)
	})

	t.Run("Registration Validation", func(t *testing.T) {
		// Test duplicate email
		email := fmt.Sprintf("dupetest%d@example.com", time.Now().UnixNano()%1000000)
		username := "dupuser"
		password := "password123"

		// First registration should succeed
		_, _, err := authService.Register(context.Background(), email, username, password)
		require.NoError(t, err)

		// Second registration with same email should fail
		_, _, err = authService.Register(context.Background(), email, username+"2", password)
		assert.Error(t, err)
		assert.Equal(t, authDomain.ErrUserAlreadyExists, err)

		// Test duplicate username
		email2 := fmt.Sprintf("dupetest2%d@example.com", time.Now().UnixNano()%1000000)
		_, _, err = authService.Register(context.Background(), email2, username, password)
		assert.Error(t, err)
		assert.Equal(t, authDomain.ErrUserAlreadyExists, err)
	})

	t.Run("Session Management", func(t *testing.T) {
		email := fmt.Sprintf("sessionuser%d@example.com", time.Now().UnixNano()%1000000)
		username := "sessionuser"
		password := "password123"

		// Register user
		userID, session, err := authService.Register(context.Background(), email, username, password)
		require.NoError(t, err)
		assert.Greater(t, userID, 0)
		assert.NotNil(t, session)

		// Test GetSession
		retrievedSession, err := authService.GetSession(context.Background(), session.Token)
		require.NoError(t, err)
		assert.Equal(t, userID, retrievedSession.UserID)
		assert.Equal(t, session.Token, retrievedSession.Token)

		// Test RefreshSession
		originalExpiresAt := retrievedSession.ExpiresAt
		refreshedSession, err := authService.RefreshSession(context.Background(), session.Token)
		require.NoError(t, err)
		assert.Equal(t, session.Token, refreshedSession.Token)
		assert.True(t, refreshedSession.ExpiresAt.After(originalExpiresAt))
	})
}