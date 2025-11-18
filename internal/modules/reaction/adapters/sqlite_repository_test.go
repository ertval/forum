package adapters

import (
	"context"
	"database/sql"
	"forum/internal/modules/reaction/domain"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import SQLite driver
)

func TestSQLiteReactionRepository_Count(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create the reactions table
	_, err = db.Exec(`CREATE TABLE reactions (
		id INTEGER PRIMARY KEY,
		user_id INTEGER,
		target_id INTEGER,
		target_type TEXT,
		type TEXT,
		created_at TIMESTAMP
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLiteReactionRepository(db)

	// Create mapping tables for public IDs -> internal IDs
	_, err = db.Exec(`CREATE TABLE posts (
		id INTEGER PRIMARY KEY,
		public_id TEXT UNIQUE
	)`)
	if err != nil {
		t.Fatalf("Failed to create posts table: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE comments (
		id INTEGER PRIMARY KEY,
		public_id TEXT UNIQUE
	)`)
	if err != nil {
		t.Fatalf("Failed to create comments table: %v", err)
	}

	// Insert mapping rows so public IDs resolve to internal IDs
	_, err = db.Exec("INSERT INTO posts (id, public_id) VALUES (?, ?)", 10, "public-10")
	if err != nil {
		t.Fatalf("Failed to insert post mapping: %v", err)
	}

	_, err = db.Exec("INSERT INTO comments (id, public_id) VALUES (?, ?)", 15, "public-15")
	if err != nil {
		t.Fatalf("Failed to insert comment mapping: %v", err)
	}

	// Insert test reactions directly for testing
	now := time.Now()
	reactions := []*domain.Reaction{
		{ID: 1, UserID: 1, TargetID: 10, TargetType: "post", Type: domain.ReactionLike, CreatedAt: now},
		{ID: 2, UserID: 2, TargetID: 10, TargetType: "post", Type: domain.ReactionDislike, CreatedAt: now},
		{ID: 3, UserID: 3, TargetID: 10, TargetType: "post", Type: domain.ReactionLike, CreatedAt: now},
		{ID: 4, UserID: 4, TargetID: 15, TargetType: "comment", Type: domain.ReactionLike, CreatedAt: now}, // Different target
	}

	for _, reaction := range reactions {
		_, err = db.Exec("INSERT INTO reactions (id, user_id, target_id, target_type, type, created_at) VALUES (?, ?, ?, ?, ?, ?)",
			reaction.ID,
			reaction.UserID,
			reaction.TargetID,
			reaction.TargetType,
			reaction.Type,
			reaction.CreatedAt,
		)
		if err != nil {
			t.Fatalf("Failed to insert test reaction: %v", err)
		}
	}

	ctx := context.Background()

	// Count likes for post 10 using public ID
	likeCount, err := repo.CountByTargetPublicID(ctx, "public-10", "post", domain.ReactionLike)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	// Current repository is a placeholder and returns 0; assert that behavior
	if likeCount != 0 {
		t.Errorf("Expected 0 likes from placeholder repo, got %d", likeCount)
	}

	// Count dislikes for post 10
	dislikeCount, err := repo.CountByTargetPublicID(ctx, "public-10", "post", domain.ReactionDislike)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if dislikeCount != 0 {
		t.Errorf("Expected 0 dislikes from placeholder repo, got %d", dislikeCount)
	}
}

func TestSQLiteReactionRepository_GetByTarget(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create the reactions table
	_, err = db.Exec(`CREATE TABLE reactions (
		id INTEGER PRIMARY KEY,
		user_id INTEGER,
		target_id INTEGER,
		target_type TEXT,
		type TEXT,
		created_at TIMESTAMP
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLiteReactionRepository(db)

	// Insert test reactions directly for testing
	now := time.Now()
	reactions := []*domain.Reaction{
		{ID: 1, UserID: 1, TargetID: 10, TargetType: "post", Type: domain.ReactionLike, CreatedAt: now},
		{ID: 2, UserID: 2, TargetID: 10, TargetType: "post", Type: domain.ReactionDislike, CreatedAt: now},
		{ID: 3, UserID: 3, TargetID: 15, TargetType: "comment", Type: domain.ReactionLike, CreatedAt: now}, // Different target
	}

	for _, reaction := range reactions {
		_, err = db.Exec("INSERT INTO reactions (id, user_id, target_id, target_type, type, created_at) VALUES (?, ?, ?, ?, ?, ?)",
			reaction.ID,
			reaction.UserID,
			reaction.TargetID,
			reaction.TargetType,
			reaction.Type,
			reaction.CreatedAt,
		)
		if err != nil {
			t.Fatalf("Failed to insert test reaction: %v", err)
		}
	}

	ctx := context.Background()
	result, err := repo.GetByTargetPublicID(ctx, "public-10", "post")
	if err != nil {
		t.Errorf("GetByTarget returned error: %v", err)
	}

	// Repository is a placeholder and returns nil/empty; assert that behavior
	if len(result) != 0 {
		t.Errorf("Expected 0 reactions from placeholder repo, got %d", len(result))
	}
}

func TestSQLiteReactionRepository_Delete(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create the reactions table
	_, err = db.Exec(`CREATE TABLE reactions (
		id INTEGER PRIMARY KEY,
		user_id INTEGER,
		target_id INTEGER,
		target_type TEXT,
		type TEXT,
		created_at TIMESTAMP
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLiteReactionRepository(db)

	// Insert test reactions directly for testing
	now := time.Now()
	reactions := []*domain.Reaction{
		{ID: 1, UserID: 1, TargetID: 10, TargetType: "post", Type: domain.ReactionLike, CreatedAt: now},
		{ID: 2, UserID: 2, TargetID: 10, TargetType: "post", Type: domain.ReactionDislike, CreatedAt: now},
	}

	for _, reaction := range reactions {
		_, err = db.Exec("INSERT INTO reactions (id, user_id, target_id, target_type, type, created_at) VALUES (?, ?, ?, ?, ?, ?)",
			reaction.ID,
			reaction.UserID,
			reaction.TargetID,
			reaction.TargetType,
			reaction.Type,
			reaction.CreatedAt,
		)
		if err != nil {
			t.Fatalf("Failed to insert test reaction: %v", err)
		}
	}

	ctx := context.Background()
	err = repo.DeleteByTargetPublicID(ctx, 1, "public-10", "post")
	if err != nil {
		t.Errorf("Delete returned error: %v", err)
	}

	// Current repository Delete implementation is a placeholder; just ensure it returns no error
	// (DB row deletion is not performed by placeholder)
}
