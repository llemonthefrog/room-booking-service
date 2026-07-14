package repository

import (
	"avito-task/internal/domain"
	"context"

	"github.com/google/uuid"
)

type RoomRepository interface {
	Create(ctx context.Context, room *domain.Room) error
	GetAll(ctx context.Context) ([]*domain.Room, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Room, error)
}
