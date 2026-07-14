package repository

import (
	"avito-task/internal/domain"
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	GetById(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
}
