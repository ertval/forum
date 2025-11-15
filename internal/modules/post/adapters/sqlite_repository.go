// OUTPUT ADAPTER - SQLite Repository
// Package adapters implements the SQLite repository for posts.
package adapters

import (
	"context"
	"database/sql"
	"fmt"
	"forum/internal/modules/post/domain"
	"forum/internal/modules/post/ports"
)

// SQLitePostRepository implements the PostRepository interface using SQLite.
type SQLitePostRepository struct {
	db *sql.DB
}

// NewSQLitePostRepository creates a new SQLite post repository.
func NewSQLitePostRepository(db *sql.DB) ports.PostRepository {
	return &SQLitePostRepository{db: db}
}

// Create stores a new post in the database.
func (r *SQLitePostRepository) Create(ctx context.Context, post *domain.Post) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert post
	query := `
		INSERT INTO posts (id, title, content, author_id, image_path, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	var imagePath *string
	if post.ImageURL != "" {
		imagePath = &post.ImageURL
	}

	_, err = tx.ExecContext(ctx, query,
		post.ID,
		post.Title,
		post.Content,
		post.UserID,
		imagePath,
		post.CreatedAt,
		post.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert post: %w", err)
	}

	// Insert post-category associations
	for _, categoryName := range post.Categories {
		// Get category ID by name (case-insensitive)
		var categoryID string
		err := tx.QueryRowContext(ctx, "SELECT id FROM categories WHERE LOWER(name) = LOWER(?)", categoryName).Scan(&categoryID)
		if err != nil {
			return fmt.Errorf("category %s not found: %w", categoryName, err)
		}

		// Insert association
		_, err = tx.ExecContext(ctx, "INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)",
			post.ID, categoryID)
		if err != nil {
			return fmt.Errorf("failed to insert post-category association: %w", err)
		}
	}

	return tx.Commit()
}

// GetByID retrieves a post by ID with its categories.
func (r *SQLitePostRepository) GetByID(ctx context.Context, postID string) (*domain.Post, error) {
	query := `
		SELECT 
			p.id, p.title, p.content, p.author_id, p.image_path,
			p.created_at, p.updated_at,
			u.username,
			COALESCE(like_counts.count, 0) as like_count,
			COALESCE(dislike_counts.count, 0) as dislike_count,
			COALESCE(comment_counts.count, 0) as comment_count
		FROM posts p
		LEFT JOIN users u ON p.author_id = u.id
		LEFT JOIN (
			SELECT target_id, COUNT(*) as count 
			FROM reactions 
			WHERE target_type = 'post' AND type = 'like'
			GROUP BY target_id
		) like_counts ON p.id = like_counts.target_id
		LEFT JOIN (
			SELECT target_id, COUNT(*) as count 
			FROM reactions 
			WHERE target_type = 'post' AND type = 'dislike'
			GROUP BY target_id
		) dislike_counts ON p.id = dislike_counts.target_id
		LEFT JOIN (
			SELECT post_id, COUNT(*) as count 
			FROM comments 
			GROUP BY post_id
		) comment_counts ON p.id = comment_counts.post_id
		WHERE p.id = ?
	`

	var post domain.Post
	var imageURL sql.NullString
	var username sql.NullString

	err := r.db.QueryRowContext(ctx, query, postID).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.UserID,
		&imageURL,
		&post.CreatedAt,
		&post.UpdatedAt,
		&username,
		&post.LikeCount,
		&post.DislikeCount,
		&post.CommentCount,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrPostNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query post: %w", err)
	}

	if imageURL.Valid {
		post.ImageURL = "/static/uploads/" + imageURL.String
	}
	if username.Valid {
		post.AuthorUsername = username.String
		post.Author = username.String // Set both fields for compatibility
	}

	// Load categories
	categories, err := r.getPostCategories(ctx, post.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	post.Categories = categories

	return &post, nil
}

// Update updates an existing post.
func (r *SQLitePostRepository) Update(ctx context.Context, post *domain.Post) error {
	query := `
		UPDATE posts 
		SET title = ?, content = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query, post.Title, post.Content, post.UpdatedAt, post.ID)
	if err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrPostNotFound
	}

	return nil
}

// Delete removes a post.
func (r *SQLitePostRepository) Delete(ctx context.Context, postID string) error {
	query := "DELETE FROM posts WHERE id = ?"

	result, err := r.db.ExecContext(ctx, query, postID)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrPostNotFound
	}

	return nil
}

// List returns filtered posts.
func (r *SQLitePostRepository) List(ctx context.Context, filter ports.PostFilter) ([]*domain.Post, error) {
	query := `
		SELECT DISTINCT 
			p.id, p.title, p.content, p.author_id, p.image_path, 
			p.created_at, p.updated_at,
			u.username,
			COALESCE(like_counts.count, 0) as like_count,
			COALESCE(dislike_counts.count, 0) as dislike_count,
			COALESCE(comment_counts.count, 0) as comment_count
		FROM posts p
		LEFT JOIN users u ON p.author_id = u.id
		LEFT JOIN (
			SELECT target_id, COUNT(*) as count 
			FROM reactions 
			WHERE target_type = 'post' AND type = 'like'
			GROUP BY target_id
		) like_counts ON p.id = like_counts.target_id
		LEFT JOIN (
			SELECT target_id, COUNT(*) as count 
			FROM reactions 
			WHERE target_type = 'post' AND type = 'dislike'
			GROUP BY target_id
		) dislike_counts ON p.id = dislike_counts.target_id
		LEFT JOIN (
			SELECT post_id, COUNT(*) as count 
			FROM comments 
			GROUP BY post_id
		) comment_counts ON p.id = comment_counts.post_id
	`

	var conditions []string
	var args []interface{}

	// Filter by user (created posts)
	if filter.UserID != "" {
		conditions = append(conditions, "p.author_id = ?")
		args = append(args, filter.UserID)
	}

	// Filter by categories
	if len(filter.Categories) > 0 {
		query += `
		LEFT JOIN post_categories pc ON p.id = pc.post_id
		LEFT JOIN categories c ON pc.category_id = c.id
		`
		conditions = append(conditions, "c.name IN (?"+repeatPlaceholders(len(filter.Categories)-1)+")")
		for _, cat := range filter.Categories {
			args = append(args, cat)
		}
	}

	// Filter by liked posts
	if filter.LikedByUserID != "" {
		query += `
		INNER JOIN reactions r ON p.id = r.target_id 
		`
		conditions = append(conditions, "r.user_id = ? AND r.target_type = 'post' AND r.type = 'like'")
		args = append(args, filter.LikedByUserID)
	}

	// Add WHERE conditions
	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
	}

	// Order by creation date (newest first)
	query += " ORDER BY p.created_at DESC"

	// Apply pagination
	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query posts: %w", err)
	}
	defer rows.Close()

	var posts []*domain.Post
	for rows.Next() {
		var post domain.Post
		var imageURL sql.NullString
		var username sql.NullString

		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.UserID,
			&imageURL,
			&post.CreatedAt,
			&post.UpdatedAt,
			&username,
			&post.LikeCount,
			&post.DislikeCount,
			&post.CommentCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}

		if imageURL.Valid {
			post.ImageURL = "/static/uploads/" + imageURL.String
		}
		if username.Valid {
			post.AuthorUsername = username.String
			post.Author = username.String // Set both fields for compatibility
		}

		// Load categories for this post
		categories, err := r.getPostCategories(ctx, post.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get categories for post %s: %w", post.ID, err)
		}
		post.Categories = categories

		posts = append(posts, &post)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating posts: %w", err)
	}

	return posts, nil
}

// getPostCategories retrieves category names for a specific post.
func (r *SQLitePostRepository) getPostCategories(ctx context.Context, postID string) ([]string, error) {
	query := `
		SELECT c.name 
		FROM categories c
		INNER JOIN post_categories pc ON c.id = pc.category_id
		WHERE pc.post_id = ?
		ORDER BY c.name
	`

	rows, err := r.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		categories = append(categories, name)
	}

	return categories, rows.Err()
}

// repeatPlaceholders returns a string of comma-separated question marks.
func repeatPlaceholders(count int) string {
	if count <= 0 {
		return ""
	}
	result := ""
	for i := 0; i < count; i++ {
		result += ", ?"
	}
	return result
}

// SQLiteCategoryRepository implements the CategoryRepository interface.
type SQLiteCategoryRepository struct {
	db *sql.DB
}

// NewSQLiteCategoryRepository creates a new SQLite category repository.
func NewSQLiteCategoryRepository(db *sql.DB) ports.CategoryRepository {
	return &SQLiteCategoryRepository{db: db}
}

// Create stores a new category.
func (r *SQLiteCategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	query := `
		INSERT INTO categories (id, name, description, created_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`

	_, err := r.db.ExecContext(ctx, query, category.ID, category.Name, category.Description)
	if err != nil {
		return fmt.Errorf("failed to insert category: %w", err)
	}

	return nil
}

// GetByID retrieves a category by ID.
func (r *SQLiteCategoryRepository) GetByID(ctx context.Context, categoryID string) (*domain.Category, error) {
	query := "SELECT id, name, description FROM categories WHERE id = ?"

	var category domain.Category
	var description sql.NullString

	err := r.db.QueryRowContext(ctx, query, categoryID).Scan(
		&category.ID,
		&category.Name,
		&description,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrCategoryNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query category: %w", err)
	}

	if description.Valid {
		category.Description = description.String
	}

	return &category, nil
}

// GetByName retrieves a category by name (case-insensitive).
func (r *SQLiteCategoryRepository) GetByName(ctx context.Context, name string) (*domain.Category, error) {
	query := "SELECT id, name, description FROM categories WHERE LOWER(name) = LOWER(?)"

	var category domain.Category
	var description sql.NullString

	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&category.ID,
		&category.Name,
		&description,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrCategoryNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query category: %w", err)
	}

	if description.Valid {
		category.Description = description.String
	}

	return &category, nil
}

// List retrieves all categories.
func (r *SQLiteCategoryRepository) List(ctx context.Context) ([]*domain.Category, error) {
	query := "SELECT id, name, description FROM categories ORDER BY name"

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categories []*domain.Category
	for rows.Next() {
		var category domain.Category
		var description sql.NullString

		err := rows.Scan(&category.ID, &category.Name, &description)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}

		if description.Valid {
			category.Description = description.String
		}

		categories = append(categories, &category)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating categories: %w", err)
	}

	return categories, nil
}

// Delete removes a category.
func (r *SQLiteCategoryRepository) Delete(ctx context.Context, categoryID string) error {
	query := "DELETE FROM categories WHERE id = ?"

	result, err := r.db.ExecContext(ctx, query, categoryID)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrCategoryNotFound
	}

	return nil
}
