// OUTPUT ADAPTER - SQLite Repository
// Package adapters implements the SQLite repository for posts.
package adapters

import (
	"context"
	"database/sql"
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
// TODO: Implement post creation with categories.
func (r *SQLitePostRepository) Create(ctx context.Context, post *domain.Post) error {
	// Implementation placeholder
	// 1. INSERT INTO posts (user_id, title, content, image_url, created_at)
	// 2. Get inserted post ID
	// 3. INSERT INTO post_categories for each category
	return nil
}

// GetByID retrieves a post by ID with its categories.
// TODO: Implement post retrieval with categories.
func (r *SQLitePostRepository) GetByID(ctx context.Context, postID int) (*domain.Post, error) {
	// Implementation placeholder
	// 1. SELECT post from posts WHERE id = ?
	// 2. SELECT categories from post_categories JOIN categories
	// 3. Combine into Post entity
	return nil, nil
}

// Update updates an existing post.
// TODO: Implement post update.
func (r *SQLitePostRepository) Update(ctx context.Context, post *domain.Post) error {
	// Implementation placeholder
	// UPDATE posts SET title=?, content=?, updated_at=CURRENT_TIMESTAMP WHERE id=?
	return nil
}

// Delete removes a post.
// TODO: Implement post deletion with cascading to related tables.
func (r *SQLitePostRepository) Delete(ctx context.Context, postID int) error {
	// Implementation placeholder
	// DELETE FROM posts WHERE id = ?
	// (post_categories, comments, reactions should cascade)
	return nil
}

// List returns filtered posts.
// TODO: Implement post listing with filters.
func (r *SQLitePostRepository) List(ctx context.Context, filter ports.PostFilter) ([]*domain.Post, error) {
	// Implementation placeholder
	// Build query with WHERE clauses based on filter
	// - Filter by UserID (created by user)
	// - Filter by Categories (posts in specific categories)
	// - Filter by LikedByUserID (posts liked by user via reactions join)
	// Apply LIMIT and OFFSET for pagination
	// Load categories for each post
	return nil, nil
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
// TODO: Implement category creation.
func (r *SQLiteCategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	// INSERT INTO categories (name, description) VALUES (?, ?)
	return nil
}

// GetByID retrieves a category by ID.
// TODO: Implement category retrieval by ID.
func (r *SQLiteCategoryRepository) GetByID(ctx context.Context, categoryID int) (*domain.Category, error) {
	// SELECT id, name, description FROM categories WHERE id = ?
	return nil, nil
}

// GetByName retrieves a category by name.
// TODO: Implement category retrieval by name.
func (r *SQLiteCategoryRepository) GetByName(ctx context.Context, name string) (*domain.Category, error) {
	// SELECT id, name, description FROM categories WHERE name = ?
	return nil, nil
}

// List retrieves all categories.
// TODO: Implement category listing.
func (r *SQLiteCategoryRepository) List(ctx context.Context) ([]*domain.Category, error) {
	// SELECT id, name, description FROM categories ORDER BY name
	return nil, nil
}

// Delete removes a category.
// TODO: Implement category deletion.
func (r *SQLiteCategoryRepository) Delete(ctx context.Context, categoryID int) error {
	// DELETE FROM categories WHERE id = ?
	return nil
}
