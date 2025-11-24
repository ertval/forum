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

	// ErrTitleTooLong is returned when post title exceeds max length.
	ErrTitleTooLong = errors.New("post title too long (max 300 characters)")

	// ErrContentTooLong is returned when post content exceeds max length.
	ErrContentTooLong = errors.New("post content too long (max 50000 characters)")

	// ErrEmptyCategoryName is returned when category name is empty.
	ErrEmptyCategoryName = errors.New("category name cannot be empty")

	// ErrCategoryNameTooLong is returned when category name exceeds max length.
	ErrCategoryNameTooLong = errors.New("category name too long (max 50 characters)")

	// ErrCategoryDescriptionTooLong is returned when category description exceeds max length.
	ErrCategoryDescriptionTooLong = errors.New("category description too long (max 500 characters)")

	// ErrUnauthorized is returned when user is not authorized.
	ErrUnauthorized = errors.New("unauthorized")
)
