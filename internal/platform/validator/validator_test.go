package validator

import (
	"testing"
)

func TestUsernameValidation(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		// Valid single names
		{"valid single name", "Alice", false},
		{"valid lowercase name", "alice", false},
		{"valid single name with mixed case", "McDonald", false},
		{"valid two char name", "Li", false},

		// Valid full names (multi-word)
		{"valid full name", "Alice Smith", false},
		{"valid full name with mixed case", "John McDonald", false},
		{"valid triple name", "Alice Mary Jane", false},

		// Valid - handle-style with digits, hyphens, underscores
		{"with digits", "alice123", false},
		{"with underscore", "alice_smith", false},
		{"with hyphen", "Alice-Smith", false},

		// Invalid - doesn't start with letter
		{"starts with digit", "123alice", true},
		{"starts with underscore", "_alice", true},

		// Invalid - length
		{"too short", "A", true},
		{"empty", "", true},
		{"whitespace only", "   ", true},

		// Invalid - special characters not allowed
		{"with apostrophe", "O'Brien", true},
		{"with at sign", "alice@bob", true},

		// Edge cases
		{"fifty chars valid", "Abcdefghijklmnopqrstuvwxyz Abcdefghijklmnopqrstu", false},
		{"fifty one chars invalid", "Abcdefghijklmnopqrstuvwxyz Abcdefghijklmnopqrstuvxy", true},
		{"leading spaces", "  Alice", false},  // TrimSpace should handle
		{"trailing spaces", "Alice  ", false}, // TrimSpace should handle
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			v.Username("username", tt.username)

			hasError := !v.Valid()
			if hasError != tt.wantErr {
				t.Errorf("Username(%q): got error=%v, want error=%v. Errors: %v",
					tt.username, hasError, tt.wantErr, v.Errors())
			}
		})
	}
}
