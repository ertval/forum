package domain

import (
	"testing"
	"time"
)

func TestReaction_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		reaction *Reaction
		expected bool
	}{
		{
			name: "valid post like",
			reaction: &Reaction{
				ID:         1,
				UserID:     5,
				TargetID:   10,
				TargetType: "post",
				Type:       ReactionLike,
				CreatedAt:  time.Now(),
			},
			expected: true,
		},
		{
			name: "valid comment dislike",
			reaction: &Reaction{
				ID:         1,
				UserID:     5,
				TargetID:   15,
				TargetType: "comment",
				Type:       ReactionDislike,
				CreatedAt:  time.Now(),
			},
			expected: true,
		},
		{
			name: "invalid target type",
			reaction: &Reaction{
				ID:         1,
				UserID:     5,
				TargetID:   10,
				TargetType: "invalid",
				Type:       ReactionLike,
				CreatedAt:  time.Now(),
			},
			expected: false,
		},
		{
			name: "empty target type",
			reaction: &Reaction{
				ID:         1,
				UserID:     5,
				TargetID:   10,
				TargetType: "",
				Type:       ReactionLike,
				CreatedAt:  time.Now(),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.reaction.IsValid()
			if result != tt.expected {
				t.Errorf("IsValid() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestReaction_StructFields(t *testing.T) {
	now := time.Now()
	reaction := &Reaction{
		ID:         1,
		UserID:     5,
		TargetID:   10,
		TargetType: "post",
		Type:       ReactionLike,
		CreatedAt:  now,
	}

	if reaction.ID != 1 {
		t.Errorf("Expected ID 1, got %d", reaction.ID)
	}
	if reaction.UserID != 5 {
		t.Errorf("Expected UserID 5, got %d", reaction.UserID)
	}
	if reaction.TargetID != 10 {
		t.Errorf("Expected TargetID 10, got %d", reaction.TargetID)
	}
	if reaction.TargetType != "post" {
		t.Errorf("Expected TargetType 'post', got '%s'", reaction.TargetType)
	}
	if reaction.Type != ReactionLike {
		t.Errorf("Expected Type '%s', got '%s'", ReactionLike, reaction.Type)
	}
	if !reaction.CreatedAt.Equal(now) {
		t.Errorf("Expected CreatedAt %v, got %v", now, reaction.CreatedAt)
	}
}

func TestReactionTypeConstants(t *testing.T) {
	if ReactionLike != "like" {
		t.Errorf("Expected ReactionLike to be 'like', got '%s'", ReactionLike)
	}
	if ReactionDislike != "dislike" {
		t.Errorf("Expected ReactionDislike to be 'dislike', got '%s'", ReactionDislike)
	}
}
