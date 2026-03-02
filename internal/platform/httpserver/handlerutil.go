// Package httpserver provides shared HTTP handler utilities.
// This file contains shared functions for building current user data
// and resolving public UUIDs to internal IDs, eliminating duplication
// across module handler packages.
package httpserver

import (
	"context"
	"fmt"
)

// CurrentUserData holds user data extracted for template rendering.
type CurrentUserData struct {
	PublicID     string
	Username     string
	Email        string
	AvatarURL    string
	PostCount    int
	CommentCount int
}

// UserLookup provides user lookup capabilities for handler utilities.
// Handler types implement this interface to supply user data without
// coupling the platform layer to module-specific domain types.
type UserLookup interface {
	// LookupUser returns user data by internal ID, or nil if not found.
	LookupUser(ctx context.Context, userID int) (*CurrentUserData, error)
	// LookupInternalID resolves a public UUID to an internal database ID.
	LookupInternalID(ctx context.Context, publicID string) (int, error)
	// LookupReactionCount returns the total reaction count for a user.
	LookupReactionCount(ctx context.Context, userID int) (int, error)
}

// BuildCurrentUser fetches full user info (including cached stats) and returns
// a map suitable for templates. It always returns a map (never nil).
func BuildCurrentUser(ctx context.Context, userID int, lookup UserLookup) map[string]interface{} {
	emptyUser := map[string]interface{}{
		"PublicID":      "",
		"Username":      "",
		"Email":         "",
		"AvatarURL":     "",
		"PostCount":     0,
		"CommentCount":  0,
		"ReactionCount": 0,
	}

	data, err := lookup.LookupUser(ctx, userID)
	if err != nil || data == nil {
		return emptyUser
	}

	reactionCount := 0
	if count, err := lookup.LookupReactionCount(ctx, userID); err == nil {
		reactionCount = count
	}

	return map[string]interface{}{
		"PublicID":      data.PublicID,
		"Username":      data.Username,
		"Email":         data.Email,
		"AvatarURL":     data.AvatarURL,
		"PostCount":     data.PostCount,
		"CommentCount":  data.CommentCount,
		"ReactionCount": reactionCount,
	}
}

// GetInternalUserID converts a PublicID (UUID) from context to internal INT ID.
// SECURITY: Ensures public UUID is never exposed, only used for lookups.
func GetInternalUserID(ctx context.Context, userPublicID string, lookup UserLookup) (int, error) {
	if userPublicID == "" {
		return 0, fmt.Errorf("user ID required")
	}

	id, err := lookup.LookupInternalID(ctx, userPublicID)
	if err != nil {
		return 0, fmt.Errorf("user not found")
	}

	return id, nil
}
