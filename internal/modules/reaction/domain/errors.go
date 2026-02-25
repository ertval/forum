// Package domain contains error definitions for the reaction module.
package domain

import "errors"

// Domain errors for the reaction module.
var (
	// ErrReactionNotFound is returned when a reaction doesn't exist.
	ErrReactionNotFound = errors.New("reaction not found")

	// ErrDuplicateReaction is returned when a user tries to add a duplicate reaction.
	ErrDuplicateReaction = errors.New("reaction already exists")

	// ErrInvalidTarget is returned when the target type is invalid.
	ErrInvalidTarget = errors.New("invalid target type, must be 'post' or 'comment'")

	// ErrInvalidReactionType is returned when the reaction type is invalid.
	ErrInvalidReactionType = errors.New("invalid reaction type, must be 'like' or 'dislike'")

	// ErrInvalidUserID is returned when the user ID is invalid.
	ErrInvalidUserID = errors.New("invalid user ID")

	// ErrInvalidTargetID is returned when the target ID is invalid.
	ErrInvalidTargetID = errors.New("invalid target ID")
)
