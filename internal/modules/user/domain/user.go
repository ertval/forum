// Package domain contains the core business entities for the user module.
// Domain entities represent users, roles, and related concepts.
package domain

import (
	"time"
)

// User represents a forum user.
type User struct {
	ID           int       // Internal unique identifier (INT PRIMARY KEY)
	PublicID     string    // Public UUID identifier (exposed in API)
	Email        string    // User's email address (unique)
	Username     string    // User's display name (unique)
	PasswordHash string    // Hashed password
	Role         Role      // User's role (Guest, User, Moderator, Admin)
	CreatedAt    time.Time // Account creation timestamp
	UpdatedAt    time.Time // Last update timestamp
	IsActive     bool      // Account active status
}

// Role represents a user's permission level.
type Role string

const (
	// RoleGuest represents an unregistered visitor (view-only access).
	RoleGuest Role = "guest"

	// RoleUser represents a registered user (can create, comment, react).
	RoleUser Role = "user"

	// RoleModerator represents a moderator (can monitor, delete, report).
	RoleModerator Role = "moderator"

	// RoleAdmin represents an administrator (full system access).
	RoleAdmin Role = "admin"
)

// HasPermission checks if the user has permission for an action based on their role.
// TODO: Implement permission logic.
func (u *User) HasPermission(action string) bool {
	// Implementation placeholder
	// Define permissions for each role
	return false
}

// CanModerate checks if the user can perform moderation actions.
func (u *User) CanModerate() bool {
	return u.Role == RoleModerator || u.Role == RoleAdmin
}

// IsAdmin checks if the user is an administrator.
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// UserProfile represents a user's public profile information.
type UserProfile struct {
	UserID       int
	Username     string
	Role         Role
	PostCount    int // Number of posts created
	CommentCount int // Number of comments made
	CreatedAt    time.Time
}
