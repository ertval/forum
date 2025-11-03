// Package domain contains error definitions for the post module.
package domain

import "errors"

var (
    ErrPostNotFound = errors.New("post not found")
    ErrCategoryNotFound = errors.New("category not found")
    ErrInvalidImage = errors.New("invalid image file")
)
