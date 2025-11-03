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
// TODO: Implement comment creation.
func (r *SQLiteCommentRepository) Create(ctx context.Context, comment *domain.Comment) error {
	// Implementation placeholder
	// INSERT INTO comments (post_id, user_id, content, created_at)
	// VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	return nil
}

// GetByID retrieves a comment by ID.
// TODO: Implement comment retrieval by ID.
func (r *SQLiteCommentRepository) GetByID(ctx context.Context, commentID int) (*domain.Comment, error) {
	// Implementation placeholder
	// SELECT id, post_id, user_id, content, created_at, updated_at
	// FROM comments WHERE id = ?
	return nil, nil
}

// Update updates an existing comment.
// TODO: Implement comment update.
func (r *SQLiteCommentRepository) Update(ctx context.Context, comment *domain.Comment) error {
	// Implementation placeholder
	// UPDATE comments SET content=?, updated_at=CURRENT_TIMESTAMP WHERE id=?
	return nil
}

// Delete removes a comment.
// TODO: Implement comment deletion.
func (r *SQLiteCommentRepository) Delete(ctx context.Context, commentID int) error {
	// Implementation placeholder
	// DELETE FROM comments WHERE id = ?
	return nil
}

// ListByPostID retrieves all comments for a post.
// TODO: Implement comment listing by post ID.
func (r *SQLiteCommentRepository) ListByPostID(ctx context.Context, postID int) ([]*domain.Comment, error) {
	// Implementation placeholder
	// SELECT id, post_id, user_id, content, created_at, updated_at
	// FROM comments WHERE post_id = ? ORDER BY created_at ASC
	return nil, nil
}
