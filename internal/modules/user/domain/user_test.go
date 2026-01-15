package domain

import (
	"testing"
	"time"
)

func TestUser_HasPermission(t *testing.T) {
	tests := []struct {
		name       string
		role       Role
		action     string
		wantResult bool
	}{
		// Admin has all permissions
		{"admin can view", RoleAdmin, PermissionViewContent, true},
		{"admin can create post", RoleAdmin, PermissionCreatePost, true},
		{"admin can create comment", RoleAdmin, PermissionCreateComment, true},
		{"admin can react", RoleAdmin, PermissionReact, true},
		{"admin can edit own", RoleAdmin, PermissionEditOwn, true},
		{"admin can delete own", RoleAdmin, PermissionDeleteOwn, true},
		{"admin can edit any", RoleAdmin, PermissionEditAny, true},
		{"admin can delete any", RoleAdmin, PermissionDeleteAny, true},
		{"admin can moderate", RoleAdmin, PermissionModerate, true},
		{"admin can manage users", RoleAdmin, PermissionManageUsers, true},
		{"admin can manage categories", RoleAdmin, PermissionManageCategories, true},

		// Moderator permissions
		{"moderator can view", RoleModerator, PermissionViewContent, true},
		{"moderator can create post", RoleModerator, PermissionCreatePost, true},
		{"moderator can create comment", RoleModerator, PermissionCreateComment, true},
		{"moderator can react", RoleModerator, PermissionReact, true},
		{"moderator can edit own", RoleModerator, PermissionEditOwn, true},
		{"moderator can delete own", RoleModerator, PermissionDeleteOwn, true},
		{"moderator can edit any", RoleModerator, PermissionEditAny, true},
		{"moderator can delete any", RoleModerator, PermissionDeleteAny, true},
		{"moderator can moderate", RoleModerator, PermissionModerate, true},
		{"moderator cannot manage users", RoleModerator, PermissionManageUsers, false},
		{"moderator cannot manage categories", RoleModerator, PermissionManageCategories, false},

		// User permissions
		{"user can view", RoleUser, PermissionViewContent, true},
		{"user can create post", RoleUser, PermissionCreatePost, true},
		{"user can create comment", RoleUser, PermissionCreateComment, true},
		{"user can react", RoleUser, PermissionReact, true},
		{"user can edit own", RoleUser, PermissionEditOwn, true},
		{"user can delete own", RoleUser, PermissionDeleteOwn, true},
		{"user cannot edit any", RoleUser, PermissionEditAny, false},
		{"user cannot delete any", RoleUser, PermissionDeleteAny, false},
		{"user cannot moderate", RoleUser, PermissionModerate, false},
		{"user cannot manage users", RoleUser, PermissionManageUsers, false},
		{"user cannot manage categories", RoleUser, PermissionManageCategories, false},

		// Guest permissions (view only)
		{"guest can view", RoleGuest, PermissionViewContent, true},
		{"guest cannot create post", RoleGuest, PermissionCreatePost, false},
		{"guest cannot create comment", RoleGuest, PermissionCreateComment, false},
		{"guest cannot react", RoleGuest, PermissionReact, false},
		{"guest cannot edit own", RoleGuest, PermissionEditOwn, false},
		{"guest cannot delete own", RoleGuest, PermissionDeleteOwn, false},
		{"guest cannot edit any", RoleGuest, PermissionEditAny, false},
		{"guest cannot delete any", RoleGuest, PermissionDeleteAny, false},
		{"guest cannot moderate", RoleGuest, PermissionModerate, false},
		{"guest cannot manage users", RoleGuest, PermissionManageUsers, false},
		{"guest cannot manage categories", RoleGuest, PermissionManageCategories, false},

		// Unknown action returns false for non-admin
		{"user unknown action", RoleUser, "unknown_action", false},
		{"moderator unknown action", RoleModerator, "unknown_action", false},
		{"guest unknown action", RoleGuest, "unknown_action", false},
		{"admin allows unknown action", RoleAdmin, "unknown_action", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Role: tt.role}
			result := user.HasPermission(tt.action)
			if result != tt.wantResult {
				t.Errorf("HasPermission(%q) for role %q = %v, want %v",
					tt.action, tt.role, result, tt.wantResult)
			}
		})
	}
}

func TestUser_CanModerate(t *testing.T) {
	tests := []struct {
		name     string
		role     Role
		expected bool
	}{
		{
			name:     "admin can moderate",
			role:     RoleAdmin,
			expected: true,
		},
		{
			name:     "moderator can moderate",
			role:     RoleModerator,
			expected: true,
		},
		{
			name:     "user cannot moderate",
			role:     RoleUser,
			expected: false,
		},
		{
			name:     "guest cannot moderate",
			role:     RoleGuest,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Role: tt.role}
			result := user.CanModerate()
			if result != tt.expected {
				t.Errorf("CanModerate() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestUser_IsAdmin(t *testing.T) {
	tests := []struct {
		name     string
		role     Role
		expected bool
	}{
		{
			name:     "admin is admin",
			role:     RoleAdmin,
			expected: true,
		},
		{
			name:     "moderator is not admin",
			role:     RoleModerator,
			expected: false,
		},
		{
			name:     "user is not admin",
			role:     RoleUser,
			expected: false,
		},
		{
			name:     "guest is not admin",
			role:     RoleGuest,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Role: tt.role}
			result := user.IsAdmin()
			if result != tt.expected {
				t.Errorf("IsAdmin() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestUser_StructFields(t *testing.T) {
	now := time.Now()
	user := &User{
		ID:           1,
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashed_password",
		Role:         RoleUser,
		CreatedAt:    now,
		UpdatedAt:    now,
		IsActive:     true,
	}

	if user.ID != 1 {
		t.Errorf("Expected ID 1, got %d", user.ID)
	}
	if user.Email != "test@example.com" {
		t.Errorf("Expected Email 'test@example.com', got '%s'", user.Email)
	}
	if user.Username != "testuser" {
		t.Errorf("Expected Username 'testuser', got '%s'", user.Username)
	}
	if user.PasswordHash != "hashed_password" {
		t.Errorf("Expected PasswordHash 'hashed_password', got '%s'", user.PasswordHash)
	}
	if user.Role != RoleUser {
		t.Errorf("Expected Role '%s', got '%s'", RoleUser, user.Role)
	}
	if !user.CreatedAt.Equal(now) {
		t.Errorf("Expected CreatedAt %v, got %v", now, user.CreatedAt)
	}
	if !user.UpdatedAt.Equal(now) {
		t.Errorf("Expected UpdatedAt %v, got %v", now, user.UpdatedAt)
	}
	if !user.IsActive {
		t.Error("Expected IsActive to be true")
	}
}

func TestRoleConstants(t *testing.T) {
	if RoleGuest != "guest" {
		t.Errorf("Expected RoleGuest to be 'guest', got '%s'", RoleGuest)
	}
	if RoleUser != "user" {
		t.Errorf("Expected RoleUser to be 'user', got '%s'", RoleUser)
	}
	if RoleModerator != "moderator" {
		t.Errorf("Expected RoleModerator to be 'moderator', got '%s'", RoleModerator)
	}
	if RoleAdmin != "admin" {
		t.Errorf("Expected RoleAdmin to be 'admin', got '%s'", RoleAdmin)
	}
}

func TestUser_PostAndCommentCounts(t *testing.T) {
	now := time.Now()
	user := &User{
		ID:           1,
		PublicID:     "user-uuid-1",
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashed_password",
		Role:         RoleUser,
		PostCount:    5,
		CommentCount: 10,
		CreatedAt:    now,
		UpdatedAt:    now,
		IsActive:     true,
	}

	if user.PostCount != 5 {
		t.Errorf("Expected PostCount 5, got %d", user.PostCount)
	}
	if user.CommentCount != 10 {
		t.Errorf("Expected CommentCount 10, got %d", user.CommentCount)
	}
}
