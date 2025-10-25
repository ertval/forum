package models

// user.go defines the User model and related database operations.
// It handles user creation, authentication, and retrieval.

import (
	"time"
)

// User represents a registered forum user
type User struct {
	ID           int
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// CreateUser inserts a new user into the database
// Password should already be hashed before calling this function
func CreateUser(username, email, passwordHash string) (*User, error) {
	// Insert user into database
	// Return created user with ID
	return nil, nil
}

// GetUserByID retrieves a user by their ID
func GetUserByID(id int) (*User, error) {
	// Query user from database by ID
	return nil, nil
}

// GetUserByEmail retrieves a user by their email address
func GetUserByEmail(email string) (*User, error) {
	// Query user from database by email
	return nil, nil
}

// GetUserByUsername retrieves a user by their username
func GetUserByUsername(username string) (*User, error) {
	// Query user from database by username
	return nil, nil
}

// ValidatePassword checks if the provided password matches the user's hashed password
func (u *User) ValidatePassword(password string) bool {
	// Compare password with stored hash using bcrypt
	return false
}
