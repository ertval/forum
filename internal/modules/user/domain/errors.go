// Package domain contains the core business entities for the user module.
package domain

import "errors"

// Domain errors for the user module.
var (
	// ErrUserNotFound is returned when a user doesn't exist.
	ErrUserNotFound = errors.New("user not found")

	// ErrInvalidRole is returned when an invalid role is specified.
	ErrInvalidRole = errors.New("invalid role")

	// ErrUnauthorized is returned when a user lacks permission.
	ErrUnauthorized = errors.New("unauthorized action")

	// ErrUserInactive is returned when accessing an inactive user account.
	ErrUserInactive = errors.New("user account is inactive")

	// ErrCannotDemoteAdmin is returned when trying to demote the last admin.
	ErrCannotDemoteAdmin = errors.New("cannot demote the last administrator")
)
