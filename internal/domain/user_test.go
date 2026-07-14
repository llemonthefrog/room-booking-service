package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewUser(t *testing.T) {
	password := "hashed_password_123"

	tests := []struct {
		name    string
		email   string
		role    Role
		wantErr bool
	}{
		{
			name:    "Valid admin user",
			email:   "admin@itmo.ru",
			role:    RoleAdmin,
			wantErr: false,
		},
		{
			name:    "Valid regular user",
			email:   "user@itmo.ru",
			role:    RoleUser,
			wantErr: false,
		},
		{
			name:    "Invalid email format",
			email:   "invalid-email",
			role:    RoleUser,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := NewUser(tt.email, password, tt.role)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewUser() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if u.Email != tt.email {
					t.Errorf("expected email %s, got %s", tt.email, u.Email)
				}
				if u.Role != tt.role {
					t.Errorf("expected role %s, got %s", tt.role, u.Role)
				}
				if u.PasswordHash != password {
					t.Error("password hash mismatch")
				}
				if u.Id == uuid.Nil {
					t.Error("expected generated UUID, got nil")
				}
			}
		})
	}
}
