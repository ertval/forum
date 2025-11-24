// OUTPUT ADAPTER - SQLite Repository
// Package adapters implements the SQLite repository for comments.
package adapters

import (
	"context"
	"database/sql"
	"forum/internal/modules/comment/domain"
	"forum/internal/modules/comment/ports"
)

// SQLiteCommentRepository implements the CommentRepository interface using SQLite.
type SQLiteCommentRepository struct {
	db *sql.DB
}

// NewSQLiteCommentRepository creates a new SQLite comment repository.
func NewSQLiteCommentRepository(db *sql.DB) ports.CommentRepository {
	return &SQLiteCommentRepository{db: db}
}

// Create stores a new comment in the database.
// TODO: Implement comment creation with UUID generation.
func (r *SQLiteCommentRepository) Create(ctx context.Context, comment *domain.Comment) error {
	// Implementation placeholder
	// 1. Generate UUID for PublicID
	// 2. INSERT INTO comments (public_id, post_id, user_id, content, created_at)
	//    VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
	return nil
}

// GetByPublicID retrieves a comment by its public UUID.
// TODO: Implement comment retrieval by public UUID.
func (r *SQLiteCommentRepository) GetByPublicID(ctx context.Context, commentPublicID string) (*domain.Comment, error) {
	// Implementation placeholder
	// SELECT id, public_id, post_id, user_id, content, created_at, updated_at
	// FROM comments WHERE public_id = ?
	return nil, nil
}

// Update updates an existing comment.
// TODO: Implement comment update.
func (r *SQLiteCommentRepository) Update(ctx context.Context, comment *domain.Comment) error {
	// Implementation placeholder
	// UPDATE comments SET content=?, updated_at=CURRENT_TIMESTAMP WHERE id=?
	return nil
}

// DeleteByPublicID removes a comment by its public UUID.
// TODO: Implement comment deletion by public UUID.
func (r *SQLiteCommentRepository) DeleteByPublicID(ctx context.Context, commentPublicID string) error {
	// Implementation placeholder
	// DELETE FROM comments WHERE public_id = ?
	return nil
}

// ListByPostPublicID retrieves all comments for a post by the post's public UUID.
// TODO: Implement comment listing by post public UUID.
func (r *SQLiteCommentRepository) ListByPostPublicID(ctx context.Context, postPublicID string) ([]*domain.Comment, error) {
	// Implementation placeholder
	// SELECT c.id, c.public_id, c.post_id, c.user_id, c.content, c.created_at, c.updated_at
	// FROM comments c
	// JOIN posts p ON c.post_id = p.id
	// WHERE p.public_id = ? ORDER BY c.created_at ASC
	return nil, nil
}
