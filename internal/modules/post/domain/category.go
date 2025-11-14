// Package domain contains the core business entities for the post module.
package domain

// Category represents a post category.
type Category struct {
	ID          string // Unique category identifier (UUID)
	Name        string // Category name (unique, used for filtering)
	Description string // Category description
}

// Validate checks if the category is valid.
// TODO: Implement category validation.
func (c *Category) Validate() error {
	// Check name is not empty
	// Check name length
	return nil
}
