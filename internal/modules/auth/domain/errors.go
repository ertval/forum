// Package domain contains the core business entities for the auth module.
package domain

import "errors"

// Domain errors for the auth module.
var (
	// ErrInvalidCredentials is returned when login credentials are incorrect.
	ErrInvalidCredentials = errors.New("invalid email or password")

	// ErrSessionNotFound is returned when a session doesn't exist.
	ErrSessionNotFound = errors.New("session not found")

	// ErrSessionExpired is returned when a session has expired.
	ErrSessionExpired = errors.New("session has expired")

	// ErrInvalidSession is returned when a session is invalid.
	ErrInvalidSession = errors.New("invalid session")

	// ErrEmailAlreadyExists is returned when trying to register with an existing email.
	ErrEmailAlreadyExists = errors.New("user with this email already exists")

	// ErrUsernameAlreadyExists is returned when trying to register with an existing username.
	ErrUsernameAlreadyExists = errors.New("user with this username already exists")

	// ErrInvalidEmail is returned when email format is invalid.
	ErrInvalidEmail = errors.New("invalid email format")

	// ErrWeakPassword is returned when password doesn't meet requirements.
	ErrWeakPassword = errors.New("password doesn't meet security requirements")

	// ErrInvalidUsername is returned when username format is invalid.
	ErrInvalidUsername = errors.New("invalid username: must start with a capital letter and contain only letters (e.g., Alice or Alice Smith)")
)
