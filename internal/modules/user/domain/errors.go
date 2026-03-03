// Package domain contains error definitions for the user module.
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

	// ErrInvalidPublicID is returned when the public identifier is empty or invalid.
	ErrInvalidPublicID = errors.New("invalid public id")
)

// PasswordValidationError provides specific feedback about which password
// criteria were not met.
type PasswordValidationError struct {
	Message string
}

func (e *PasswordValidationError) Error() string { return e.Message }

// Is allows errors.Is(err, ErrWeakPassword) to return true for PasswordValidationError.
func (e *PasswordValidationError) Is(target error) bool {
	return target == ErrWeakPassword
}

// IsPasswordValidationError checks whether the given error is a PasswordValidationError.
func IsPasswordValidationError(err error) bool {
	var pve *PasswordValidationError
	return errors.As(err, &pve)
}
