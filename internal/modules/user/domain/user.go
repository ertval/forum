package domain
// Package domain contains the core user business logic and entities.
package domain

import "time"

// Role represents user roles in the forum.
type Role string

const (
	RoleGuest     Role = "guest"
	RoleUser      Role = "user"
	RoleModerator Role = "moderator"
	RoleAdmin     Role = "admin"
)

// User represents a user entity.
type User struct {
	ID        string
	Email     string
	Username  string
	Password  string // Hashed password
	Role      Role
	CreatedAt time.Time
	UpdatedAt time.Time

	// OAuth fields
	OAuthProvider   string
	OAuthProviderID string
}

// IsAdmin checks if the user is an administrator.
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// IsModerator checks if the user is a moderator or admin.
func (u *User) IsModerator() bool {
	return u.Role == RoleModerator || u.Role == RoleAdmin
}

// CanModerate checks if the user can moderate content.
func (u *User) CanModerate() bool {
	return u.IsModerator()
}
