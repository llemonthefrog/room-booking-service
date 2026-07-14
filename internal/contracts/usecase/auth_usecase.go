package usecase

import (
	"avito-task/internal/domain"
	"context"
)

type AuthUseCase interface {
	Register(ctx context.Context, email, password string, role domain.Role) (*domain.User, string, error)
	Login(ctx context.Context, email, password string) (string, error)
}
