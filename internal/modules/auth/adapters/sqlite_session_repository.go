// OUTPUT ADAPTER - SQLite Repository
// Package adapters implements the SQLite repository for sessions.
// This adapter provides database persistence for session entities.
package adapters

import (
	"context"
	"database/sql"
	"fmt"

	"forum/internal/modules/auth/domain"
	"forum/internal/modules/auth/ports"

	"github.com/gofrs/uuid/v5"
)

// SQLiteSessionRepository implements the SessionRepository interface using SQLite.
type SQLiteSessionRepository struct {
	db *sql.DB
}

// NewSQLiteSessionRepository creates a new SQLite session repository.
func NewSQLiteSessionRepository(db *sql.DB) ports.SessionRepository {
	return &SQLiteSessionRepository{
		db: db,
	}
}

// Create stores a new session in the database.
func (r *SQLiteSessionRepository) Create(ctx context.Context, session *domain.Session) error {
	// Generate public UUID if not already set
	if session.PublicID == "" {
		publicID, err := uuid.NewV4()
		if err != nil {
			return err
		}
		session.PublicID = publicID.String()
	}

	query := `INSERT INTO sessions (public_id, user_id, token, expires_at, created_at, ip_address, user_agent) 
              VALUES (?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query,
		session.PublicID,
		session.UserID,
		session.Token,
		session.ExpiresAt,
		session.CreatedAt,
		session.IPAddress,
		session.UserAgent)

	if err != nil {
		return err
	}

	// Get the auto-generated internal ID
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	session.ID = int(id)
	return nil
}

// GetByToken retrieves a session by its token.
func (r *SQLiteSessionRepository) GetByToken(ctx context.Context, token string) (*domain.Session, error) {
	query := `SELECT id, public_id, user_id, token, expires_at, created_at, ip_address, user_agent
              FROM sessions WHERE token = ?`

	row := r.db.QueryRowContext(ctx, query, token)

	var session domain.Session
	err := row.Scan(
		&session.ID,
		&session.PublicID,
		&session.UserID,
		&session.Token,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.IPAddress,
		&session.UserAgent,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrSessionNotFound
		}
		return nil, err
	}

	return &session, nil
}

// GetByUserID retrieves all active sessions for a user.
func (r *SQLiteSessionRepository) GetByUserID(ctx context.Context, userID int) ([]*domain.Session, error) {
	query := `SELECT id, public_id, user_id, token, expires_at, created_at, ip_address, user_agent
              FROM sessions WHERE user_id = ? AND expires_at > CURRENT_TIMESTAMP`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*domain.Session
	for rows.Next() {
		var session domain.Session
		err := rows.Scan(
			&session.ID,
			&session.PublicID,
			&session.UserID,
			&session.Token,
			&session.ExpiresAt,
			&session.CreatedAt,
			&session.IPAddress,
			&session.UserAgent,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, &session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating sessions: %w", err)
	}

	return sessions, nil
}

// Update updates an existing session in the database.
func (r *SQLiteSessionRepository) Update(ctx context.Context, session *domain.Session) error {
	query := `UPDATE sessions SET expires_at = ? WHERE token = ?`

	_, err := r.db.ExecContext(ctx, query, session.ExpiresAt, session.Token)
	return err
}

// Delete removes a session from the database by its token.
func (r *SQLiteSessionRepository) Delete(ctx context.Context, token string) error {
	query := `DELETE FROM sessions WHERE token = ?`

	result, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrSessionNotFound
	}

	return nil
}

// DeleteByUserID removes all sessions for a user.
func (r *SQLiteSessionRepository) DeleteByUserID(ctx context.Context, userID int) error {
	query := `DELETE FROM sessions WHERE user_id = ?`

	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// DeleteExpired removes all expired sessions from the database.
func (r *SQLiteSessionRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at < CURRENT_TIMESTAMP`

	_, err := r.db.ExecContext(ctx, query)
	return err
}
