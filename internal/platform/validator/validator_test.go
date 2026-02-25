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
		{"valid single name with mixed case", "McDonald", false},
		{"valid two char name", "Li", false},

		// Valid full names
		{"valid full name", "Alice Smith", false},
		{"valid full name with mixed case", "John McDonald", false},
		{"valid triple name", "Alice Mary Jane", false},

		// Invalid - capitalization
		{"lowercase single name", "alice", true},
		{"lowercase second name", "Alice smith", true},
		{"all lowercase", "alice smith", true},

		// Invalid - length
		{"too short", "A", true},
		{"empty", "", true},
		{"whitespace only", "   ", true},

		// Invalid - special characters
		{"with numbers", "Alice123", true},
		{"with hyphen", "Alice-Smith", true},
		{"with apostrophe", "O'Brien", true},
		{"with underscore", "Alice_Smith", true},

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
