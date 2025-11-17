// Package domain contains the core business entities for the user module.
// Domain entities represent users, roles, and related concepts.
package domain

import (
	"time"
)

// User represents a forum user.
type User struct {
	ID           int       `json:"-"`                    // Internal unique identifier (INT PRIMARY KEY) - never expose
	PublicID     string    `json:"id"`                   // Public UUID identifier (exposed in API)
	Email        string    `json:"email"`                // User's email address (unique)
	Username     string    `json:"username"`             // User's display name (unique)
	PasswordHash string    `json:"-"`                    // Hashed password - never expose
	Role         Role      `json:"role"`                 // User's role (Guest, User, Moderator, Admin)
	CreatedAt    time.Time `json:"created_at"`           // Account creation timestamp
	UpdatedAt    time.Time `json:"updated_at"`           // Last update timestamp
	IsActive     bool      `json:"is_active"`            // Account active status
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
	UserID       int       `json:"-"`                    // Internal ID - never expose
	PublicUserID string    `json:"id"`                   // Public UUID identifier
	Username     string    `json:"username"`             // User's display name
	Role         Role      `json:"role"`                 // User's role
	PostCount    int       `json:"post_count"`           // Number of posts created
	CommentCount int       `json:"comment_count"`        // Number of comments made
	CreatedAt    time.Time `json:"created_at"`           // Account creation timestamp
}
