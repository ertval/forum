// Package domain contains the core business entities for the post module.
package domain

import "time"

// Category represents a post category.
type Category struct {
	ID          int       `json:"-"`                     // Internal unique identifier (INT PRIMARY KEY)
	PublicID    string    `json:"id"`                    // Public UUID identifier (exposed in API)
	Name        string    `json:"name"`                  // Category name (unique, used for filtering)
	Description string    `json:"description,omitempty"` // Category description
	CreatedAt   time.Time `json:"created_at"`            // Category creation timestamp
}

// Validate checks if the category is valid.
func (c *Category) Validate() error {
	if c.Name == "" {
		return ErrEmptyCategoryName
	}
	if len(c.Name) > 50 {
		return ErrCategoryNameTooLong
	}
	if len(c.Description) > 500 {
		return ErrCategoryDescriptionTooLong
	}
	return nil
}
