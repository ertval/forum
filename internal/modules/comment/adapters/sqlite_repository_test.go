package adapters

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"forum/internal/modules/comment/domain"

	_ "github.com/mattn/go-sqlite3" // Import SQLite driver
)

// setupCommentTestDB creates an in-memory SQLite database with the correct schema
func setupCommentTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create the comments table matching the actual schema
	_, err = db.Exec(`
		CREATE TABLE comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			public_id TEXT UNIQUE NOT NULL,
			post_id INTEGER NOT NULL,
			author_id INTEGER NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_comments_post_id ON comments(post_id);
		CREATE INDEX idx_comments_author_id ON comments(author_id);
		
		-- Create posts table for join queries
		CREATE TABLE posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			public_id TEXT UNIQUE NOT NULL
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	return db
}

// TestSQLiteCommentRepository_Create tests comment creation
func TestSQLiteCommentRepository_Create(t *testing.T) {
	db := setupCommentTestDB(t)
	defer db.Close()

	repo := NewSQLiteCommentRepository(db)

	// Insert a post first (required for foreign key relationship)
	_, err := db.Exec("INSERT INTO posts (id, public_id) VALUES (?, ?)", 10, "post-uuid-10")
	if err != nil {
		t.Fatalf("Failed to insert test post: %v", err)
	}

	comment := &domain.Comment{
		PostID:  10,
		UserID:  5,
		Content: "Test comment content",
	}

	ctx := context.Background()
	err = repo.Create(ctx, comment)
	if err != nil {
		t.Errorf("Create returned error: %v", err)
	}

	// Verify the comment was created with a PublicID
	if comment.PublicID == "" {
		t.Error("PublicID was not set after Create")
	}

	// Verify the comment exists in the database
	var id int
	var publicID string
	err = db.QueryRow("SELECT id, public_id FROM comments WHERE content = ?", comment.Content).Scan(&id, &publicID)
	if err != nil {
		t.Errorf("Comment was not created in database: %v", err)
	}
	if publicID == "" {
		t.Error("PublicID not stored in database")
	}
}

// TestSQLiteCommentRepository_GetByPublicID tests retrieval by public UUID
func TestSQLiteCommentRepository_GetByPublicID(t *testing.T) {
	db := setupCommentTestDB(t)
	defer db.Close()

	repo := NewSQLiteCommentRepository(db)

	// Insert test data directly
	now := time.Now()
	testPublicID := "test-comment-uuid-123"
	_, err := db.Exec(`
		INSERT INTO comments (id, public_id, post_id, author_id, content, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		1, testPublicID, 10, 5, "Test comment content", now, now)
	if err != nil {
		t.Fatalf("Failed to insert test comment: %v", err)
	}

	ctx := context.Background()
	result, err := repo.GetByPublicID(ctx, testPublicID)
	if err != nil {
		t.Fatalf("GetByPublicID returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected comment, got nil")
	}

	if result.PublicID != testPublicID {
		t.Errorf("Expected PublicID %s, got %s", testPublicID, result.PublicID)
	}
	if result.Content != "Test comment content" {
		t.Errorf("Expected content 'Test comment content', got %s", result.Content)
	}
	if result.UserID != 5 {
		t.Errorf("Expected UserID 5, got %d", result.UserID)
	}
}

// TestSQLiteCommentRepository_GetByPublicID_NotFound tests retrieval of non-existent comment
func TestSQLiteCommentRepository_GetByPublicID_NotFound(t *testing.T) {
	db := setupCommentTestDB(t)
	defer db.Close()

	repo := NewSQLiteCommentRepository(db)

	ctx := context.Background()
	result, err := repo.GetByPublicID(ctx, "non-existent-uuid")

	if err != domain.ErrCommentNotFound {
		t.Errorf("Expected ErrCommentNotFound, got: %v", err)
	}
	if result != nil {
		t.Errorf("Expected nil result for non-existent ID, got: %v", result)
	}
}

// TestSQLiteCommentRepository_ListByPostPublicID tests listing comments for a post
func TestSQLiteCommentRepository_ListByPostPublicID(t *testing.T) {
	db := setupCommentTestDB(t)
	defer db.Close()

	repo := NewSQLiteCommentRepository(db)

	// Insert test posts
	_, err := db.Exec("INSERT INTO posts (id, public_id) VALUES (?, ?)", 10, "post-uuid-10")
	if err != nil {
		t.Fatalf("Failed to insert test post: %v", err)
	}
	_, err = db.Exec("INSERT INTO posts (id, public_id) VALUES (?, ?)", 11, "post-uuid-11")
	if err != nil {
		t.Fatalf("Failed to insert test post: %v", err)
	}

	// Insert test comments
	now := time.Now()
	_, err = db.Exec(`INSERT INTO comments (id, public_id, post_id, author_id, content, created_at, updated_at) VALUES 
		(1, 'comment-1', 10, 5, 'First comment', ?, ?),
		(2, 'comment-2', 10, 6, 'Second comment', ?, ?),
		(3, 'comment-3', 11, 5, 'Third comment', ?, ?)`,
		now, now, now, now, now, now)
	if err != nil {
		t.Fatalf("Failed to insert test comments: %v", err)
	}

	ctx := context.Background()
	result, err := repo.ListByPostPublicID(ctx, "post-uuid-10")
	if err != nil {
		t.Fatalf("ListByPostPublicID returned error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 comments for post 10, got %d", len(result))
	}
}

// TestSQLiteCommentRepository_Update tests comment update
func TestSQLiteCommentRepository_Update(t *testing.T) {
	db := setupCommentTestDB(t)
	defer db.Close()

	repo := NewSQLiteCommentRepository(db)

	// Insert a comment
	now := time.Now()
	testPublicID := "update-test-uuid"
	_, err := db.Exec(`
		INSERT INTO comments (id, public_id, post_id, author_id, content, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		1, testPublicID, 10, 5, "Original content", now, now)
	if err != nil {
		t.Fatalf("Failed to insert test comment: %v", err)
	}

	// Update the comment
	updatedComment := &domain.Comment{
		PublicID: testPublicID,
		Content:  "Updated content",
	}

	ctx := context.Background()
	err = repo.Update(ctx, updatedComment)
	if err != nil {
		t.Errorf("Update returned error: %v", err)
	}

	// Verify the update
	var content string
	err = db.QueryRow("SELECT content FROM comments WHERE public_id = ?", testPublicID).Scan(&content)
	if err != nil {
		t.Fatalf("Failed to query updated comment: %v", err)
	}
	if content != "Updated content" {
		t.Errorf("Expected 'Updated content', got %s", content)
	}
}

// TestSQLiteCommentRepository_DeleteByPublicID tests comment deletion
func TestSQLiteCommentRepository_Delete(t *testing.T) {
	db := setupCommentTestDB(t)
	defer db.Close()

	repo := NewSQLiteCommentRepository(db)

	// Insert a comment
	now := time.Now()
	testPublicID := "delete-test-uuid"
	_, err := db.Exec(`
		INSERT INTO comments (id, public_id, post_id, author_id, content, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		1, testPublicID, 10, 5, "Test comment content", now, now)
	if err != nil {
		t.Fatalf("Failed to insert test comment: %v", err)
	}

	ctx := context.Background()
	err = repo.DeleteByPublicID(ctx, testPublicID)
	if err != nil {
		t.Errorf("DeleteByPublicID returned error: %v", err)
	}

	// Verify deletion
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM comments WHERE public_id = ?", testPublicID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query comment count: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 comments after deletion, got %d", count)
	}
}

// TestSQLiteCommentRepository_ListByUser tests listing comments by user
func TestSQLiteCommentRepository_ListByUser(t *testing.T) {
	db := setupCommentTestDB(t)
	defer db.Close()

	repo := NewSQLiteCommentRepository(db)

	// Insert test posts
	_, err := db.Exec("INSERT INTO posts (id, public_id) VALUES (?, ?)", 10, "post-uuid-10")
	if err != nil {
		t.Fatalf("Failed to insert test post: %v", err)
	}
	_, err = db.Exec("INSERT INTO posts (id, public_id) VALUES (?, ?)", 11, "post-uuid-11")
	if err != nil {
		t.Fatalf("Failed to insert test post: %v", err)
	}

	// Insert test comments
	now := time.Now()
	_, err = db.Exec(`INSERT INTO comments (id, public_id, post_id, author_id, content, created_at, updated_at) VALUES 
		(1, 'comment-1', 10, 5, 'User 5 comment 1', ?, ?),
		(2, 'comment-2', 11, 5, 'User 5 comment 2', ?, ?),
		(3, 'comment-3', 10, 6, 'User 6 comment', ?, ?)`,
		now, now, now, now, now, now)
	if err != nil {
		t.Fatalf("Failed to insert test comments: %v", err)
	}

	ctx := context.Background()
	result, err := repo.ListByUser(ctx, 5)
	if err != nil {
		t.Fatalf("ListByUser returned error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 comments for user 5, got %d", len(result))
	}
}
