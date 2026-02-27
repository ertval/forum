// Package domain contains error definitions for the comment module.
package domain

import "errors"

// Domain errors for the comment module.
var (
	// ErrCommentNotFound is returned when a comment doesn't exist.
	ErrCommentNotFound = errors.New("comment not found")

	// ErrUnauthorizedEdit is returned when a user tries to edit someone else's comment.
	ErrUnauthorizedEdit = errors.New("unauthorized to edit comment")

	// ErrEmptyContent is returned when comment content is empty.
	ErrEmptyContent = errors.New("comment content cannot be empty")

	// ErrContentTooLong is returned when comment exceeds maximum length.
	ErrContentTooLong = errors.New("comment content exceeds maximum length")
)
