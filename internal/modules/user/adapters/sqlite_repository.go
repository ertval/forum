// OUTPUT ADAPTER - SQLite Repository
// Package adapters implements the SQLite repository for users.
package adapters

import (
	"context"
	"database/sql"
	"forum/internal/modules/user/domain"
	"forum/internal/modules/user/ports"

	"github.com/gofrs/uuid/v5"
)

// SQLiteUserRepository implements the UserRepository interface using SQLite.
type SQLiteUserRepository struct {
	db *sql.DB
}

// NewSQLiteUserRepository creates a new SQLite user repository.
func NewSQLiteUserRepository(db *sql.DB) ports.UserRepository {
	return &SQLiteUserRepository{
		db: db,
	}
}

// Create stores a new user in the database.
func (r *SQLiteUserRepository) Create(ctx context.Context, user *domain.User) error {
	// Generate public UUID
	publicID, err := uuid.NewV4()
	if err != nil {
		return err
	}
	user.PublicID = publicID.String()

	query := `INSERT INTO users (public_id, email, username, password_hash, role, post_count, comment_count, created_at, updated_at, is_active)
              VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query,
		user.PublicID,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.Role,
		user.PostCount,
		user.CommentCount,
		user.CreatedAt,
		user.UpdatedAt,
		user.IsActive,
	)

	if err != nil {
		return err
	}

	// Get the auto-generated ID and set it in the user object
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.ID = int(id)
	return nil
}

// GetByID retrieves a user by their ID.
func (r *SQLiteUserRepository) GetByID(ctx context.Context, userID int) (*domain.User, error) {
	query := `SELECT id, public_id, email, username, password_hash, role, post_count, comment_count, created_at, updated_at, is_active
              FROM users WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, userID)

	var user domain.User
	var isActive int // SQLite stores booleans as integers (0 or 1)

	err := row.Scan(
		&user.ID,
		&user.PublicID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.PostCount,
		&user.CommentCount,
		&user.CreatedAt,
		&user.UpdatedAt,
		&isActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	user.IsActive = isActive == 1

	return &user, nil
}

// GetByPublicID retrieves a user by their public UUID.
func (r *SQLiteUserRepository) GetByPublicID(ctx context.Context, publicID string) (*domain.User, error) {
	query := `SELECT id, public_id, email, username, password_hash, role, post_count, comment_count, created_at, updated_at, is_active
              FROM users WHERE public_id = ?`

	row := r.db.QueryRowContext(ctx, query, publicID)

	var user domain.User
	var isActive int // SQLite stores booleans as integers (0 or 1)

	err := row.Scan(
		&user.ID,
		&user.PublicID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.PostCount,
		&user.CommentCount,
		&user.CreatedAt,
		&user.UpdatedAt,
		&isActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	user.IsActive = isActive == 1

	return &user, nil
}

// GetByEmail retrieves a user by their email address.
func (r *SQLiteUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, public_id, email, username, password_hash, role, post_count, comment_count, created_at, updated_at, is_active
              FROM users WHERE email = ?`

	row := r.db.QueryRowContext(ctx, query, email)

	var user domain.User
	var isActive int // SQLite stores booleans as integers (0 or 1)

	err := row.Scan(
		&user.ID,
		&user.PublicID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.PostCount,
		&user.CommentCount,
		&user.CreatedAt,
		&user.UpdatedAt,
		&isActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	user.IsActive = isActive == 1

	return &user, nil
}

// GetByUsername retrieves a user by their username.
func (r *SQLiteUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `SELECT id, public_id, email, username, password_hash, role, post_count, comment_count, created_at, updated_at, is_active
              FROM users WHERE username = ?`

	row := r.db.QueryRowContext(ctx, query, username)

	var user domain.User
	var isActive int // SQLite stores booleans as integers (0 or 1)

	err := row.Scan(
		&user.ID,
		&user.PublicID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.PostCount,
		&user.CommentCount,
		&user.CreatedAt,
		&user.UpdatedAt,
		&isActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	user.IsActive = isActive == 1

	return &user, nil
}

// Update updates an existing user in the database.
func (r *SQLiteUserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `UPDATE users 
              SET email=?, username=?, password_hash=?, role=?, post_count=?, comment_count=?, is_active=?, updated_at=?
              WHERE id=?`

	_, err := r.db.ExecContext(ctx, query,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.Role,
		user.PostCount,
		user.CommentCount,
		user.IsActive,
		user.UpdatedAt,
		user.ID,
	)

	return err
}

// Delete removes a user from the database.
func (r *SQLiteUserRepository) Delete(ctx context.Context, userID int) error {
	query := `DELETE FROM users WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// List returns a paginated list of users.
func (r *SQLiteUserRepository) List(ctx context.Context, offset, limit int) ([]*domain.User, error) {
	query := `SELECT id, public_id, email, username, password_hash, role, post_count, comment_count, created_at, updated_at, is_active
              FROM users ORDER BY created_at DESC`

	// Add pagination if limit is specified
	var rows *sql.Rows
	var err error
	if limit > 0 {
		query += ` LIMIT ? OFFSET ?`
		rows, err = r.db.QueryContext(ctx, query, limit, offset)
	} else {
		rows, err = r.db.QueryContext(ctx, query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		var isActive int // SQLite stores booleans as integers (0 or 1)

		err := rows.Scan(
			&user.ID,
			&user.PublicID,
			&user.Email,
			&user.Username,
			&user.PasswordHash,
			&user.Role,
			&user.PostCount,
			&user.CommentCount,
			&user.CreatedAt,
			&user.UpdatedAt,
			&isActive,
		)
		if err != nil {
			return nil, err
		}

		user.IsActive = isActive == 1
		users = append(users, &user)
	}

	return users, nil
}

// Count returns the total number of users.
func (r *SQLiteUserRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM users`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// ExistsByEmail checks if a user with the given email exists.
func (r *SQLiteUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT COUNT(*) FROM users WHERE email = ?`

	var count int
	err := r.db.QueryRowContext(ctx, query, email).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ExistsByUsername checks if a user with the given username exists.
func (r *SQLiteUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	query := `SELECT COUNT(*) FROM users WHERE username = ?`

	var count int
	err := r.db.QueryRowContext(ctx, query, username).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// IncrementPostCount atomically increments the user's post count.
func (r *SQLiteUserRepository) IncrementPostCount(ctx context.Context, userID int) error {
	query := `UPDATE users SET post_count = post_count + 1 WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// DecrementPostCount atomically decrements the user's post count.
func (r *SQLiteUserRepository) DecrementPostCount(ctx context.Context, userID int) error {
	query := `UPDATE users SET post_count = MAX(0, post_count - 1) WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// IncrementCommentCount atomically increments the user's comment count.
func (r *SQLiteUserRepository) IncrementCommentCount(ctx context.Context, userID int) error {
	query := `UPDATE users SET comment_count = comment_count + 1 WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// DecrementCommentCount atomically decrements the user's comment count.
func (r *SQLiteUserRepository) DecrementCommentCount(ctx context.Context, userID int) error {
	query := `UPDATE users SET comment_count = MAX(0, comment_count - 1) WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// IncrementReactionCount atomically increments the user's reaction count.
func (r *SQLiteUserRepository) IncrementReactionCount(ctx context.Context, userID int) error {
	query := `UPDATE users SET reaction_count = reaction_count + 1 WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// DecrementReactionCount atomically decrements the user's reaction count.
func (r *SQLiteUserRepository) DecrementReactionCount(ctx context.Context, userID int) error {
	query := `UPDATE users SET reaction_count = MAX(0, reaction_count - 1) WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}
