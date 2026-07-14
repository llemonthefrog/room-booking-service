package services

import (
	"avito-task/internal/domain"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_Login(t *testing.T) {
	secret := "super_secret_test_key"
	email := "test@itmo.ru"
	password := "password"

	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	existingUser := &domain.User{
		Id:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		Role:         domain.RoleUser,
	}

	tests := []struct {
		name        string
		mockGet     func(ctx context.Context, email string) (*domain.User, error)
		password    string
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Success login",
			mockGet: func(ctx context.Context, email string) (*domain.User, error) {
				return existingUser, nil
			},
			password: password,
			wantErr:  false,
		},
		{
			name: "User not found",
			mockGet: func(ctx context.Context, email string) (*domain.User, error) {
				return nil, domain.ErrNotFound
			},
			password:    password,
			wantErr:     true,
			expectedErr: domain.ErrInvalidCredentials,
		},
		{
			name: "Wrong password",
			mockGet: func(ctx context.Context, email string) (*domain.User, error) {
				return existingUser, nil
			},
			password:    "wrong_pass",
			wantErr:     true,
			expectedErr: domain.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepo{GetByEmailFn: tt.mockGet}
			svc := NewAuthService(secret, repo)

			token, err := svc.Login(context.Background(), email, tt.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
			if !tt.wantErr && token == "" {
				t.Error("expected token, got empty string")
			}
		})
	}
}

func TestAuthService_Register(t *testing.T) {
	secret := "super_secret_test_key"
	email := "test@itmo.ru"
	password := "password"

	repo := &mockUserRepo{
		CreateFn: func(ctx context.Context, user *domain.User) error {
			return nil
		},
		GetByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
			hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			return &domain.User{Id: uuid.New(), Email: email, PasswordHash: string(hash), Role: domain.RoleUser}, nil
		},
	}

	svc := NewAuthService(secret, repo)
	user, token, err := svc.Register(context.Background(), email, password, domain.RoleUser)

	if err != nil {
		t.Errorf("Register() unexpected error: %v", err)
	}
	if user.Email != email {
		t.Errorf("expected email %s, got %s", email, user.Email)
	}
	if token == "" {
		t.Error("expected token, got empty string")
	}
}
