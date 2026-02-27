// OUTPUT ADAPTER - SQLite Repository
// Package adapters implements the SQLite repository for posts.
package adapters

import (
	"context"
	"database/sql"
	"fmt"
	"forum/internal/modules/post/domain"
	"forum/internal/modules/post/ports"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
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

	// Generate public UUID
	publicID, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("failed to generate UUID: %w", err)
	}
	post.PublicID = publicID.String()

	// Insert post
	query := `
		INSERT INTO posts (public_id, title, content, author_id, image_path, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	var imagePath *string
	if post.ImageURL != "" {
		imagePath = &post.ImageURL
	}

	result, err := tx.ExecContext(ctx, query,
		post.PublicID,
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

	// Get the auto-generated internal ID
	postID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get post ID: %w", err)
	}
	post.ID = int(postID)

	// Insert post-category associations
	for _, categoryName := range post.Categories {
		// Get category internal ID by name (case-insensitive)
		var categoryID int
		err := tx.QueryRowContext(ctx, "SELECT id FROM categories WHERE LOWER(name) = LOWER(?)", categoryName).Scan(&categoryID)
		if err != nil {
			return fmt.Errorf("category %s not found: %w", categoryName, err)
		}

		// Insert association using internal IDs
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
			p.id, p.public_id, p.title, p.content, p.author_id, p.image_path,
			p.created_at, p.updated_at,
			u.public_id as user_public_id, u.username,
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
		WHERE p.public_id = ?
	`

	var post domain.Post
	var imageURL sql.NullString
	var username sql.NullString
	var userPublicID sql.NullString

	err := r.db.QueryRowContext(ctx, query, postID).Scan(
		&post.ID,
		&post.PublicID,
		&post.Title,
		&post.Content,
		&post.UserID,
		&imageURL,
		&post.CreatedAt,
		&post.UpdatedAt,
		&userPublicID,
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
		normalized := normalizeImagePath(imageURL.String)
		if normalized != "" {
			post.ImageURL = "/static/uploads/" + normalized
		}
	}
	if username.Valid {
		post.AuthorUsername = username.String
		post.Author = username.String // Set both fields for compatibility
	}
	if userPublicID.Valid {
		post.UserPublicID = userPublicID.String
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
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		UPDATE posts 
		SET title = ?, content = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := tx.ExecContext(ctx, query, post.Title, post.Content, post.UpdatedAt, post.ID)
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

	// Update categories: delete old associations and insert new ones
	_, err = tx.ExecContext(ctx, "DELETE FROM post_categories WHERE post_id = ?", post.ID)
	if err != nil {
		return fmt.Errorf("failed to delete old post-category associations: %w", err)
	}

	// Insert new post-category associations
	for _, categoryName := range post.Categories {
		var categoryID int
		err := tx.QueryRowContext(ctx, "SELECT id FROM categories WHERE LOWER(name) = LOWER(?)", categoryName).Scan(&categoryID)
		if err != nil {
			return fmt.Errorf("category %s not found: %w", categoryName, err)
		}

		_, err = tx.ExecContext(ctx, "INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)",
			post.ID, categoryID)
		if err != nil {
			return fmt.Errorf("failed to insert post-category association: %w", err)
		}
	}

	return tx.Commit()
}

// Delete removes a post.
func (r *SQLitePostRepository) Delete(ctx context.Context, postID string) error {
	query := "DELETE FROM posts WHERE public_id = ?"

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

// UpdateImagePath updates only the image_path field for a post.
func (r *SQLitePostRepository) UpdateImagePath(ctx context.Context, postID string, imagePath string) error {
	var imgPath *string
	if imagePath != "" {
		imgPath = &imagePath
	}

	query := "UPDATE posts SET image_path = ?, updated_at = ? WHERE public_id = ?"
	result, err := r.db.ExecContext(ctx, query, imgPath, time.Now(), postID)
	if err != nil {
		return fmt.Errorf("failed to update image path: %w", err)
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

// GetImagePath retrieves the image_path for a post by its public ID.
func (r *SQLitePostRepository) GetImagePath(ctx context.Context, postID string) (string, error) {
	var imagePath sql.NullString
	query := "SELECT image_path FROM posts WHERE public_id = ?"
	err := r.db.QueryRowContext(ctx, query, postID).Scan(&imagePath)
	if err == sql.ErrNoRows {
		return "", domain.ErrPostNotFound
	}
	if err != nil {
		return "", fmt.Errorf("failed to get image path: %w", err)
	}
	if imagePath.Valid {
		return imagePath.String, nil
	}
	return "", nil
}

// List returns filtered posts.
func (r *SQLitePostRepository) List(ctx context.Context, filter domain.PostFilter) ([]*domain.Post, error) {
	query := `
		SELECT DISTINCT 
			p.id, p.public_id, p.title, p.content, p.author_id, p.image_path, 
			p.created_at, p.updated_at,
			u.public_id as user_public_id, u.username,
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

	// Filter by user (created posts) - now using user public_id string
	if filter.UserID != "" {
		conditions = append(conditions, "u.public_id = ?")
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
		INNER JOIN users liked_user ON r.user_id = liked_user.id
		`
		conditions = append(conditions, "liked_user.public_id = ? AND r.target_type = 'post' AND r.type = 'like'")
		args = append(args, filter.LikedByUserID)
	}

	// Filter by posts where user has commented
	if filter.CommenterID != "" {
		query += `
		INNER JOIN comments cmt ON p.id = cmt.post_id
		INNER JOIN users cmt_user ON cmt.author_id = cmt_user.id
		`
		conditions = append(conditions, "cmt_user.public_id = ?")
		args = append(args, filter.CommenterID)
	}

	// Filter by date
	if filter.DateFilter != "" && filter.DateFilter != "all" {
		switch filter.DateFilter {
		case "today":
			conditions = append(conditions, "DATE(p.created_at) = DATE('now')")
		case "week":
			conditions = append(conditions, "p.created_at >= DATE('now', '-7 days')")
		case "month":
			conditions = append(conditions, "p.created_at >= DATE('now', '-30 days')")
		}
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
		var userPublicID sql.NullString

		err := rows.Scan(
			&post.ID,
			&post.PublicID,
			&post.Title,
			&post.Content,
			&post.UserID,
			&imageURL,
			&post.CreatedAt,
			&post.UpdatedAt,
			&userPublicID,
			&username,
			&post.LikeCount,
			&post.DislikeCount,
			&post.CommentCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}

		if imageURL.Valid {
			normalized := normalizeImagePath(imageURL.String)
			if normalized != "" {
				post.ImageURL = "/static/uploads/" + normalized
			}
		}
		if username.Valid {
			post.AuthorUsername = username.String
			post.Author = username.String // Set both fields for compatibility
		}
		if userPublicID.Valid {
			post.UserPublicID = userPublicID.String
		}

		// Load categories for this post
		categories, err := r.getPostCategories(ctx, post.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get categories for post %d: %w", post.ID, err)
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
func (r *SQLitePostRepository) getPostCategories(ctx context.Context, postID int) ([]string, error) {
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

func normalizeImagePath(path string) string {
	normalized := strings.TrimSpace(path)
	if normalized == "" {
		return ""
	}

	normalized = strings.TrimPrefix(normalized, "./")
	normalized = strings.TrimPrefix(normalized, "/")
	normalized = strings.TrimPrefix(normalized, "static/uploads/")
	normalized = strings.TrimPrefix(normalized, "uploads/")

	if strings.Contains(strings.ToLower(normalized), "seed-placeholder") {
		return ""
	}

	return normalized
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
	// Generate public UUID
	publicID, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("failed to generate UUID: %w", err)
	}
	category.PublicID = publicID.String()

	query := `
		INSERT INTO categories (public_id, name, description, created_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`

	result, err := r.db.ExecContext(ctx, query, category.PublicID, category.Name, category.Description)
	if err != nil {
		return fmt.Errorf("failed to insert category: %w", err)
	}

	// Get the auto-generated internal ID
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get category ID: %w", err)
	}
	category.ID = int(id)

	return nil
}

// GetByID retrieves a category by ID.
func (r *SQLiteCategoryRepository) GetByID(ctx context.Context, categoryID string) (*domain.Category, error) {
	query := "SELECT id, public_id, name, description FROM categories WHERE public_id = ?"

	var category domain.Category
	var description sql.NullString

	err := r.db.QueryRowContext(ctx, query, categoryID).Scan(
		&category.ID,
		&category.PublicID,
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
	query := "SELECT id, public_id, name, description FROM categories WHERE LOWER(name) = LOWER(?)"

	var category domain.Category
	var description sql.NullString

	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&category.ID,
		&category.PublicID,
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
	query := "SELECT id, public_id, name, description FROM categories ORDER BY name"

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categories []*domain.Category
	for rows.Next() {
		var category domain.Category
		var description sql.NullString

		err := rows.Scan(&category.ID, &category.PublicID, &category.Name, &description)
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
	query := "DELETE FROM categories WHERE public_id = ?"

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
