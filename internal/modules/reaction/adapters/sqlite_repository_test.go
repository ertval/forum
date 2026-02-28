package adapters

import (
	"context"
	"database/sql"
	"fmt"
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

	// Create all required tables first
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

	// Create the reactions table
	_, err = db.Exec(`CREATE TABLE reactions (
		id INTEGER PRIMARY KEY,
		public_id TEXT UNIQUE NOT NULL,
		user_id INTEGER NOT NULL,
		target_id INTEGER NOT NULL,
		target_type TEXT NOT NULL,
		type TEXT NOT NULL,
		created_at DATETIME NOT NULL
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLiteReactionRepository(db)

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
		_, err = db.Exec("INSERT INTO reactions (public_id, user_id, target_id, target_type, type, created_at) VALUES (?, ?, ?, ?, ?, ?)",
			fmt.Sprintf("reaction-%d", reaction.ID), // Provide actual public_id to ensure NOT NULL constraint satisfied
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
	// With real implementation, expect actual count
	if likeCount != 2 { // 2 likes from reaction IDs 1 and 3
		t.Errorf("Expected 2 likes, got %d", likeCount)
	}

	// Count dislikes for post 10
	dislikeCount, err := repo.CountByTargetPublicID(ctx, "public-10", "post", domain.ReactionDislike)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if dislikeCount != 1 { // 1 dislike from reaction ID 2
		t.Errorf("Expected 1 dislike, got %d", dislikeCount)
	}
}

func TestSQLiteReactionRepository_GetByTarget(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create all required tables first
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

	// Create the reactions table
	_, err = db.Exec(`CREATE TABLE reactions (
		id INTEGER PRIMARY KEY,
		public_id TEXT UNIQUE NOT NULL,
		user_id INTEGER NOT NULL,
		target_id INTEGER NOT NULL,
		target_type TEXT NOT NULL,
		type TEXT NOT NULL,
		created_at DATETIME NOT NULL
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert mapping rows so public IDs resolve to internal IDs
	_, err = db.Exec("INSERT INTO posts (id, public_id) VALUES (?, ?)", 10, "public-10")
	if err != nil {
		t.Fatalf("Failed to insert posts mapping: %v", err)
	}
	_, err = db.Exec("INSERT INTO comments (id, public_id) VALUES (?, ?)", 15, "public-15")
	if err != nil {
		t.Fatalf("Failed to insert comments mapping: %v", err)
	}

	repo := NewSQLiteReactionRepository(db)

	// Insert test reactions directly for testing
	now := time.Now()
	reactions := []*domain.Reaction{
		{ID: 1, UserID: 1, PublicTargetID: "public-10", TargetID: 10, TargetType: "post", Type: domain.ReactionLike, CreatedAt: now},
		{ID: 2, UserID: 2, PublicTargetID: "public-10", TargetID: 10, TargetType: "post", Type: domain.ReactionDislike, CreatedAt: now},
		{ID: 3, UserID: 3, PublicTargetID: "public-15", TargetID: 15, TargetType: "comment", Type: domain.ReactionLike, CreatedAt: now}, // Different target
	}

	for _, reaction := range reactions {
		_, err = db.Exec("INSERT INTO reactions (public_id, user_id, target_id, target_type, type, created_at) VALUES (?, ?, ?, ?, ?, ?)",
			fmt.Sprintf("reaction-%d", reaction.ID), // Provide actual public_id to ensure NOT NULL constraint satisfied
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

	// With real implementation, expect actual reactions
	if len(result) != 2 { // 2 reactions for post with ID 10 (reactions 1 and 2)
		t.Errorf("Expected 2 reactions, got %d", len(result))
	}
}

func TestSQLiteReactionRepository_Delete(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create all required tables first
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

	// Create the reactions table
	_, err = db.Exec(`CREATE TABLE reactions (
		id INTEGER PRIMARY KEY,
		public_id TEXT UNIQUE NOT NULL,
		user_id INTEGER NOT NULL,
		target_id INTEGER NOT NULL,
		target_type TEXT NOT NULL,
		type TEXT NOT NULL,
		created_at DATETIME NOT NULL
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert mapping rows so public IDs resolve to internal IDs
	_, err = db.Exec("INSERT INTO posts (id, public_id) VALUES (?, ?)", 10, "public-10")
	if err != nil {
		t.Fatalf("Failed to insert posts mapping: %v", err)
	}

	repo := NewSQLiteReactionRepository(db)

	// Insert test reactions directly for testing
	now := time.Now()
	reactions := []*domain.Reaction{
		{ID: 1, UserID: 1, PublicTargetID: "public-10", TargetID: 10, TargetType: "post", Type: domain.ReactionLike, CreatedAt: now},
		{ID: 2, UserID: 2, PublicTargetID: "public-10", TargetID: 10, TargetType: "post", Type: domain.ReactionDislike, CreatedAt: now},
	}

	for _, reaction := range reactions {
		_, err = db.Exec("INSERT INTO reactions (public_id, user_id, target_id, target_type, type, created_at) VALUES (?, ?, ?, ?, ?, ?)",
			fmt.Sprintf("reaction-%d", reaction.ID), // Provide actual public_id to ensure NOT NULL constraint satisfied
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

	// Verify that the reaction was deleted
	remainingReactions, err := repo.GetByTargetPublicID(ctx, "public-10", "post")
	if err != nil {
		t.Errorf("GetByTarget after delete returned error: %v", err)
	}

	// Should have 1 reaction remaining (the one from user 2)
	if len(remainingReactions) != 1 {
		t.Errorf("Expected 1 remaining reaction after deletion, got %d", len(remainingReactions))
	}

	// Verify the remaining reaction is from user 2 (not user 1)
	if len(remainingReactions) > 0 && remainingReactions[0].UserID != 2 {
		t.Errorf("Expected remaining reaction to be from user 2, got user %d", remainingReactions[0].UserID)
	}
}
