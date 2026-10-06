package services

import (
	"avito-task/internal/domain"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type mockUserRepo struct {
	CreateFn     func(ctx context.Context, user *domain.User) error
	GetByIdFn    func(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmailFn func(ctx context.Context, email string) (*domain.User, error)
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	return m.CreateFn(ctx, user)
}

func (m *mockUserRepo) GetById(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if m.GetByIdFn != nil {
		return m.GetByIdFn(ctx, id)
	}
	return nil, nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.GetByEmailFn != nil {
		return m.GetByEmailFn(ctx, email)
	}
	return nil, nil
}

func TestDummyAuthService_Login(t *testing.T) {
	secretKey := "super_secret_test_key"

	tests := []struct {
		name        string
		roleStr     string
		mockCreate  func(ctx context.Context, user *domain.User) error
		expectError error
		wantToken   bool
	}{
		{
			name:    "admin",
			roleStr: string(domain.RoleAdmin),
			mockCreate: func(ctx context.Context, user *domain.User) error {
				if user.Id != DummyAdminId || user.Email != "admin@itmo.ru" {
					t.Errorf("unexpected admin user data: %+v", user)
				}
				return nil
			},
			expectError: nil,
			wantToken:   true,
		},
		{
			name:    "user",
			roleStr: string(domain.RoleUser),
			mockCreate: func(ctx context.Context, user *domain.User) error {
				if user.Id != DummyUserId || user.Email != "user@itmo.ru" {
					t.Errorf("unexpected user data: %+v", user)
				}
				return nil
			},
			expectError: nil,
			wantToken:   true,
		},
		{
			name:        "Invalid Role",
			roleStr:     "guest",
			mockCreate:  nil,
			expectError: domain.ErrNotFound,
			wantToken:   false,
		},
		{
			name:    "Repository Error",
			roleStr: string(domain.RoleAdmin),
			mockCreate: func(ctx context.Context, user *domain.User) error {
				return errors.New("insert failed")
			},
			expectError: errors.New("insert failed"),
			wantToken:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepo{CreateFn: tt.mockCreate}
			svc := NewDummyAuthService(secretKey, repo)

			token, err := svc.Login(context.Background(), tt.roleStr)

			if tt.expectError != nil {
				if err == nil || err.Error() != tt.expectError.Error() {
					t.Errorf("expected error %v, got %v", tt.expectError, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.wantToken && token == "" {
				t.Error("expected token, got empty string")
			}
			if !tt.wantToken && token != "" {
				t.Errorf("expected no token, got %s", token)
			}
		})
	}
}

func TestDummyLoginExistingUser(t *testing.T) {
	for _, tt := range []struct {
		name      string
		user      *domain.User
		err       error
		wantToken bool
	}{
		{"same user", &domain.User{Id: DummyUserId, Role: domain.RoleUser}, nil, true},
		{"email belongs to another user", nil, domain.ErrNotFound, false},
		{"wrong role", &domain.User{Id: DummyUserId, Role: domain.RoleAdmin}, nil, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepo{
				CreateFn:  func(context.Context, *domain.User) error { return domain.ErrAlreadyExists },
				GetByIdFn: func(context.Context, uuid.UUID) (*domain.User, error) { return tt.user, tt.err },
			}
			token, err := NewDummyAuthService("secret", repo).Login(context.Background(), "user")
			if tt.wantToken && (err != nil || token == "") {
				t.Fatalf("token = %q, error = %v", token, err)
			}
			if !tt.wantToken && (err == nil || token != "") {
				t.Fatalf("unexpected token = %q, error = %v", token, err)
			}
		})
	}
}
