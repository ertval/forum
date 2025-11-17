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
	
	// Count likes for post 10
	likeCount, err := repo.Count(ctx, 10, "post", domain.ReactionLike)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if likeCount != 2 {
		t.Errorf("Expected 2 likes, got %d", likeCount)
	}
	
	// Count dislikes for post 10
	dislikeCount, err := repo.Count(ctx, 10, "post", domain.ReactionDislike)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if dislikeCount != 1 {
		t.Errorf("Expected 1 dislike, got %d", dislikeCount)
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
	result, err := repo.GetByTarget(ctx, 10, "post")
	if err != nil {
		t.Errorf("GetByTarget returned error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 reactions for post 10, got %d", len(result))
	}

	// Verify all returned reactions belong to the correct target
	for _, reaction := range result {
		if reaction.TargetID != 10 || reaction.TargetType != "post" {
			t.Errorf("Expected TargetID 10 and TargetType 'post', got %d and %s", reaction.TargetID, reaction.TargetType)
		}
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
	err = repo.Delete(ctx, 1, 10, "post")
	if err != nil {
		t.Errorf("Delete returned error: %v", err)
	}

	// Verify the reaction was deleted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM reactions WHERE user_id = ? AND target_id = ? AND target_type = ?", 1, 10, "post").Scan(&count)
	if err != nil {
		t.Errorf("Failed to query deleted reaction: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 reactions after deletion, got %d", count)
	}
}