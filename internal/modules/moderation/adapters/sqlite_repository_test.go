package adapters

import (
	"context"
	"database/sql"
	"forum/internal/modules/moderation/domain"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import SQLite driver
)

func setupModerationTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			public_id TEXT UNIQUE NOT NULL,
			email TEXT,
			username TEXT,
			password_hash TEXT,
			avatar_path TEXT,
			role TEXT,
			post_count INTEGER DEFAULT 0,
			comment_count INTEGER DEFAULT 0,
			reaction_count INTEGER DEFAULT 0,
			created_at DATETIME,
			updated_at DATETIME,
			is_active INTEGER DEFAULT 1
		);

		CREATE TABLE posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			public_id TEXT UNIQUE NOT NULL,
			title TEXT,
			content TEXT,
			author_id INTEGER,
			image_path TEXT,
			created_at DATETIME,
			updated_at DATETIME
		);

		CREATE TABLE comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			public_id TEXT UNIQUE NOT NULL,
			post_id INTEGER,
			author_id INTEGER,
			content TEXT,
			created_at DATETIME,
			updated_at DATETIME
		);

		CREATE TABLE reports (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			public_id TEXT UNIQUE NOT NULL,
			reporter_id INTEGER NOT NULL,
			moderator_id INTEGER,
			target_id INTEGER NOT NULL,
			target_type TEXT NOT NULL,
			reason TEXT NOT NULL,
			status TEXT NOT NULL,
			response TEXT,
			created_at DATETIME NOT NULL,
			reviewed_at DATETIME
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return db
}

func seedModerationTestData(t *testing.T, db *sql.DB) {
	now := time.Now()

	_, err := db.Exec(`
		INSERT INTO users (id, public_id, email, username, password_hash, role, created_at, updated_at, is_active)
		VALUES
			(1, 'user-reporter-public', 'r@example.com', 'reporter', 'x', 'user', ?, ?, 1),
			(2, 'user-moderator-public', 'm@example.com', 'moderator', 'x', 'moderator', ?, ?, 1),
			(3, 'user-author-public', 'a@example.com', 'author', 'x', 'user', ?, ?, 1)
	`, now, now, now, now, now, now)
	if err != nil {
		t.Fatalf("failed seeding users: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO posts (id, public_id, title, content, author_id, created_at, updated_at)
		VALUES (10, 'post-public-id', 'title', 'content', 3, ?, ?)
	`, now, now)
	if err != nil {
		t.Fatalf("failed seeding posts: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO comments (id, public_id, post_id, author_id, content, created_at, updated_at)
		VALUES (20, 'comment-public-id', 10, 3, 'comment', ?, ?)
	`, now, now)
	if err != nil {
		t.Fatalf("failed seeding comments: %v", err)
	}
}

func TestSQLiteReportRepository_CreateAndGetByPublicID(t *testing.T) {
	db := setupModerationTestDB(t)
	defer db.Close()
	seedModerationTestData(t, db)

	repo := NewSQLiteReportRepository(db)

	now := time.Now()
	report := &domain.Report{
		ReporterID: 1,
		TargetID:   10,
		TargetType: "post",
		Reason:     "Inappropriate content",
		Status:     domain.StatusPending,
		CreatedAt:  now,
	}

	ctx := context.Background()
	err := repo.Create(ctx, report)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if report.PublicID == "" {
		t.Fatal("expected PublicID to be generated")
	}

	stored, err := repo.GetByPublicID(ctx, report.PublicID)
	if err != nil {
		t.Fatalf("GetByPublicID() error = %v", err)
	}
	if stored.TargetType != "post" {
		t.Fatalf("target_type = %q, want post", stored.TargetType)
	}
	if stored.PublicReporterID != "user-reporter-public" {
		t.Fatalf("reporter_id = %q, want user-reporter-public", stored.PublicReporterID)
	}
	if stored.PublicTargetID != "post-public-id" {
		t.Fatalf("target_id = %q, want post-public-id", stored.PublicTargetID)
	}
}

func TestSQLiteReportRepository_ListAndUpdate(t *testing.T) {
	db := setupModerationTestDB(t)
	defer db.Close()
	seedModerationTestData(t, db)

	now := time.Now()
	_, err := db.Exec(`
		INSERT INTO reports (id, public_id, reporter_id, moderator_id, target_id, target_type, reason, status, response, created_at, reviewed_at)
		VALUES
			(1, 'report-1', 1, NULL, 10, 'post', 'spam', 'pending', NULL, ?, NULL),
			(2, 'report-2', 1, 2, 20, 'comment', 'abuse', 'reviewed', 'warning', ?, ?)
	`, now, now, now)
	if err != nil {
		t.Fatalf("failed seeding reports: %v", err)
	}

	repo := NewSQLiteReportRepository(db)
	ctx := context.Background()

	pending, err := repo.List(ctx, domain.StatusPending)
	if err != nil {
		t.Fatalf("List(pending) error = %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("pending reports len = %d, want 1", len(pending))
	}

	reviewed, err := repo.GetByPublicID(ctx, "report-2")
	if err != nil {
		t.Fatalf("GetByPublicID(report-2) error = %v", err)
	}
	if reviewed.PublicModeratorID != "user-moderator-public" {
		t.Fatalf("moderator_id = %q, want user-moderator-public", reviewed.PublicModeratorID)
	}
	if reviewed.PublicTargetID != "comment-public-id" {
		t.Fatalf("target_id = %q, want comment-public-id", reviewed.PublicTargetID)
	}

	moderatorID := 2
	reviewed.Status = domain.StatusResolved
	reviewed.Response = "resolved"
	reviewed.ModeratorID = &moderatorID
	updatedAt := time.Now()
	reviewed.ReviewedAt = &updatedAt

	err = repo.Update(ctx, reviewed)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	updated, err := repo.GetByPublicID(ctx, "report-2")
	if err != nil {
		t.Fatalf("GetByPublicID(report-2) after update error = %v", err)
	}
	if updated.Status != domain.StatusResolved {
		t.Fatalf("status = %q, want %q", updated.Status, domain.StatusResolved)
	}
	if updated.Response != "resolved" {
		t.Fatalf("response = %q, want resolved", updated.Response)
	}
}

func TestSQLiteReportRepository_ResolveTargetID(t *testing.T) {
	db := setupModerationTestDB(t)
	defer db.Close()
	seedModerationTestData(t, db)

	repo := NewSQLiteReportRepository(db)
	ctx := context.Background()

	postID, err := repo.ResolveTargetID(ctx, "post", "post-public-id")
	if err != nil {
		t.Fatalf("ResolveTargetID(post) error = %v", err)
	}
	if postID != 10 {
		t.Fatalf("post target id = %d, want 10", postID)
	}

	commentID, err := repo.ResolveTargetID(ctx, "comment", "comment-public-id")
	if err != nil {
		t.Fatalf("ResolveTargetID(comment) error = %v", err)
	}
	if commentID != 20 {
		t.Fatalf("comment target id = %d, want 20", commentID)
	}
	_, err = repo.ResolveTargetID(ctx, "post", "missing")
	if err != domain.ErrInvalidTarget {
		t.Fatalf("ResolveTargetID(missing) error = %v, want %v", err, domain.ErrInvalidTarget)
	}

	_, err = repo.ResolveTargetID(ctx, "user", "whatever")
	if err != domain.ErrInvalidTargetType {
		t.Fatalf("ResolveTargetID(invalid type) error = %v, want %v", err, domain.ErrInvalidTargetType)
	}
}
