// OUTPUT ADAPTER - SQLite Repository
// Package adapters implements the SQLite repository for comments.
package adapters

import (
	"context"
	"database/sql"
	"fmt"

	"forum/internal/modules/comment/domain"
	"forum/internal/modules/comment/ports"

	"github.com/gofrs/uuid/v5"
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
func (r *SQLiteCommentRepository) Create(ctx context.Context, comment *domain.Comment) error {
	// Generate UUID for PublicID
	u, err := uuid.NewV4()
	if err != nil {
		return err
	}
	publicID := u.String()
	comment.PublicID = publicID

	// Execute the insert query
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO comments (public_id, post_id, author_id, content, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, publicID, comment.PostID, comment.UserID, comment.Content, comment.CreatedAt, comment.UpdatedAt)

	return err
}

// GetByPublicID retrieves a comment by its public UUID.
func (r *SQLiteCommentRepository) GetByPublicID(ctx context.Context, commentPublicID string) (*domain.Comment, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT c.id, c.public_id, c.post_id, c.author_id, c.content, c.created_at, c.updated_at,
		       p.public_id AS post_public_id, u.public_id AS user_public_id
		FROM comments c
		LEFT JOIN posts p ON c.post_id = p.id
		LEFT JOIN users u ON c.author_id = u.id
		WHERE c.public_id = ?
	`, commentPublicID)

	var comment domain.Comment
	var postPublicID, userPublicID sql.NullString
	err := row.Scan(
		&comment.ID,
		&comment.PublicID,
		&comment.PostID,
		&comment.UserID,
		&comment.Content,
		&comment.CreatedAt,
		&comment.UpdatedAt,
		&postPublicID,
		&userPublicID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrCommentNotFound
		}
		return nil, err
	}

	if postPublicID.Valid {
		comment.PublicPostID = postPublicID.String
	}
	if userPublicID.Valid {
		comment.PublicUserID = userPublicID.String
	}

	return &comment, nil
}

// Update updates an existing comment.
func (r *SQLiteCommentRepository) Update(ctx context.Context, comment *domain.Comment) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE comments
		SET content = ?, updated_at = CURRENT_TIMESTAMP
		WHERE public_id = ?
	`, comment.Content, comment.PublicID)
	return err
}

// DeleteByPublicID removes a comment by its public UUID.
func (r *SQLiteCommentRepository) DeleteByPublicID(ctx context.Context, commentPublicID string) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM comments
		WHERE public_id = ?
	`, commentPublicID)
	return err
}

// ListByPostPublicID retrieves all comments for a post by the post's public UUID.
func (r *SQLiteCommentRepository) ListByPostPublicID(ctx context.Context, postPublicID string) ([]*domain.Comment, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.id, c.public_id, c.post_id, c.author_id, c.content, c.created_at, c.updated_at,
		       p.public_id AS post_public_id, u.public_id AS user_public_id
		FROM comments c
		JOIN posts p ON c.post_id = p.id
		LEFT JOIN users u ON c.author_id = u.id
		WHERE p.public_id = ?
		ORDER BY c.created_at ASC
	`, postPublicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := make([]*domain.Comment, 0)
	for rows.Next() {
		var comment domain.Comment
		var postPublicID, userPublicID string
		err := rows.Scan(
			&comment.ID,
			&comment.PublicID,
			&comment.PostID,
			&comment.UserID,
			&comment.Content,
			&comment.CreatedAt,
			&comment.UpdatedAt,
			&postPublicID,
			&userPublicID,
		)
		if err != nil {
			return nil, err
		}
		comment.PublicPostID = postPublicID
		comment.PublicUserID = userPublicID
		comments = append(comments, &comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating comments: %w", err)
	}

	return comments, nil
}

// ListByUser retrieves all comments made by a specific user.
func (r *SQLiteCommentRepository) ListByUser(ctx context.Context, userID int) ([]*domain.Comment, error) {
	query := `
		SELECT c.id, c.public_id, c.post_id, c.author_id, c.content, c.created_at, c.updated_at, p.public_id as post_public_id
		FROM comments c
		INNER JOIN posts p ON c.post_id = p.id
		WHERE c.author_id = ?
		ORDER BY c.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*domain.Comment
	for rows.Next() {
		var comment domain.Comment
		var postPublicID string
		err := rows.Scan(
			&comment.ID,
			&comment.PublicID,
			&comment.PostID,
			&comment.UserID,
			&comment.Content,
			&comment.CreatedAt,
			&comment.UpdatedAt,
			&postPublicID,
		)
		if err != nil {
			return nil, err
		}
		comment.PublicPostID = postPublicID
		comments = append(comments, &comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating user comments: %w", err)
	}

	return comments, nil
}

// ListByUserPaginated retrieves comments made by a user with limit and offset.
func (r *SQLiteCommentRepository) ListByUserPaginated(ctx context.Context, userID int, limit, offset int) ([]*domain.Comment, error) {
	query := `
		SELECT c.id, c.public_id, c.post_id, c.author_id, c.content, c.created_at, c.updated_at, p.public_id as post_public_id
		FROM comments c
		INNER JOIN posts p ON c.post_id = p.id
		WHERE c.author_id = ?
		ORDER BY c.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*domain.Comment
	for rows.Next() {
		var comment domain.Comment
		var postPublicID string
		err := rows.Scan(
			&comment.ID,
			&comment.PublicID,
			&comment.PostID,
			&comment.UserID,
			&comment.Content,
			&comment.CreatedAt,
			&comment.UpdatedAt,
			&postPublicID,
		)
		if err != nil {
			return nil, err
		}
		comment.PublicPostID = postPublicID
		comments = append(comments, &comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating paginated comments: %w", err)
	}

	return comments, nil
}
