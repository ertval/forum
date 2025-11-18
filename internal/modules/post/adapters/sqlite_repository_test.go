package adapters

import (
	"context"
	"database/sql"
	"forum/internal/modules/post/domain"
	"forum/internal/modules/post/ports"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import SQLite driver
)

// setupTestDB creates an in-memory SQLite database with the correct schema
func setupTestDB(t *testing.T) *sql.DB {
	// Use shared in-memory SQLite so multiple connections see same schema.
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create categories table with correct schema
	_, err = db.Exec(`CREATE TABLE categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE NOT NULL,
		name TEXT UNIQUE NOT NULL,
		description TEXT,
		created_at DATETIME NOT NULL
	)`)
	if err != nil {
		t.Fatalf("Failed to create categories table: %v", err)
	}

	// Create posts table with correct schema
	_, err = db.Exec(`CREATE TABLE posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE NOT NULL,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		author_id INTEGER NOT NULL,
		image_path TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	)`)
	if err != nil {
		t.Fatalf("Failed to create posts table: %v", err)
	}

	// Create post_categories table with correct schema
	_, err = db.Exec(`CREATE TABLE post_categories (
		post_id INTEGER NOT NULL,
		category_id INTEGER NOT NULL,
		PRIMARY KEY (post_id, category_id)
	)`)
	if err != nil {
		t.Fatalf("Failed to create post_categories table: %v", err)
	}

	// Create reactions table (repository queries may reference it)
	_, err = db.Exec(`CREATE TABLE reactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		target_type TEXT,
		target_id INTEGER,
		type TEXT
	)`)
	if err != nil {
		t.Fatalf("Failed to create reactions table: %v", err)
	}

	// Create comments table (repository queries may reference it)
	_, err = db.Exec(`CREATE TABLE comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE NOT NULL,
		post_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME NOT NULL
	)`)
	if err != nil {
		t.Fatalf("Failed to create comments table: %v", err)
	}

	// Create users table (needed for author_id foreign key)
	_, err = db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE NOT NULL,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'user',
		post_count INTEGER NOT NULL DEFAULT 0,
		comment_count INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		is_active INTEGER DEFAULT 1,
		bio TEXT
	)`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	return db
}

func TestSQLitePostRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert a test user
	_, err := db.Exec("INSERT INTO users (public_id, username, email, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		"user-uuid-1", "testuser", "test@example.com", "hash", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// Insert a test category
	_, err = db.Exec("INSERT INTO categories (public_id, name, description, created_at) VALUES (?, ?, ?, ?)",
		"cat-uuid-1", "General", "General discussions", time.Now())
	if err != nil {
		t.Fatalf("Failed to insert test category: %v", err)
	}

	repo := NewSQLitePostRepository(db)

	now := time.Now()
	post := &domain.Post{
		UserID:     1,
		Title:      "Test Post",
		Content:    "Test content",
		CreatedAt:  now,
		UpdatedAt:  now,
		Categories: []string{"General"},
	}

	ctx := context.Background()
	err = repo.Create(ctx, post)
	if err != nil {
		t.Errorf("Create returned error: %v", err)
	}

	// Verify the post was created - check by public_id which should be set by Create
	if post.PublicID == "" {
		t.Error("PublicID was not set by Create")
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM posts WHERE title = ?", post.Title).Scan(&count)
	if err != nil {
		t.Errorf("Failed to query database: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 post, got %d", count)
	}
}

func TestSQLitePostRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert a test user
	_, err := db.Exec("INSERT INTO users (public_id, username, email, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		"user-uuid-1", "testuser", "test@example.com", "hash", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// Insert a test category
	result, err := db.Exec("INSERT INTO categories (public_id, name, description, created_at) VALUES (?, ?, ?, ?)",
		"cat-uuid-1", "General", "General discussions", time.Now())
	if err != nil {
		t.Fatalf("Failed to insert test category: %v", err)
	}
	catID, _ := result.LastInsertId()

	repo := NewSQLitePostRepository(db)

	// Insert a post directly for testing
	now := time.Now()
	postPublicID := "test-post-public-id"
	result, err = db.Exec("INSERT INTO posts (public_id, title, content, author_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		postPublicID, "Test Post", "Test content", 1, now, now)
	if err != nil {
		t.Fatalf("Failed to insert test post: %v", err)
	}
	postID, _ := result.LastInsertId()

	// Link category to post
	_, err = db.Exec("INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)", postID, catID)
	if err != nil {
		t.Fatalf("Failed to link category: %v", err)
	}

	ctx := context.Background()
	post, err := repo.GetByID(ctx, postPublicID)
	if err != nil {
		t.Errorf("GetByID returned error: %v", err)
	}

	if post == nil {
		t.Fatal("Expected a post, got nil")
	}

	if post.PublicID != postPublicID {
		t.Errorf("Expected PublicID %s, got %s", postPublicID, post.PublicID)
	}
	if post.Title != "Test Post" {
		t.Errorf("Expected Title 'Test Post', got '%s'", post.Title)
	}
	if post.Content != "Test content" {
		t.Errorf("Expected Content 'Test content', got '%s'", post.Content)
	}
	if post.UserID != 1 {
		t.Errorf("Expected UserID 1, got %d", post.UserID)
	}
}

func TestSQLitePostRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert a test user
	_, err := db.Exec("INSERT INTO users (public_id, username, email, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		"user-uuid-1", "testuser", "test@example.com", "hash", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// Insert test categories
	result, err := db.Exec("INSERT INTO categories (public_id, name, description, created_at) VALUES (?, ?, ?, ?)",
		"cat-uuid-1", "General", "General discussions", time.Now())
	if err != nil {
		t.Fatalf("Failed to insert test category: %v", err)
	}
	catID1, _ := result.LastInsertId()

	result, err = db.Exec("INSERT INTO categories (public_id, name, description, created_at) VALUES (?, ?, ?, ?)",
		"cat-uuid-2", "Technology", "Tech discussions", time.Now())
	if err != nil {
		t.Fatalf("Failed to insert test category: %v", err)
	}
	_, _ = result.LastInsertId()

	repo := NewSQLitePostRepository(db)

	// Insert a post directly for testing
	now := time.Now()
	postPublicID := "test-post-public-id"
	result, err = db.Exec("INSERT INTO posts (public_id, title, content, author_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		postPublicID, "Original Title", "Original content", 1, now, now)
	if err != nil {
		t.Fatalf("Failed to insert test post: %v", err)
	}
	postID, _ := result.LastInsertId()

	// Link first category
	_, err = db.Exec("INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)", postID, catID1)
	if err != nil {
		t.Fatalf("Failed to link category: %v", err)
	}

	// Prepare updated post
	updatedPost := &domain.Post{
		ID:         int(postID),
		PublicID:   postPublicID,
		UserID:     1,
		Title:      "Updated Title",
		Content:    "Updated content",
		CreatedAt:  now,
		UpdatedAt:  now.Add(1 * time.Hour),
		Categories: []string{"General", "Technology"},
	}

	ctx := context.Background()
	err = repo.Update(ctx, updatedPost)
	if err != nil {
		t.Errorf("Update returned error: %v", err)
	}

	// Verify the update in the database
	var title, content string
	err = db.QueryRow("SELECT title, content FROM posts WHERE id = ?", postID).Scan(&title, &content)
	if err != nil {
		t.Errorf("Failed to query updated post: %v", err)
	}

	if title != "Updated Title" {
		t.Errorf("Expected title 'Updated Title', got '%s'", title)
	}
	if content != "Updated content" {
		t.Errorf("Expected content 'Updated content', got '%s'", content)
	}

	// Verify categories were updated
	var catCount int
	err = db.QueryRow("SELECT COUNT(*) FROM post_categories WHERE post_id = ?", postID).Scan(&catCount)
	if err != nil {
		t.Errorf("Failed to query categories: %v", err)
	}
	if catCount != 2 {
		t.Errorf("Expected 2 categories, got %d", catCount)
	}
}

func TestSQLitePostRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert a test user
	_, err := db.Exec("INSERT INTO users (public_id, username, email, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		"user-uuid-1", "testuser", "test@example.com", "hash", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// Insert a test category
	result, err := db.Exec("INSERT INTO categories (public_id, name, description, created_at) VALUES (?, ?, ?, ?)",
		"cat-uuid-1", "General", "General discussions", time.Now())
	if err != nil {
		t.Fatalf("Failed to insert test category: %v", err)
	}
	catID, _ := result.LastInsertId()

	repo := NewSQLitePostRepository(db)

	// Insert a post directly for testing
	now := time.Now()
	postPublicID := "test-post-public-id"
	result, err = db.Exec("INSERT INTO posts (public_id, title, content, author_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		postPublicID, "Test Post", "Test content", 1, now, now)
	if err != nil {
		t.Fatalf("Failed to insert test post: %v", err)
	}
	postID, _ := result.LastInsertId()

	// Link category
	_, err = db.Exec("INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)", postID, catID)
	if err != nil {
		t.Fatalf("Failed to link category: %v", err)
	}

	ctx := context.Background()
	err = repo.Delete(ctx, postPublicID)
	if err != nil {
		t.Errorf("Delete returned error: %v", err)
	}

	// Verify the post was deleted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM posts WHERE public_id = ?", postPublicID).Scan(&count)
	if err != nil {
		t.Errorf("Failed to query database: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 posts, got %d", count)
	}
}

func TestSQLitePostRepository_List(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test users
	_, err := db.Exec("INSERT INTO users (public_id, username, email, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		"user-uuid-1", "testuser1", "test1@example.com", "hash", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	_, err = db.Exec("INSERT INTO users (public_id, username, email, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		"user-uuid-2", "testuser2", "test2@example.com", "hash", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// Insert test categories
	result, err := db.Exec("INSERT INTO categories (public_id, name, description, created_at) VALUES (?, ?, ?, ?)",
		"cat-uuid-1", "General", "General discussions", time.Now())
	if err != nil {
		t.Fatalf("Failed to insert test category: %v", err)
	}
	catID1, _ := result.LastInsertId()

	result, err = db.Exec("INSERT INTO categories (public_id, name, description, created_at) VALUES (?, ?, ?, ?)",
		"cat-uuid-2", "Technology", "Tech discussions", time.Now())
	if err != nil {
		t.Fatalf("Failed to insert test category: %v", err)
	}
	catID2, _ := result.LastInsertId()

	repo := NewSQLitePostRepository(db)

	// Insert test posts
	now := time.Now()
	result, err = db.Exec("INSERT INTO posts (public_id, title, content, author_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		"post-uuid-1", "Post 1", "Content 1", 1, now, now)
	if err != nil {
		t.Fatalf("Failed to insert test post: %v", err)
	}
	postID1, _ := result.LastInsertId()

	result, err = db.Exec("INSERT INTO posts (public_id, title, content, author_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		"post-uuid-2", "Post 2", "Content 2", 2, now, now)
	if err != nil {
		t.Fatalf("Failed to insert test post: %v", err)
	}
	postID2, _ := result.LastInsertId()

	// Link categories
	_, err = db.Exec("INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)", postID1, catID1)
	if err != nil {
		t.Fatalf("Failed to link category: %v", err)
	}
	_, err = db.Exec("INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)", postID2, catID2)
	if err != nil {
		t.Fatalf("Failed to link category: %v", err)
	}

	ctx := context.Background()

	// Test list all posts
	filter := ports.PostFilter{
		Limit:  10,
		Offset: 0,
	}
	posts, err := repo.List(ctx, filter)
	if err != nil {
		t.Errorf("List returned error: %v", err)
	}
	if len(posts) != 2 {
		t.Errorf("Expected 2 posts, got %d", len(posts))
	}

	// Test filter by user
	filter = ports.PostFilter{
		UserID: "user-uuid-1",
		Limit:  10,
		Offset: 0,
	}
	posts, err = repo.List(ctx, filter)
	if err != nil {
		t.Errorf("List returned error: %v", err)
	}
	if len(posts) != 1 {
		t.Errorf("Expected 1 post for user, got %d", len(posts))
	}

	// Test filter by category
	filter = ports.PostFilter{
		Categories: []string{"General"},
		Limit:      10,
		Offset:     0,
	}
	posts, err = repo.List(ctx, filter)
	if err != nil {
		t.Errorf("List returned error: %v", err)
	}
	if len(posts) != 1 {
		t.Errorf("Expected 1 post for category, got %d", len(posts))
	}
}
