// OUTPUT ADAPTER - SQLite Repository
// Package adapters implements the SQLite repository for reactions.
package adapters

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"forum/internal/modules/reaction/domain"
	"forum/internal/modules/reaction/ports"

	"github.com/gofrs/uuid/v5"
)

// Reaction SQL constants.
const reactionColumns = `id, public_id, user_id, target_id, target_type, type, created_at`

// targetTableForType maps a target type to its database table name.
// Returns empty string for invalid target types.
func targetTableForType(targetType string) string {
	switch targetType {
	case "post":
		return "posts"
	case "comment":
		return "comments"
	default:
		return ""
	}
}

// scanReaction scans a reaction row from any scanner.
func scanReaction(scanner interface{ Scan(dest ...any) error }) (*domain.Reaction, error) {
	var reaction domain.Reaction
	var createdAt sql.NullTime

	err := scanner.Scan(
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

	return &reaction, nil
}

// SQLiteReactionRepository implements the ReactionRepository interface using SQLite.
type SQLiteReactionRepository struct {
	db *sql.DB
}

// NewSQLiteReactionRepository creates a new SQLite reaction repository.
func NewSQLiteReactionRepository(db *sql.DB) ports.ReactionRepository {
	return &SQLiteReactionRepository{db: db}
}

// resolveTargetID resolves a public UUID to an internal integer ID for a post or comment.
// Returns domain.ErrTargetNotFound if the target does not exist.
func (r *SQLiteReactionRepository) resolveTargetID(ctx context.Context, publicID, targetType string) (int, error) {
	table := targetTableForType(targetType)
	if table == "" {
		return 0, domain.ErrInvalidTarget
	}

	var id int
	err := r.db.QueryRowContext(ctx, "SELECT id FROM "+table+" WHERE public_id = ?", publicID).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, domain.ErrTargetNotFound
		}
		return 0, err
	}
	return id, nil
}

// resolveTargetIDTx resolves a public UUID to an internal integer ID within a transaction.
func (r *SQLiteReactionRepository) resolveTargetIDTx(ctx context.Context, tx *sql.Tx, publicID, targetType string) (int, error) {
	table := targetTableForType(targetType)
	if table == "" {
		return 0, domain.ErrInvalidTarget
	}

	var id int
	err := tx.QueryRowContext(ctx, "SELECT id FROM "+table+" WHERE public_id = ?", publicID).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, domain.ErrTargetNotFound
		}
		return 0, err
	}
	return id, nil
}

// Create stores a new reaction in the database.
func (r *SQLiteReactionRepository) Create(ctx context.Context, reaction *domain.Reaction) error {
	// Generate UUID for PublicID
	publicID, err := uuid.NewV4()
	if err != nil {
		return err
	}
	reaction.PublicID = publicID.String()

	// Resolve target public ID to internal ID
	targetID, err := r.resolveTargetID(ctx, reaction.PublicTargetID, reaction.TargetType)
	if err != nil {
		return err
	}
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

	targetTable := targetTableForType(targetType)

	// Delete directly using target public UUID resolution in a subquery.
	deleteQuery := fmt.Sprintf(`
		DELETE FROM reactions
		WHERE user_id = ?
		  AND target_type = ?
		  AND target_id = (SELECT id FROM %s WHERE public_id = ?)
	`, targetTable)

	result, err := r.db.ExecContext(ctx, deleteQuery, userID, targetType, targetPublicID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		// Preserve semantic distinction between target missing and reaction missing.
		existsQuery := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE public_id = ?)", targetTable)
		var targetExists bool
		err = r.db.QueryRowContext(ctx, existsQuery, targetPublicID).Scan(&targetExists)
		if err != nil {
			return err
		}
		if !targetExists {
			return domain.ErrTargetNotFound
		}
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

	// Resolve target public ID to internal ID
	targetID, err := r.resolveTargetID(ctx, targetPublicID, targetType)
	if err != nil {
		return nil, err
	}

	// Get all reactions for this target
	selectQuery := `SELECT ` + reactionColumns + ` FROM reactions
		WHERE target_id = ? AND target_type = ?
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, selectQuery, targetID, targetType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reactions []*domain.Reaction
	for rows.Next() {
		reaction, scanErr := scanReaction(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		reactions = append(reactions, reaction)
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

	// Resolve target public ID to internal ID
	targetID, err := r.resolveTargetID(ctx, targetPublicID, targetType)
	if err != nil {
		return nil, err
	}

	// Get the user's reaction for this target
	selectQuery := `SELECT ` + reactionColumns + ` FROM reactions
		WHERE user_id = ? AND target_id = ? AND target_type = ?`

	reaction, scanErr := scanReaction(r.db.QueryRowContext(ctx, selectQuery, userID, targetID, targetType))
	if scanErr != nil {
		if scanErr == sql.ErrNoRows {
			return nil, domain.ErrReactionNotFound
		}
		return nil, scanErr
	}

	return reaction, nil
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

	// Resolve target public ID to internal ID
	targetID, err := r.resolveTargetID(ctx, targetPublicID, targetType)
	if err != nil {
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

// CountLikesAndDislikesByTargetPublicID returns both likes and dislikes in a single query.
func (r *SQLiteReactionRepository) CountLikesAndDislikesByTargetPublicID(ctx context.Context, targetPublicID string, targetType string) (likes, dislikes int, err error) {
	// Validate target type
	if targetType != "post" && targetType != "comment" {
		return 0, 0, domain.ErrInvalidTarget
	}

	// Resolve target ID once
	targetID, err := r.resolveTargetID(ctx, targetPublicID, targetType)
	if err != nil {
		if err == domain.ErrTargetNotFound {
			return 0, 0, nil // Target doesn't exist, return zero counts
		}
		return 0, 0, err
	}

	// Count both in one query
	query := `SELECT
		COALESCE(SUM(CASE WHEN type = 'like' THEN 1 ELSE 0 END), 0) as likes,
		COALESCE(SUM(CASE WHEN type = 'dislike' THEN 1 ELSE 0 END), 0) as dislikes
	FROM reactions WHERE target_id = ? AND target_type = ?`

	err = r.db.QueryRowContext(ctx, query, targetID, targetType).Scan(&likes, &dislikes)
	if err != nil {
		return 0, 0, err
	}
	return likes, dislikes, nil
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

// ListByUserID returns all reactions made by a user, newest first.
// Includes resolved target public UUID in Reaction.PublicTargetID.
func (r *SQLiteReactionRepository) ListByUserID(ctx context.Context, userID int) ([]*domain.Reaction, error) {
	query := `
		SELECT
			r.id,
			r.public_id,
			r.user_id,
			r.target_id,
			r.target_type,
			r.type,
			r.created_at,
			COALESCE(p.public_id, c.public_id) AS target_public_id
		FROM reactions r
		LEFT JOIN posts p ON r.target_type = 'post' AND p.id = r.target_id
		LEFT JOIN comments c ON r.target_type = 'comment' AND c.id = r.target_id
		WHERE r.user_id = ?
		ORDER BY r.created_at DESC, r.id DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reactions := make([]*domain.Reaction, 0, 16)
	for rows.Next() {
		var reaction domain.Reaction
		var createdAt sql.NullTime
		var publicTargetID sql.NullString

		err := rows.Scan(
			&reaction.ID,
			&reaction.PublicID,
			&reaction.UserID,
			&reaction.TargetID,
			&reaction.TargetType,
			&reaction.Type,
			&createdAt,
			&publicTargetID,
		)
		if err != nil {
			return nil, err
		}

		if createdAt.Valid {
			reaction.CreatedAt = createdAt.Time
		}
		if publicTargetID.Valid {
			reaction.PublicTargetID = publicTargetID.String
		}

		reactions = append(reactions, &reaction)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating user reactions: %w", err)
	}

	return reactions, nil
}

// CountBatchByTargetPublicIDs returns like and dislike counts for multiple targets in a single query.
func (r *SQLiteReactionRepository) CountBatchByTargetPublicIDs(ctx context.Context, targetPublicIDs []string, targetType string) (map[string]map[string]int, error) {
	if targetType != "post" && targetType != "comment" {
		return nil, domain.ErrInvalidTarget
	}
	if len(targetPublicIDs) == 0 {
		return make(map[string]map[string]int), nil
	}

	targetTable := targetTableForType(targetType)

	placeholders := strings.Repeat("?,", len(targetPublicIDs))
	placeholders = placeholders[:len(placeholders)-1]

	query := fmt.Sprintf(`
		SELECT t.public_id, COALESCE(r.type, ''), COUNT(r.id)
		FROM %s t
		LEFT JOIN reactions r ON r.target_id = t.id AND r.target_type = ?
		WHERE t.public_id IN (%s)
		GROUP BY t.public_id, r.type
	`, targetTable, placeholders)

	args := make([]interface{}, 0, 1+len(targetPublicIDs))
	args = append(args, targetType)
	for _, id := range targetPublicIDs {
		args = append(args, id)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]map[string]int)
	for rows.Next() {
		var publicID, reactionType string
		var count int
		if err := rows.Scan(&publicID, &reactionType, &count); err != nil {
			return nil, err
		}
		if _, ok := result[publicID]; !ok {
			result[publicID] = make(map[string]int)
		}
		if reactionType != "" {
			result[publicID][reactionType] = count
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating batch reaction counts: %w", err)
	}

	return result, nil
}

// ToggleReaction atomically handles the full reaction toggle flow in a single transaction.
// It resolves the target, checks for an existing reaction, and either:
// - Deletes the reaction if the same type already exists (toggle off, removed=true)
// - Updates the reaction type if a different type exists (removed=false)
// - Creates a new reaction if none exists (removed=false)
func (r *SQLiteReactionRepository) ToggleReaction(ctx context.Context, reaction *domain.Reaction) (action domain.ToggleAction, err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	// Resolve the target within the transaction
	targetID, err := r.resolveTargetIDTx(ctx, tx, reaction.PublicTargetID, reaction.TargetType)
	if err != nil {
		return "", err
	}
	reaction.TargetID = targetID

	// Check for an existing reaction within the same transaction
	var existingType string
	err = tx.QueryRowContext(ctx,
		"SELECT type FROM reactions WHERE user_id = ? AND target_id = ? AND target_type = ?",
		reaction.UserID, targetID, reaction.TargetType,
	).Scan(&existingType)

	if err == nil {
		// Existing reaction found
		if domain.ReactionType(existingType) == reaction.Type {
			// Same type — toggle off (delete)
			_, err = tx.ExecContext(ctx,
				"DELETE FROM reactions WHERE user_id = ? AND target_id = ? AND target_type = ?",
				reaction.UserID, targetID, reaction.TargetType,
			)
			if err != nil {
				return "", err
			}
			if err := tx.Commit(); err != nil {
				return "", err
			}
			return domain.ToggleActionRemoved, nil
		}
		// Different type — update in place
		_, err = tx.ExecContext(ctx,
			"UPDATE reactions SET type = ? WHERE user_id = ? AND target_id = ? AND target_type = ?",
			reaction.Type, reaction.UserID, targetID, reaction.TargetType,
		)
		if err != nil {
			return "", err
		}
		if err := tx.Commit(); err != nil {
			return "", err
		}
		return domain.ToggleActionUpdated, nil
	}

	if err != sql.ErrNoRows {
		return "", err
	}

	// No existing reaction — generate UUID and insert
	publicID, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	reaction.PublicID = publicID.String()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO reactions (public_id, user_id, target_id, target_type, type, created_at)
		 VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		reaction.PublicID, reaction.UserID, targetID, reaction.TargetType, reaction.Type,
	)
	if err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return domain.ToggleActionCreated, nil
}
