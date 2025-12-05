// OUTPUT ADAPTER - SQLite Repository
// Package adapters implements the SQLite repository for reactions.
package adapters

import (
	"context"
	"database/sql"
	"forum/internal/modules/reaction/domain"
	"forum/internal/modules/reaction/ports"
)

// SQLiteReactionRepository implements the ReactionRepository interface using SQLite.
type SQLiteReactionRepository struct {
	db *sql.DB
}

// NewSQLiteReactionRepository creates a new SQLite reaction repository.
func NewSQLiteReactionRepository(db *sql.DB) ports.ReactionRepository {
	return &SQLiteReactionRepository{db: db}
}

// Create stores a new reaction in the database.
// TODO: Implement reaction creation with UUID generation.
func (r *SQLiteReactionRepository) Create(ctx context.Context, reaction *domain.Reaction) error {
	// Implementation placeholder
	// 1. Generate UUID for PublicID
	// 2. INSERT INTO reactions (public_id, user_id, target_id, target_type, type, created_at)
	//    VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	return nil
}

// DeleteByTargetPublicID removes a user's reaction from a target by target's public UUID.
// TODO: Implement reaction deletion by target public UUID.
func (r *SQLiteReactionRepository) DeleteByTargetPublicID(ctx context.Context, userID int, targetPublicID string, targetType string) error {
	// Implementation placeholder
	// For posts: DELETE FROM reactions WHERE user_id = ? AND target_id = (SELECT id FROM posts WHERE public_id = ?) AND target_type = ?
	// For comments: DELETE FROM reactions WHERE user_id = ? AND target_id = (SELECT id FROM comments WHERE public_id = ?) AND target_type = ?
	return nil
}

// GetByTargetPublicID retrieves all reactions for a specific target by its public UUID.
// TODO: Implement retrieving reactions by target public UUID.
func (r *SQLiteReactionRepository) GetByTargetPublicID(ctx context.Context, targetPublicID string, targetType string) ([]*domain.Reaction, error) {
	// Implementation placeholder
	// For posts: SELECT id, public_id, user_id, target_id, target_type, type, created_at
	//            FROM reactions WHERE target_id = (SELECT id FROM posts WHERE public_id = ?) AND target_type = ?
	// For comments: Similar join with comments table
	return nil, nil
}

// GetByUserAndTargetPublicID retrieves a user's reaction for a specific target by target's public UUID.
// TODO: Implement retrieving user's reaction for a target by target public UUID.
func (r *SQLiteReactionRepository) GetByUserAndTargetPublicID(ctx context.Context, userID int, targetPublicID string, targetType string) (*domain.Reaction, error) {
	// Implementation placeholder
	// Similar to GetByTargetPublicID but also filter by user_id
	return nil, nil
}

// CountByTargetPublicID returns the number of reactions of a specific type for a target by its public UUID.
// TODO: Implement reaction counting by target public UUID.
func (r *SQLiteReactionRepository) CountByTargetPublicID(ctx context.Context, targetPublicID string, targetType string, reactionType domain.ReactionType) (int, error) {
	// Implementation placeholder
	// SELECT COUNT(*) FROM reactions
	// WHERE target_id = (SELECT id FROM posts/comments WHERE public_id = ?) AND target_type = ? AND type = ?
	return 0, nil
}

// CountByUserID returns the total number of reactions given by a user.
func (r *SQLiteReactionRepository) CountByUserID(ctx context.Context, userID int) (int, error) {
	query := `SELECT COUNT(*) FROM reactions WHERE user_id = ?`
	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
