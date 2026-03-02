package unit

import (
	"testing"
	"time"

	authDomain "forum/internal/modules/auth/domain"
	userDomain "forum/internal/modules/user/domain"
)

func TestSessionValidityRules(t *testing.T) {
	t.Run("valid session requires ids and unexpired timestamp", func(t *testing.T) {
		session := &authDomain.Session{
			ID:        10,
			UserID:    42,
			ExpiresAt: time.Now().Add(5 * time.Minute),
		}

		if session.IsExpired() {
			t.Fatal("expected non-expired session")
		}
		if !session.IsValid() {
			t.Fatal("expected session to be valid")
		}
	})

	t.Run("expired session is invalid", func(t *testing.T) {
		session := &authDomain.Session{
			ID:        10,
			UserID:    42,
			ExpiresAt: time.Now().Add(-1 * time.Minute),
		}

		if !session.IsExpired() {
			t.Fatal("expected expired session")
		}
		if session.IsValid() {
			t.Fatal("expected expired session to be invalid")
		}
	})

	t.Run("missing ids makes session invalid", func(t *testing.T) {
		session := &authDomain.Session{ExpiresAt: time.Now().Add(5 * time.Minute)}
		if session.IsValid() {
			t.Fatal("expected session with zero IDs to be invalid")
		}
	})
}

func TestUserRolePermissions(t *testing.T) {
	tests := []struct {
		name       string
		role       userDomain.Role
		permission userDomain.Permission
		allowed    bool
	}{
		{name: "guest can view", role: userDomain.RoleGuest, permission: userDomain.PermissionViewContent, allowed: true},
		{name: "guest cannot create", role: userDomain.RoleGuest, permission: userDomain.PermissionCreatePost, allowed: false},
		{name: "user can create post", role: userDomain.RoleUser, permission: userDomain.PermissionCreatePost, allowed: true},
		{name: "user cannot moderate", role: userDomain.RoleUser, permission: userDomain.PermissionModerate, allowed: false},
		{name: "moderator can delete any", role: userDomain.RoleModerator, permission: userDomain.PermissionDeleteAny, allowed: true},
		{name: "admin can manage users", role: userDomain.RoleAdmin, permission: userDomain.PermissionManageUsers, allowed: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u := &userDomain.User{Role: tc.role}
			if got := u.HasPermission(tc.permission); got != tc.allowed {
				t.Fatalf("permission mismatch for role=%s permission=%s: got=%v want=%v", tc.role, tc.permission, got, tc.allowed)
			}
		})
	}
}
