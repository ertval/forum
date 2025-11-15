package unit

import (
	"testing"

	"forum/internal/modules/auth/application"
	"forum/internal/modules/auth/domain"
)

func TestValidateCredentials(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		wantErr  bool
	}{
		{
			name:     "valid credentials",
			email:    "user@example.com",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "empty email",
			email:    "",
			password: "password123",
			wantErr:  true,
		},
		{
			name:     "invalid email",
			email:    "invalid-email",
			password: "password123",
			wantErr:  true,
		},
		{
			name:     "empty password",
			email:    "user@example.com",
			password: "",
			wantErr:  true,
		},
		{
			name:     "short password",
			email:    "user@example.com",
			password: "12345",
			wantErr:  true,
		},
		{
			name:     "both empty",
			email:    "",
			password: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creds := &domain.Credentials{
				Email:    tt.email,
				Password: tt.password,
			}
			err := application.ValidateCredentials(creds)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCredentials() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
