package models

// category.go defines the Category model and related database operations.
// It handles category creation and retrieval.

import (
	"time"
)

// Category represents a post category/tag
type Category struct {
	ID          int
	Name        string
	Description string
	CreatedAt   time.Time
}

// GetAllCategories retrieves all available categories
func GetAllCategories() ([]*Category, error) {
	// Query all categories from database
	return nil, nil
}

// GetCategoryByID retrieves a category by its ID
func GetCategoryByID(id int) (*Category, error) {
	// Query category from database by ID
	return nil, nil
}

// GetCategoryByName retrieves a category by its name
func GetCategoryByName(name string) (*Category, error) {
	// Query category from database by name
	return nil, nil
}

// GetCategoriesForPost retrieves all categories associated with a post
func GetCategoriesForPost(postID int) ([]*Category, error) {
	// Query categories for a specific post
	return nil, nil
}
