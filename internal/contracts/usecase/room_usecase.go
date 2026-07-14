package usecase

import (
	"avito-task/internal/domain"
	"context"
)

type RoomUseCase interface {
	Create(ctx context.Context, name string) (*domain.Room, error)
	GetList(ctx context.Context) ([]*domain.Room, error)
}
