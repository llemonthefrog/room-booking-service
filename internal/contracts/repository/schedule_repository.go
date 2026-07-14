package repository

import (
	"avito-task/internal/domain"
	"context"

	"github.com/google/uuid"
)

type ScheduleRepository interface {
	GetByRoomID(ctx context.Context, roomID uuid.UUID) (*domain.Schedule, error)
	Save(ctx context.Context, s *domain.Schedule) error
}
