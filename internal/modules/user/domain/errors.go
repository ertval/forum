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

	// ErrInvalidEmail is returned when email format is invalid.
	ErrInvalidEmail = errors.New("invalid email")

	// ErrInvalidUsername is returned when username format is invalid.
	ErrInvalidUsername = errors.New("invalid username")

	// ErrWeakPassword is returned when password does not meet requirements.
	ErrWeakPassword = errors.New("password must be at least 8 characters")

	// ErrEmailAlreadyExists is returned when another user already has the email.
	ErrEmailAlreadyExists = errors.New("email already exists")

	// ErrUsernameAlreadyExists is returned when another user already has the username.
	ErrUsernameAlreadyExists = errors.New("username already exists")
)
