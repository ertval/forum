// Package domain - errors
// This file contains domain-specific errors for the auth module.
package domain

import "errors"

var (
	// ErrInvalidCredentials is returned when login credentials are incorrect.
	ErrInvalidCredentials = errors.New("invalid email or password")

	// ErrSessionNotFound is returned when a session is not found.
	ErrSessionNotFound = errors.New("session not found")

	// ErrSessionExpired is returned when a session has expired.
	ErrSessionExpired = errors.New("session has expired")

	// ErrInvalidToken is returned when a session token is invalid.
	ErrInvalidToken = errors.New("invalid session token")

	// ErrUserNotFound is returned when a user is not found during authentication.
	ErrUserNotFound = errors.New("user not found")

	// ErrEmailAlreadyExists is returned when registering with an existing email.
	ErrEmailAlreadyExists = errors.New("email already exists")

	// ErrUsernameAlreadyExists is returned when registering with an existing username.
	ErrUsernameAlreadyExists = errors.New("username already exists")

	// ErrWeakPassword is returned when the password doesn't meet security requirements.
	ErrWeakPassword = errors.New("password does not meet security requirements")

	// ErrInvalidEmail is returned when the email format is invalid.
	ErrInvalidEmail = errors.New("invalid email format")

	// ErrOAuthProviderError is returned when there's an error with OAuth provider.
	ErrOAuthProviderError = errors.New("OAuth provider error")
)
