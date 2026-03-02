package adapters

import (
	"context"
	"database/sql"
	"forum/internal/modules/auth/domain"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import SQLite driver
)

// Helper function to create test schema matching production
func createSessionsTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS sessions (
id INTEGER PRIMARY KEY AUTOINCREMENT,
public_id TEXT UNIQUE NOT NULL,
user_id INTEGER NOT NULL,
token TEXT UNIQUE NOT NULL,
expires_at DATETIME NOT NULL,
created_at DATETIME NOT NULL,
ip_address TEXT,
user_agent TEXT
)`)
	return err
}

// Helper to create session repository for tests, failing the test on error.
func mustNewSessionRepo(t *testing.T, db *sql.DB) *SQLiteSessionRepository {
	t.Helper()
	repo, err := NewSQLiteSessionRepository(db)
	if err != nil {
		t.Fatalf("Failed to create session repository: %v", err)
	}
	return repo.(*SQLiteSessionRepository)
}

func TestSQLiteSessionRepository_Create(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := createSessionsTable(db); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := mustNewSessionRepo(t, db)

	session := &domain.Session{
		// ID and PublicID will be set by repository
		UserID:    1,
		Token:     "test-token",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
		IPAddress: "192.168.1.1",
		UserAgent: "test-agent",
	}

	ctx := context.Background()
	err = repo.Create(ctx, session)
	if err != nil {
		t.Errorf("Create returned error: %v", err)
	}

	// Verify the session was created and IDs were set
	if session.ID == 0 {
		t.Error("Expected ID to be set after Create")
	}
	if session.PublicID == "" {
		t.Error("Expected PublicID to be set after Create")
	}

	// Verify in database
	var id int
	var publicID string
	err = db.QueryRow("SELECT id, public_id FROM sessions WHERE token = ?", session.Token).Scan(&id, &publicID)
	if err != nil {
		t.Errorf("Session was not created in database: %v", err)
	}
	if id != session.ID {
		t.Errorf("Expected ID %d, got %d", session.ID, id)
	}
	if publicID != session.PublicID {
		t.Errorf("Expected PublicID %s, got %s", session.PublicID, publicID)
	}
}

func TestSQLiteSessionRepository_GetByToken(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := createSessionsTable(db); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := mustNewSessionRepo(t, db)

	// Insert a session using repository
	session := &domain.Session{
		UserID:    1,
		Token:     "test-token",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
		IPAddress: "192.168.1.1",
		UserAgent: "test-agent",
	}

	ctx := context.Background()
	err = repo.Create(ctx, session)
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	// Retrieve the session
	result, err := repo.GetByToken(ctx, session.Token)
	if err != nil {
		t.Errorf("GetByToken returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected a session, got nil")
	}

	if result.ID != session.ID {
		t.Errorf("Expected ID %d, got %d", session.ID, result.ID)
	}
	if result.PublicID != session.PublicID {
		t.Errorf("Expected PublicID %s, got %s", session.PublicID, result.PublicID)
	}
	if result.UserID != session.UserID {
		t.Errorf("Expected UserID %d, got %d", session.UserID, result.UserID)
	}
	if result.Token != session.Token {
		t.Errorf("Expected Token %s, got %s", session.Token, result.Token)
	}
	if result.IPAddress != session.IPAddress {
		t.Errorf("Expected IPAddress %s, got %s", session.IPAddress, result.IPAddress)
	}
	if result.UserAgent != session.UserAgent {
		t.Errorf("Expected UserAgent %s, got %s", session.UserAgent, result.UserAgent)
	}
}

func TestSQLiteSessionRepository_GetByToken_NotFound(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := createSessionsTable(db); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := mustNewSessionRepo(t, db)

	ctx := context.Background()
	_, err = repo.GetByToken(ctx, "non-existent-token")
	if err == nil {
		t.Error("Expected error for non-existent token, got nil")
	} else if err != domain.ErrSessionNotFound {
		t.Errorf("Expected ErrSessionNotFound, got %v", err)
	}
}

func TestSQLiteSessionRepository_GetByUserID(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := createSessionsTable(db); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := mustNewSessionRepo(t, db)

	// Insert test sessions
	userID := 1
	now := time.Now()
	sessions := []*domain.Session{
		{
			UserID:    userID,
			Token:     "token-1",
			ExpiresAt: now.Add(1 * time.Hour),
			CreatedAt: now,
			IPAddress: "192.168.1.1",
			UserAgent: "agent-1",
		},
		{
			UserID:    userID,
			Token:     "token-2",
			ExpiresAt: now.Add(1 * time.Hour),
			CreatedAt: now,
			IPAddress: "192.168.1.2",
			UserAgent: "agent-2",
		},
		{
			UserID:    2, // Different user
			Token:     "token-3",
			ExpiresAt: now.Add(1 * time.Hour),
			CreatedAt: now,
			IPAddress: "192.168.1.3",
			UserAgent: "agent-3",
		},
	}

	ctx := context.Background()
	for _, session := range sessions {
		err = repo.Create(ctx, session)
		if err != nil {
			t.Fatalf("Failed to create test session: %v", err)
		}
	}

	result, err := repo.GetByUserID(ctx, userID)
	if err != nil {
		t.Errorf("GetByUserID returned error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 sessions for user %d, got %d", userID, len(result))
	}

	// Verify the sessions belong to the correct user
	for _, session := range result {
		if session.UserID != userID {
			t.Errorf("Session %d has UserID %d, expected %d", session.ID, session.UserID, userID)
		}
	}
}

func TestSQLiteSessionRepository_Update(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := createSessionsTable(db); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := mustNewSessionRepo(t, db)

	// Insert a session
	session := &domain.Session{
		UserID:    1,
		Token:     "test-token",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
		IPAddress: "192.168.1.1",
		UserAgent: "test-agent",
	}

	ctx := context.Background()
	err = repo.Create(ctx, session)
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	// Update the expiration time
	newExpiresAt := time.Now().Add(2 * time.Hour)
	session.ExpiresAt = newExpiresAt

	err = repo.Update(ctx, session)
	if err != nil {
		t.Errorf("Update returned error: %v", err)
	}

	// Verify the update
	var expiresAt time.Time
	err = db.QueryRow("SELECT expires_at FROM sessions WHERE token = ?", session.Token).Scan(&expiresAt)
	if err != nil {
		t.Errorf("Failed to query updated session: %v", err)
	}

	// Compare with some tolerance for timing differences
	if expiresAt.Sub(newExpiresAt) > time.Second {
		t.Errorf("Expected ExpiresAt %v, got %v", newExpiresAt, expiresAt)
	}
}

func TestSQLiteSessionRepository_Delete(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := createSessionsTable(db); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := mustNewSessionRepo(t, db)

	// Insert a session
	session := &domain.Session{
		UserID:    1,
		Token:     "test-token",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
		IPAddress: "192.168.1.1",
		UserAgent: "test-agent",
	}

	ctx := context.Background()
	err = repo.Create(ctx, session)
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	err = repo.Delete(ctx, session.Token)
	if err != nil {
		t.Errorf("Delete returned error: %v", err)
	}

	// Verify the session was deleted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sessions WHERE token = ?", session.Token).Scan(&count)
	if err != nil {
		t.Errorf("Failed to query deleted session: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 sessions after deletion, got %d", count)
	}
}

func TestSQLiteSessionRepository_Delete_NotFound(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := createSessionsTable(db); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := mustNewSessionRepo(t, db)

	ctx := context.Background()
	err = repo.Delete(ctx, "non-existent-token")
	if err == nil {
		t.Error("Expected error for non-existent token, got nil")
	} else if err != domain.ErrSessionNotFound {
		t.Errorf("Expected ErrSessionNotFound, got %v", err)
	}
}

func TestSQLiteSessionRepository_DeleteByUserID(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := createSessionsTable(db); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := mustNewSessionRepo(t, db)

	// Insert test sessions
	userID1 := 1
	userID2 := 2
	now := time.Now()
	sessions := []*domain.Session{
		{
			UserID:    userID1,
			Token:     "token-1",
			ExpiresAt: now.Add(1 * time.Hour),
			CreatedAt: now,
			IPAddress: "192.168.1.1",
			UserAgent: "agent-1",
		},
		{
			UserID:    userID1,
			Token:     "token-2",
			ExpiresAt: now.Add(1 * time.Hour),
			CreatedAt: now,
			IPAddress: "192.168.1.2",
			UserAgent: "agent-2",
		},
		{
			UserID:    userID2,
			Token:     "token-3",
			ExpiresAt: now.Add(1 * time.Hour),
			CreatedAt: now,
			IPAddress: "192.168.1.3",
			UserAgent: "agent-3",
		},
	}

	ctx := context.Background()
	for _, session := range sessions {
		err = repo.Create(ctx, session)
		if err != nil {
			t.Fatalf("Failed to create test session: %v", err)
		}
	}

	err = repo.DeleteByUserID(ctx, userID1)
	if err != nil {
		t.Errorf("DeleteByUserID returned error: %v", err)
	}

	// Verify sessions for user 1 were deleted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sessions WHERE user_id = ?", userID1).Scan(&count)
	if err != nil {
		t.Errorf("Failed to query sessions for user %d: %v", userID1, err)
	}
	if count != 0 {
		t.Errorf("Expected 0 sessions for user %d after deletion, got %d", userID1, count)
	}

	// Verify sessions for user 2 still exist
	err = db.QueryRow("SELECT COUNT(*) FROM sessions WHERE user_id = ?", userID2).Scan(&count)
	if err != nil {
		t.Errorf("Failed to query sessions for user %d: %v", userID2, err)
	}
	if count != 1 {
		t.Errorf("Expected 1 session for user %d to still exist, got %d", userID2, count)
	}
}
