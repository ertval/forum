// OUTPUT ADAPTER - SQLite Repository
// Package adapters implements the SQLite repository for users.
package adapters

import (
	"context"
	"database/sql"
	"forum/internal/modules/user/domain"
	"forum/internal/modules/user/ports"
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
	query := `INSERT INTO users (email, username, password_hash, role, created_at, updated_at, is_active)
              VALUES (?, ?, ?, ?, ?, ?, ?)`
	
	result, err := r.db.ExecContext(ctx, query,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.Role,
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
	query := `SELECT id, email, username, password_hash, role, created_at, updated_at, is_active
              FROM users WHERE id = ?`
	
	row := r.db.QueryRowContext(ctx, query, userID)
	
	var user domain.User
	var isActive int // SQLite stores booleans as integers (0 or 1)
	
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
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
	query := `SELECT id, email, username, password_hash, role, created_at, updated_at, is_active
              FROM users WHERE email = ?`
	
	row := r.db.QueryRowContext(ctx, query, email)
	
	var user domain.User
	var isActive int // SQLite stores booleans as integers (0 or 1)
	
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
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
	query := `SELECT id, email, username, password_hash, role, created_at, updated_at, is_active
              FROM users WHERE username = ?`
	
	row := r.db.QueryRowContext(ctx, query, username)
	
	var user domain.User
	var isActive int // SQLite stores booleans as integers (0 or 1)
	
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
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
              SET email=?, username=?, password_hash=?, role=?, is_active=?, updated_at=?
              WHERE id=?`
	
	_, err := r.db.ExecContext(ctx, query,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.Role,
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
	query := `SELECT id, email, username, password_hash, role, created_at, updated_at, is_active
              FROM users ORDER BY created_at DESC LIMIT ? OFFSET ?`
	
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
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
			&user.Email,
			&user.Username,
			&user.PasswordHash,
			&user.Role,
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
