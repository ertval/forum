// Package domain contains the core business entities for the post module.
package domain

// Category represents a post category.
type Category struct {
    ID          int
    Name        string
    Description string
}
