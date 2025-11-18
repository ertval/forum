package integration

import (
	"context"
	"database/sql"
	"forum/internal/modules/user/adapters"
	"forum/internal/modules/user/domain"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// TestUserStats_PostAndCommentCounts verifies that GetUserStats returns correct counts.
func TestUserStats_PostAndCommentCounts(t *testing.T) {
	// Setup: Create in-memory database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create tables
	createTablesSQL := `CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'user',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		is_active INTEGER NOT NULL DEFAULT 1
	);
	CREATE TABLE posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE NOT NULL,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		author_id INTEGER NOT NULL,
		image_path TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
	);
	CREATE TABLE comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE NOT NULL,
		post_id INTEGER NOT NULL,
		author_id INTEGER NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
		FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
	);
	CREATE TABLE reactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE NOT NULL,
		user_id INTEGER NOT NULL,
		target_type TEXT NOT NULL,
		target_id INTEGER NOT NULL,
		reaction_type TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	if _, err := db.Exec(createTablesSQL); err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	ctx := context.Background()
	repo := adapters.NewSQLiteUserRepository(db)

	// Create test users
	now := time.Now()
	user1 := &domain.User{
		Email:        "test1@example.com",
		Username:     "testuser1",
		PasswordHash: "hash1",
		Role:         domain.RoleUser,
		CreatedAt:    now,
		UpdatedAt:    now,
		IsActive:     true,
	}

	user2 := &domain.User{
		Email:        "test2@example.com",
		Username:     "testuser2",
		PasswordHash: "hash2",
		Role:         domain.RoleUser,
		CreatedAt:    now,
		UpdatedAt:    now,
		IsActive:     true,
	}

	// Create users
	if err := repo.Create(ctx, user1); err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}
	if err := repo.Create(ctx, user2); err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	t.Logf("Created user1 with ID: %d", user1.ID)
	t.Logf("Created user2 with ID: %d", user2.ID)

	// Insert posts for user1
	_, err = db.Exec(`INSERT INTO posts (public_id, title, content, author_id, created_at, updated_at) VALUES ('post1-uuid', 'Post 1', 'Content 1', ?, ?, ?)`, user1.ID, now, now)
	if err != nil {
		t.Fatalf("Failed to insert post 1: %v", err)
	}

	_, err = db.Exec(`INSERT INTO posts (public_id, title, content, author_id, created_at, updated_at) VALUES ('post2-uuid', 'Post 2', 'Content 2', ?, ?, ?)`, user1.ID, now, now)
	if err != nil {
		t.Fatalf("Failed to insert post 2: %v", err)
	}

	// Insert one post for user2
	_, err = db.Exec(`INSERT INTO posts (public_id, title, content, author_id, created_at, updated_at) VALUES ('post3-uuid', 'Post 3', 'Content 3', ?, ?, ?)`, user2.ID, now, now)
	if err != nil {
		t.Fatalf("Failed to insert post 3: %v", err)
	}

	// Insert comments for user1 (on post 3)
	_, err = db.Exec(`INSERT INTO comments (public_id, post_id, author_id, content, created_at, updated_at) VALUES ('comment1-uuid', 3, ?, 'Comment 1', ?, ?)`, user1.ID, now, now)
	if err != nil {
		t.Fatalf("Failed to insert comment 1: %v", err)
	}

	_, err = db.Exec(`INSERT INTO comments (public_id, post_id, author_id, content, created_at, updated_at) VALUES ('comment2-uuid', 3, ?, 'Comment 2', ?, ?)`, user1.ID, now, now)
	if err != nil {
		t.Fatalf("Failed to insert comment 2: %v", err)
	}

	_, err = db.Exec(`INSERT INTO comments (public_id, post_id, author_id, content, created_at, updated_at) VALUES ('comment3-uuid', 3, ?, 'Comment 3', ?, ?)`, user1.ID, now, now)
	if err != nil {
		t.Fatalf("Failed to insert comment 3: %v", err)
	}

	// Insert reactions for user1
	_, err = db.Exec(`INSERT INTO reactions (public_id, user_id, target_type, target_id, reaction_type, created_at) VALUES ('reaction1-uuid', ?, 'post', 3, 'like', ?)`, user1.ID, now)
	if err != nil {
		t.Fatalf("Failed to insert reaction 1: %v", err)
	}

	_, err = db.Exec(`INSERT INTO reactions (public_id, user_id, target_type, target_id, reaction_type, created_at) VALUES ('reaction2-uuid', ?, 'post', 1, 'dislike', ?)`, user1.ID, now)
	if err != nil {
		t.Fatalf("Failed to insert reaction 2: %v", err)
	}

	// Test: Get stats for user1
	stats, err := repo.GetUserStats(ctx, user1.ID)
	if err != nil {
		t.Fatalf("Failed to get user stats: %v", err)
	}

	// Verify counts
	expectedPostCount := 2
	expectedCommentCount := 3
	expectedLikeCount := 1
	expectedDislikeCount := 1

	if stats.PostCount != expectedPostCount {
		t.Errorf("PostCount mismatch: expected %d, got %d", expectedPostCount, stats.PostCount)
	}

	if stats.CommentCount != expectedCommentCount {
		t.Errorf("CommentCount mismatch: expected %d, got %d", expectedCommentCount, stats.CommentCount)
	}

	if stats.LikeCount != expectedLikeCount {
		t.Errorf("LikeCount mismatch: expected %d, got %d", expectedLikeCount, stats.LikeCount)
	}

	if stats.DislikeCount != expectedDislikeCount {
		t.Errorf("DislikeCount mismatch: expected %d, got %d", expectedDislikeCount, stats.DislikeCount)
	}

	// Test: Get stats for user2
	stats2, err := repo.GetUserStats(ctx, user2.ID)
	if err != nil {
		t.Fatalf("Failed to get user2 stats: %v", err)
	}

	expectedPostCount2 := 1
	expectedCommentCount2 := 0

	if stats2.PostCount != expectedPostCount2 {
		t.Errorf("User2 PostCount mismatch: expected %d, got %d", expectedPostCount2, stats2.PostCount)
	}

	if stats2.CommentCount != expectedCommentCount2 {
		t.Errorf("User2 CommentCount mismatch: expected %d, got %d", expectedCommentCount2, stats2.CommentCount)
	}
}

// TestUserStats_EmptyStats verifies that users with no activity return zero counts.
func TestUserStats_EmptyStats(t *testing.T) {
	// Setup: Create in-memory database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create tables
	createTablesSQL := `CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'user',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		is_active INTEGER NOT NULL DEFAULT 1
	);
	CREATE TABLE posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE NOT NULL,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		author_id INTEGER NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	CREATE TABLE comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE NOT NULL,
		post_id INTEGER NOT NULL,
		author_id INTEGER NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	CREATE TABLE reactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE NOT NULL,
		user_id INTEGER NOT NULL,
		target_type TEXT NOT NULL,
		target_id INTEGER NOT NULL,
		reaction_type TEXT NOT NULL,
		created_at DATETIME NOT NULL
	);`

	if _, err := db.Exec(createTablesSQL); err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	ctx := context.Background()
	repo := adapters.NewSQLiteUserRepository(db)

	// Create test user with no activity
	now := time.Now()
	user := &domain.User{
		Email:        "inactive@example.com",
		Username:     "inactiveuser",
		PasswordHash: "hash",
		Role:         domain.RoleUser,
		CreatedAt:    now,
		UpdatedAt:    now,
		IsActive:     true,
	}

	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Test: Get stats for user with no activity
	stats, err := repo.GetUserStats(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get user stats: %v", err)
	}

	// Verify all counts are zero
	if stats.PostCount != 0 {
		t.Errorf("Expected PostCount to be 0, got %d", stats.PostCount)
	}

	if stats.CommentCount != 0 {
		t.Errorf("Expected CommentCount to be 0, got %d", stats.CommentCount)
	}

	if stats.LikeCount != 0 {
		t.Errorf("Expected LikeCount to be 0, got %d", stats.LikeCount)
	}

	if stats.DislikeCount != 0 {
		t.Errorf("Expected DislikeCount to be 0, got %d", stats.DislikeCount)
	}
}

// TestBuildCurrentUser_IntegrationWithStats tests the full flow of buildCurrentUser.
func TestBuildCurrentUser_IntegrationWithStats(t *testing.T) {
	// Setup: Create in-memory database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create tables
	createTablesSQL := `CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'user',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		is_active INTEGER NOT NULL DEFAULT 1
	);
	CREATE TABLE posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE NOT NULL,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		author_id INTEGER NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	CREATE TABLE comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE NOT NULL,
		post_id INTEGER NOT NULL,
		author_id INTEGER NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	CREATE TABLE reactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE NOT NULL,
		user_id INTEGER NOT NULL,
		target_type TEXT NOT NULL,
		target_id INTEGER NOT NULL,
		reaction_type TEXT NOT NULL,
		created_at DATETIME NOT NULL
	);`

	if _, err := db.Exec(createTablesSQL); err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	ctx := context.Background()
	repo := adapters.NewSQLiteUserRepository(db)

	// Create test user
	now := time.Now()
	user := &domain.User{
		Email:        "testbuild@example.com",
		Username:     "builduser",
		PasswordHash: "hash",
		Role:         domain.RoleUser,
		CreatedAt:    now,
		UpdatedAt:    now,
		IsActive:     true,
	}

	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Insert 5 posts
	for i := 1; i <= 5; i++ {
		_, err = db.Exec(`INSERT INTO posts (public_id, title, content, author_id, created_at, updated_at) VALUES (?, 'Post', 'Content', ?, ?, ?)`,
			"post-uuid-"+string(rune('0'+i)), user.ID, now, now)
		if err != nil {
			t.Fatalf("Failed to insert post %d: %v", i, err)
		}
	}

	// Insert 3 comments
	for i := 1; i <= 3; i++ {
		_, err = db.Exec(`INSERT INTO comments (public_id, post_id, author_id, content, created_at, updated_at) VALUES (?, ?, ?, 'Comment', ?, ?)`,
			"comment-uuid-"+string(rune('0'+i)), 1, user.ID, now, now)
		if err != nil {
			t.Fatalf("Failed to insert comment %d: %v", i, err)
		}
	}

	// Simulate buildCurrentUser flow
	// 1. Get user by ID
	fetchedUser, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to fetch user: %v", err)
	}

	// 2. Get user stats
	stats, err := repo.GetUserStats(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	// 3. Build map (simulating buildCurrentUser)
	userMap := map[string]interface{}{
		"PublicID":     fetchedUser.PublicID,
		"Username":     fetchedUser.Username,
		"Email":        fetchedUser.Email,
		"PostCount":    stats.PostCount,
		"CommentCount": stats.CommentCount,
	}

	// Verify the map has correct values
	if userMap["PostCount"] != 5 {
		t.Errorf("Expected PostCount 5, got %v", userMap["PostCount"])
	}

	if userMap["CommentCount"] != 3 {
		t.Errorf("Expected CommentCount 3, got %v", userMap["CommentCount"])
	}

	if userMap["Username"] != "builduser" {
		t.Errorf("Expected Username 'builduser', got %v", userMap["Username"])
	}

	// Verify template access pattern
	if postCount, ok := userMap["PostCount"].(int); ok {
		if postCount != 5 {
			t.Errorf("Template access pattern failed: expected 5, got %d", postCount)
		}
	} else {
		t.Error("PostCount is not an int")
	}
}
