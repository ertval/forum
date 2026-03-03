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

func setupReactionTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE posts (
		id INTEGER PRIMARY KEY,
		public_id TEXT UNIQUE
	)`)
	if err != nil {
		db.Close()
		t.Fatalf("Failed to create posts table: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE comments (
		id INTEGER PRIMARY KEY,
		public_id TEXT UNIQUE
	)`)
	if err != nil {
		db.Close()
		t.Fatalf("Failed to create comments table: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE reactions (
		id INTEGER PRIMARY KEY,
		public_id TEXT UNIQUE NOT NULL,
		user_id INTEGER NOT NULL,
		target_id INTEGER NOT NULL,
		target_type TEXT NOT NULL,
		type TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		UNIQUE(user_id, target_id, target_type)
	)`)
	if err != nil {
		db.Close()
		t.Fatalf("Failed to create reactions table: %v", err)
	}

	_, err = db.Exec("INSERT INTO posts (id, public_id) VALUES (?, ?)", 10, "public-10")
	if err != nil {
		db.Close()
		t.Fatalf("Failed to insert post mapping: %v", err)
	}

	return db
}

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

func TestSQLiteReactionRepository_ToggleReaction_CreateUpdateDelete(t *testing.T) {
	db := setupReactionTestDB(t)
	defer db.Close()

	repo := NewSQLiteReactionRepository(db)
	ctx := context.Background()

	reaction := &domain.Reaction{
		UserID:         42,
		PublicTargetID: "public-10",
		TargetType:     "post",
		Type:           domain.ReactionLike,
		CreatedAt:      time.Now(),
	}

	action, err := repo.ToggleReaction(ctx, reaction)
	if err != nil {
		t.Fatalf("ToggleReaction(create) returned error: %v", err)
	}
	if action != domain.ToggleActionCreated {
		t.Fatalf("ToggleReaction(create) action = %q, want %q", action, domain.ToggleActionCreated)
	}

	var rowCount int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM reactions WHERE user_id = ? AND target_id = ? AND target_type = ?", 42, 10, "post").Scan(&rowCount)
	if err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if rowCount != 1 {
		t.Fatalf("after create count = %d, want 1", rowCount)
	}

	reaction.Type = domain.ReactionDislike
	action, err = repo.ToggleReaction(ctx, reaction)
	if err != nil {
		t.Fatalf("ToggleReaction(update) returned error: %v", err)
	}
	if action != domain.ToggleActionUpdated {
		t.Fatalf("ToggleReaction(update) action = %q, want %q", action, domain.ToggleActionUpdated)
	}

	var reactionType string
	err = db.QueryRowContext(ctx, "SELECT type FROM reactions WHERE user_id = ? AND target_id = ? AND target_type = ?", 42, 10, "post").Scan(&reactionType)
	if err != nil {
		t.Fatalf("type query failed: %v", err)
	}
	if reactionType != string(domain.ReactionDislike) {
		t.Fatalf("after update type = %q, want %q", reactionType, domain.ReactionDislike)
	}

	action, err = repo.ToggleReaction(ctx, reaction)
	if err != nil {
		t.Fatalf("ToggleReaction(delete) returned error: %v", err)
	}
	if action != domain.ToggleActionRemoved {
		t.Fatalf("ToggleReaction(delete) action = %q, want %q", action, domain.ToggleActionRemoved)
	}

	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM reactions WHERE user_id = ? AND target_id = ? AND target_type = ?", 42, 10, "post").Scan(&rowCount)
	if err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if rowCount != 0 {
		t.Fatalf("after delete count = %d, want 0", rowCount)
	}
}

func TestSQLiteReactionRepository_DeleteByTargetPublicID_TargetNotFound(t *testing.T) {
	db := setupReactionTestDB(t)
	defer db.Close()

	repo := NewSQLiteReactionRepository(db)
	ctx := context.Background()

	err := repo.DeleteByTargetPublicID(ctx, 1, "missing-target", "post")
	if err != domain.ErrTargetNotFound {
		t.Fatalf("expected ErrTargetNotFound, got %v", err)
	}
}

func TestSQLiteReactionRepository_DeleteByTargetPublicID_ReactionNotFound(t *testing.T) {
	db := setupReactionTestDB(t)
	defer db.Close()

	repo := NewSQLiteReactionRepository(db)
	ctx := context.Background()

	err := repo.DeleteByTargetPublicID(ctx, 999, "public-10", "post")
	if err != domain.ErrReactionNotFound {
		t.Fatalf("expected ErrReactionNotFound, got %v", err)
	}
}

func TestSQLiteReactionRepository_ListByUserID_IncludesPostAndCommentTargets(t *testing.T) {
	db := setupReactionTestDB(t)
	defer db.Close()

	_, err := db.Exec("INSERT INTO comments (id, public_id) VALUES (?, ?)", 15, "comment-15")
	if err != nil {
		t.Fatalf("Failed to insert comment mapping: %v", err)
	}

	// Newer reaction on comment, older on post
	_, err = db.Exec(`
		INSERT INTO reactions (public_id, user_id, target_id, target_type, type, created_at)
		VALUES
			('rxn-post', 7, 10, 'post', 'like', '2026-03-01 10:00:00'),
			('rxn-comment', 7, 15, 'comment', 'dislike', '2026-03-02 10:00:00')
	`)
	if err != nil {
		t.Fatalf("Failed to seed reactions: %v", err)
	}

	repo := NewSQLiteReactionRepository(db)
	ctx := context.Background()

	reactions, err := repo.ListByUserID(ctx, 7)
	if err != nil {
		t.Fatalf("ListByUserID returned error: %v", err)
	}

	if len(reactions) != 2 {
		t.Fatalf("expected 2 reactions, got %d", len(reactions))
	}

	if reactions[0].TargetType != "comment" || reactions[0].PublicTargetID != "comment-15" {
		t.Fatalf("expected newest comment reaction first with public target comment-15, got targetType=%q publicTarget=%q", reactions[0].TargetType, reactions[0].PublicTargetID)
	}

	if reactions[1].TargetType != "post" || reactions[1].PublicTargetID != "public-10" {
		t.Fatalf("expected post reaction second with public target public-10, got targetType=%q publicTarget=%q", reactions[1].TargetType, reactions[1].PublicTargetID)
	}
}
