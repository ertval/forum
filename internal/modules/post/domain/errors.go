// Package domain contains error definitions for the post module.
package domain

import "errors"

// Domain errors for the post module.
var (
	// ErrPostNotFound is returned when a post doesn't exist.
	ErrPostNotFound = errors.New("post not found")

	// ErrCategoryNotFound is returned when a category doesn't exist.
	ErrCategoryNotFound = errors.New("category not found")

	// ErrInvalidImage is returned when an image file is invalid.
	ErrInvalidImage = errors.New("invalid image file")

	// ErrImageTooLarge is returned when an image exceeds size limit.
	ErrImageTooLarge = errors.New("image file too large")

	// ErrInvalidImageType is returned when image type is not allowed.
	ErrInvalidImageType = errors.New("invalid image type, must be JPEG, PNG, or GIF")

	// ErrEmptyTitle is returned when post title is empty.
	ErrEmptyTitle = errors.New("post title cannot be empty")

	// ErrEmptyContent is returned when post content is empty.
	ErrEmptyContent = errors.New("post content cannot be empty")

	// ErrNoCategories is returned when no categories are specified.
	ErrNoCategories = errors.New("post must have at least one category")
)
