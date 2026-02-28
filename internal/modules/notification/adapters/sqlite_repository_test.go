package adapters

import (
	"context"
	"database/sql"
	"forum/internal/modules/notification/domain"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func setupNotificationTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			public_id TEXT NOT NULL UNIQUE,
			username TEXT,
			email TEXT,
			password_hash TEXT,
			role TEXT,
			is_active BOOLEAN,
			created_at DATETIME,
			updated_at DATETIME,
			post_count INTEGER,
			comment_count INTEGER
		);
		CREATE TABLE posts (
			id INTEGER PRIMARY KEY,
			public_id TEXT NOT NULL UNIQUE,
			title TEXT,
			content TEXT,
			author_id INTEGER,
			created_at DATETIME,
			updated_at DATETIME
		);
		CREATE TABLE notifications (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			public_id TEXT UNIQUE NOT NULL,
			user_id INTEGER NOT NULL,
			actor_id INTEGER NOT NULL,
			target_id INTEGER NOT NULL,
			type TEXT NOT NULL,
			message TEXT NOT NULL,
			read BOOLEAN NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL
		);
	`)
	if err != nil {
		db.Close()
		t.Fatalf("Failed to create schema: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO users (id, public_id, username, email, password_hash, role, is_active, created_at, updated_at, post_count, comment_count)
		VALUES
		(1, 'user-1', 'u1', 'u1@example.com', 'hash', 'user', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 0, 0),
		(2, 'user-2', 'u2', 'u2@example.com', 'hash', 'user', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 0, 0);
		INSERT INTO posts (id, public_id, title, content, author_id, created_at, updated_at)
		VALUES (10, 'post-10', 'T', 'C', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
	`)
	if err != nil {
		db.Close()
		t.Fatalf("Failed to seed schema: %v", err)
	}

	return db
}

func TestSQLiteNotificationRepository_Create(t *testing.T) {
	db := setupNotificationTestDB(t)
	defer db.Close()

	repo := NewSQLiteNotificationRepository(db)
	notification := &domain.Notification{
		UserID:         1,
		ActorID:        2,
		Type:           domain.TypeLike,
		Message:        "Someone liked your post",
		PublicTargetID: "post-10",
		CreatedAt:      time.Now(),
	}

	err := repo.Create(context.Background(), notification)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if notification.PublicID == "" {
		t.Fatal("expected generated PublicID")
	}
	if notification.TargetID != 10 {
		t.Fatalf("expected target ID 10, got %d", notification.TargetID)
	}
}

func TestSQLiteNotificationRepository_GetByUserID(t *testing.T) {
	db := setupNotificationTestDB(t)
	defer db.Close()

	repo := NewSQLiteNotificationRepository(db)
	err := repo.Create(context.Background(), &domain.Notification{
		UserID:         1,
		ActorID:        2,
		Type:           domain.TypeComment,
		Message:        "Someone commented on your post",
		PublicTargetID: "post-10",
		CreatedAt:      time.Now(),
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	result, err := repo.GetByUserID(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetByUserID returned error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(result))
	}
	if result[0].PublicTargetID != "post-10" {
		t.Fatalf("expected target public id post-10, got %s", result[0].PublicTargetID)
	}
	if result[0].PublicActorID != "user-2" {
		t.Fatalf("expected actor public id user-2, got %s", result[0].PublicActorID)
	}
}

func TestSQLiteNotificationRepository_MarkAsRead(t *testing.T) {
	db := setupNotificationTestDB(t)
	defer db.Close()

	repo := NewSQLiteNotificationRepository(db)
	notification := &domain.Notification{
		UserID:         1,
		ActorID:        2,
		Type:           domain.TypeDislike,
		Message:        "Someone disliked your post",
		PublicTargetID: "post-10",
		CreatedAt:      time.Now(),
	}
	if err := repo.Create(context.Background(), notification); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if err := repo.MarkAsReadByPublicID(context.Background(), notification.PublicID); err != nil {
		t.Fatalf("MarkAsReadByPublicID returned error: %v", err)
	}

	result, err := repo.GetByUserID(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetByUserID returned error: %v", err)
	}
	if len(result) != 1 || !result[0].IsRead {
		t.Fatalf("expected notification to be marked as read")
	}
}
