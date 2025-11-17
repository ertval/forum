package adapters

import (
	"context"
	"database/sql"
	"forum/internal/modules/post/domain"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import SQLite driver
)

func TestSQLitePostRepository_Create(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create the posts table
	_, err = db.Exec(`CREATE TABLE posts (
		id TEXT PRIMARY KEY,
		user_id INTEGER,
		title TEXT,
		content TEXT,
		image_url TEXT,
		created_at TIMESTAMP,
		updated_at TIMESTAMP
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Create the post_categories table
	_, err = db.Exec(`CREATE TABLE post_categories (
		post_id TEXT,
		category_name TEXT,
		PRIMARY KEY (post_id, category_name)
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLitePostRepository(db)

	now := time.Now()
	post := &domain.Post{
		ID:        "test-post-id",
		UserID:    1,
		Title:     "Test Post",
		Content:   "Test content",
		CreatedAt: now,
		UpdatedAt: now,
		Categories: []string{"General"},
	}

	ctx := context.Background()
	err = repo.Create(ctx, post)
	if err != nil {
		t.Errorf("Create returned error: %v", err)
	}

	// Verify the post was created
	var id string
	err = db.QueryRow("SELECT id FROM posts WHERE title = ?", post.Title).Scan(&id)
	if err != nil {
		t.Errorf("Post was not created in database: %v", err)
	}
	if id != post.ID {
		t.Errorf("Expected ID %s, got %s", post.ID, id)
	}
}

func TestSQLitePostRepository_Get(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create the posts table
	_, err = db.Exec(`CREATE TABLE posts (
		id TEXT PRIMARY KEY,
		user_id INTEGER,
		title TEXT,
		content TEXT,
		image_url TEXT,
		created_at TIMESTAMP,
		updated_at TIMESTAMP
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Create the post_categories table
	_, err = db.Exec(`CREATE TABLE post_categories (
		post_id TEXT,
		category_name TEXT,
		PRIMARY KEY (post_id, category_name)
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLitePostRepository(db)

	// Insert a post directly for testing
	now := time.Now()
	post := &domain.Post{
		ID:        "test-post-id",
		UserID:    1,
		Title:     "Test Post",
		Content:   "Test content",
		CreatedAt: now,
		UpdatedAt: now,
		Categories: []string{"General"},
	}

	_, err = db.Exec("INSERT INTO posts (id, user_id, title, content, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		post.ID,
		post.UserID,
		post.Title,
		post.Content,
		post.CreatedAt,
		post.UpdatedAt,
	)
	if err != nil {
		t.Fatalf("Failed to insert test post: %v", err)
	}

	// Insert categories
	for _, category := range post.Categories {
		_, err = db.Exec("INSERT INTO post_categories (post_id, category_name) VALUES (?, ?)",
			post.ID, category)
		if err != nil {
			t.Fatalf("Failed to insert test category: %v", err)
		}
	}

	ctx := context.Background()
	result, err := repo.Get(ctx, post.ID)
	if err != nil {
		t.Errorf("Get returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected a post, got nil")
	}

	if result.ID != post.ID {
		t.Errorf("Expected ID %s, got %s", post.ID, result.ID)
	}
	if result.Title != post.Title {
		t.Errorf("Expected Title %s, got %s", post.Title, result.Title)
	}
	if result.Content != post.Content {
		t.Errorf("Expected Content %s, got %s", post.Content, result.Content)
	}
	if result.UserID != post.UserID {
		t.Errorf("Expected UserID %d, got %d", post.UserID, result.UserID)
	}
}

func TestSQLitePostRepository_Update(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create the posts table
	_, err = db.Exec(`CREATE TABLE posts (
		id TEXT PRIMARY KEY,
		user_id INTEGER,
		title TEXT,
		content TEXT,
		image_url TEXT,
		created_at TIMESTAMP,
		updated_at TIMESTAMP
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Create the post_categories table
	_, err = db.Exec(`CREATE TABLE post_categories (
		post_id TEXT,
		category_name TEXT,
		PRIMARY KEY (post_id, category_name)
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLitePostRepository(db)

	// Insert a post directly for testing
	now := time.Now()
	post := &domain.Post{
		ID:        "test-post-id",
		UserID:    1,
		Title:     "Original Title",
		Content:   "Original content",
		CreatedAt: now,
		UpdatedAt: now,
		Categories: []string{"General"},
	}

	_, err = db.Exec("INSERT INTO posts (id, user_id, title, content, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		post.ID,
		post.UserID,
		post.Title,
		post.Content,
		post.CreatedAt,
		post.UpdatedAt,
	)
	if err != nil {
		t.Fatalf("Failed to insert test post: %v", err)
	}

	// Insert categories
	for _, category := range post.Categories {
		_, err = db.Exec("INSERT INTO post_categories (post_id, category_name) VALUES (?, ?)",
			post.ID, category)
		if err != nil {
			t.Fatalf("Failed to insert test category: %v", err)
		}
	}

	// Prepare updated post
	updatedPost := &domain.Post{
		ID:        "test-post-id",
		UserID:    1,
		Title:     "Updated Title",
		Content:   "Updated content",
		CreatedAt: now,
		UpdatedAt: now.Add(1 * time.Hour), // Updated time
		Categories: []string{"General", "Technology"}, // New categories
	}

	ctx := context.Background()
	err = repo.Update(ctx, updatedPost)
	if err != nil {
		t.Errorf("Update returned error: %v", err)
	}

	// Verify the update in the database
	var title, content string
	err = db.QueryRow("SELECT title, content FROM posts WHERE id = ?", updatedPost.ID).Scan(&title, &content)
	if err != nil {
		t.Errorf("Failed to query updated post: %v", err)
	}

	if title != "Updated Title" {
		t.Errorf("Expected title 'Updated Title', got '%s'", title)
	}
	if content != "Updated content" {
		t.Errorf("Expected content 'Updated content', got '%s'", content)
	}
}

func TestSQLitePostRepository_Delete(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create the posts table
	_, err = db.Exec(`CREATE TABLE posts (
		id TEXT PRIMARY KEY,
		user_id INTEGER,
		title TEXT,
		content TEXT,
		image_url TEXT,
		created_at TIMESTAMP,
		updated_at TIMESTAMP
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Create the post_categories table
	_, err = db.Exec(`CREATE TABLE post_categories (
		post_id TEXT,
		category_name TEXT,
		PRIMARY KEY (post_id, category_name)
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLitePostRepository(db)

	// Insert a post directly for testing
	now := time.Now()
	post := &domain.Post{
		ID:        "test-post-id",
		UserID:    1,
		Title:     "Test Post",
		Content:   "Test content",
		CreatedAt: now,
		UpdatedAt: now,
		Categories: []string{"General"},
	}

	_, err = db.Exec("INSERT INTO posts (id, user_id, title, content, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		post.ID,
		post.UserID,
		post.Title,
		post.Content,
		post.CreatedAt,
		post.UpdatedAt,
	)
	if err != nil {
		t.Fatalf("Failed to insert test post: %v", err)
	}

	// Insert categories
	for _, category := range post.Categories {
		_, err = db.Exec("INSERT INTO post_categories (post_id, category_name) VALUES (?, ?)",
			post.ID, category)
		if err != nil {
			t.Fatalf("Failed to insert test category: %v", err)
		}
	}

	ctx := context.Background()
	err = repo.Delete(ctx, post.ID, post.UserID)
	if err != nil {
		t.Errorf("Delete returned error: %v", err)
	}

	// Verify the post was deleted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM posts WHERE id = ?", post.ID).Scan(&count)
	if err != nil {
		t.Errorf("Failed to query deleted post: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 posts after deletion, got %d", count)
	}
}

func TestSQLitePostRepository_List(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create the posts table
	_, err = db.Exec(`CREATE TABLE posts (
		id TEXT PRIMARY KEY,
		user_id INTEGER,
		title TEXT,
		content TEXT,
		image_url TEXT,
		created_at TIMESTAMP,
		updated_at TIMESTAMP
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Create the post_categories table
	_, err = db.Exec(`CREATE TABLE post_categories (
		post_id TEXT,
		category_name TEXT,
		PRIMARY KEY (post_id, category_name)
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLitePostRepository(db)

	// Insert test posts directly for testing
	now := time.Now()
	posts := []*domain.Post{
		{ID: "post-1", UserID: 1, Title: "First Post", Content: "Content 1", CreatedAt: now, UpdatedAt: now, Categories: []string{"General"}},
		{ID: "post-2", UserID: 1, Title: "Second Post", Content: "Content 2", CreatedAt: now, UpdatedAt: now, Categories: []string{"Technology"}},
		{ID: "post-3", UserID: 2, Title: "Third Post", Content: "Content 3", CreatedAt: now, UpdatedAt: now, Categories: []string{"General"}},
	}

	for _, post := range posts {
		_, err = db.Exec("INSERT INTO posts (id, user_id, title, content, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
			post.ID,
			post.UserID,
			post.Title,
			post.Content,
			post.CreatedAt,
			post.UpdatedAt,
		)
		if err != nil {
			t.Fatalf("Failed to insert test post: %v", err)
		}

		// Insert categories
		for _, category := range post.Categories {
			_, err = db.Exec("INSERT INTO post_categories (post_id, category_name) VALUES (?, ?)",
				post.ID, category)
			if err != nil {
				t.Fatalf("Failed to insert test category: %v", err)
			}
		}
	}

	ctx := context.Background()
	result, err := repo.List(ctx, PostFilters{})
	if err != nil {
		t.Errorf("List returned error: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 posts, got %d", len(result))
	}
}

func TestSQLitePostRepository_GetUserPosts(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create the posts table
	_, err = db.Exec(`CREATE TABLE posts (
		id TEXT PRIMARY KEY,
		user_id INTEGER,
		title TEXT,
		content TEXT,
		image_url TEXT,
		created_at TIMESTAMP,
		updated_at TIMESTAMP
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Create the post_categories table
	_, err = db.Exec(`CREATE TABLE post_categories (
		post_id TEXT,
		category_name TEXT,
		PRIMARY KEY (post_id, category_name)
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLitePostRepository(db)

	// Insert test posts directly for testing
	now := time.Now()
	posts := []*domain.Post{
		{ID: "post-1", UserID: 1, Title: "First Post", Content: "Content 1", CreatedAt: now, UpdatedAt: now, Categories: []string{"General"}},
		{ID: "post-2", UserID: 1, Title: "Second Post", Content: "Content 2", CreatedAt: now, UpdatedAt: now, Categories: []string{"Technology"}},
		{ID: "post-3", UserID: 2, Title: "Third Post", Content: "Content 3", CreatedAt: now, UpdatedAt: now, Categories: []string{"General"}},
	}

	for _, post := range posts {
		_, err = db.Exec("INSERT INTO posts (id, user_id, title, content, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
			post.ID,
			post.UserID,
			post.Title,
			post.Content,
			post.CreatedAt,
			post.UpdatedAt,
		)
		if err != nil {
			t.Fatalf("Failed to insert test post: %v", err)
		}

		// Insert categories
		for _, category := range post.Categories {
			_, err = db.Exec("INSERT INTO post_categories (post_id, category_name) VALUES (?, ?)",
				post.ID, category)
			if err != nil {
				t.Fatalf("Failed to insert test category: %v", err)
			}
		}
	}

	ctx := context.Background()
	result, err := repo.GetUserPosts(ctx, 1)
	if err != nil {
		t.Errorf("GetUserPosts returned error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 posts for user 1, got %d", len(result))
	}

	// Verify all returned posts belong to the correct user
	for _, post := range result {
		if post.UserID != 1 {
			t.Errorf("Expected UserID 1, got %d", post.UserID)
		}
	}
}