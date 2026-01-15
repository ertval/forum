// OUTPUT ADAPTER - SQLite Repository
// Package adapters implements the SQLite repository for reactions.
package adapters

import (
	"context"
	"database/sql"
	"fmt"
	"forum/internal/modules/reaction/domain"
	"forum/internal/modules/reaction/ports"

	"github.com/gofrs/uuid/v5"
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
func (r *SQLiteReactionRepository) Create(ctx context.Context, reaction *domain.Reaction) error {
	// Generate UUID for PublicID
	publicID, err := uuid.NewV4()
	if err != nil {
		return err
	}
	reaction.PublicID = publicID.String()

	// Get internal target ID based on PublicTargetID and targetType
	var targetID int
	switch reaction.TargetType {
	case "post":
		err = r.db.QueryRowContext(ctx, "SELECT id FROM posts WHERE public_id = ?", reaction.PublicTargetID).Scan(&targetID)
	case "comment":
		err = r.db.QueryRowContext(ctx, "SELECT id FROM comments WHERE public_id = ?", reaction.PublicTargetID).Scan(&targetID)
	default:
		return domain.ErrInvalidTarget
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return domain.ErrReactionNotFound
		}
		return err
	}

	// Update the target ID in the reaction object
	reaction.TargetID = targetID

	// Insert the reaction
	query := `
		INSERT INTO reactions (public_id, user_id, target_id, target_type, type, created_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`

	_, err = r.db.ExecContext(ctx, query, reaction.PublicID, reaction.UserID, reaction.TargetID, reaction.TargetType, reaction.Type)
	return err
}

// DeleteByTargetPublicID removes a user's reaction from a target by target's public UUID.
func (r *SQLiteReactionRepository) DeleteByTargetPublicID(ctx context.Context, userID int, targetPublicID string, targetType string) error {
	// Validate target type
	if targetType != "post" && targetType != "comment" {
		return domain.ErrInvalidTarget
	}

	// Get internal target ID based on targetPublicID and targetType
	var targetID int
	var query string
	switch targetType {
	case "post":
		query = "SELECT id FROM posts WHERE public_id = ?"
	case "comment":
		query = "SELECT id FROM comments WHERE public_id = ?"
	}

	err := r.db.QueryRowContext(ctx, query, targetPublicID).Scan(&targetID)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.ErrReactionNotFound
		}
		return err
	}

	// Delete the reaction
	deleteQuery := `
		DELETE FROM reactions
		WHERE user_id = ? AND target_id = ? AND target_type = ?
	`

	result, err := r.db.ExecContext(ctx, deleteQuery, userID, targetID, targetType)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrReactionNotFound
	}

	return nil
}

// GetByTargetPublicID retrieves all reactions for a specific target by its public UUID.
func (r *SQLiteReactionRepository) GetByTargetPublicID(ctx context.Context, targetPublicID string, targetType string) ([]*domain.Reaction, error) {
	// Validate target type
	if targetType != "post" && targetType != "comment" {
		return nil, domain.ErrInvalidTarget
	}

	// Get internal target ID based on targetPublicID and targetType
	var targetID int
	var query string
	switch targetType {
	case "post":
		query = "SELECT id FROM posts WHERE public_id = ?"
	case "comment":
		query = "SELECT id FROM comments WHERE public_id = ?"
	}

	err := r.db.QueryRowContext(ctx, query, targetPublicID).Scan(&targetID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrReactionNotFound
		}
		return nil, err
	}

	// Get all reactions for this target
	selectQuery := `
		SELECT id, public_id, user_id, target_id, target_type, type, created_at
		FROM reactions
		WHERE target_id = ? AND target_type = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, selectQuery, targetID, targetType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reactions []*domain.Reaction
	for rows.Next() {
		var reaction domain.Reaction
		var createdAt sql.NullTime

		err := rows.Scan(
			&reaction.ID,
			&reaction.PublicID,
			&reaction.UserID,
			&reaction.TargetID,
			&reaction.TargetType,
			&reaction.Type,
			&createdAt,
		)
		if err != nil {
			return nil, err
		}

		if createdAt.Valid {
			reaction.CreatedAt = createdAt.Time
		}

		reactions = append(reactions, &reaction)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating reactions: %w", err)
	}

	return reactions, nil
}

// GetByUserAndTargetPublicID retrieves a user's reaction for a specific target by target's public UUID.
func (r *SQLiteReactionRepository) GetByUserAndTargetPublicID(ctx context.Context, userID int, targetPublicID string, targetType string) (*domain.Reaction, error) {
	// Validate target type
	if targetType != "post" && targetType != "comment" {
		return nil, domain.ErrInvalidTarget
	}

	// Get internal target ID based on targetPublicID and targetType
	var targetID int
	var query string
	switch targetType {
	case "post":
		query = "SELECT id FROM posts WHERE public_id = ?"
	case "comment":
		query = "SELECT id FROM comments WHERE public_id = ?"
	}

	err := r.db.QueryRowContext(ctx, query, targetPublicID).Scan(&targetID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrReactionNotFound
		}
		return nil, err
	}

	// Get the user's reaction for this target
	selectQuery := `
		SELECT id, public_id, user_id, target_id, target_type, type, created_at
		FROM reactions
		WHERE user_id = ? AND target_id = ? AND target_type = ?
	`

	var reaction domain.Reaction
	var createdAt sql.NullTime

	err = r.db.QueryRowContext(ctx, selectQuery, userID, targetID, targetType).Scan(
		&reaction.ID,
		&reaction.PublicID,
		&reaction.UserID,
		&reaction.TargetID,
		&reaction.TargetType,
		&reaction.Type,
		&createdAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrReactionNotFound
		}
		return nil, err
	}

	if createdAt.Valid {
		reaction.CreatedAt = createdAt.Time
	}

	return &reaction, nil
}

// CountByTargetPublicID returns the number of reactions of a specific type for a target by its public UUID.
func (r *SQLiteReactionRepository) CountByTargetPublicID(ctx context.Context, targetPublicID string, targetType string, reactionType domain.ReactionType) (int, error) {
	// Validate target type
	if targetType != "post" && targetType != "comment" {
		return 0, domain.ErrInvalidTarget
	}

	// Validate reaction type
	if reactionType != domain.ReactionLike && reactionType != domain.ReactionDislike {
		return 0, domain.ErrInvalidReactionType
	}

	// Get internal target ID based on targetPublicID and targetType
	var targetID int
	var query string
	switch targetType {
	case "post":
		query = "SELECT id FROM posts WHERE public_id = ?"
	case "comment":
		query = "SELECT id FROM comments WHERE public_id = ?"
	}

	err := r.db.QueryRowContext(ctx, query, targetPublicID).Scan(&targetID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, domain.ErrReactionNotFound
		}
		return 0, err
	}

	// Count the reactions
	countQuery := `
		SELECT COUNT(*)
		FROM reactions
		WHERE target_id = ? AND target_type = ? AND type = ?
	`

	var count int
	err = r.db.QueryRowContext(ctx, countQuery, targetID, targetType, reactionType).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
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
