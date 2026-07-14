package usecase

import (
	"avito-task/internal/domain"
	"context"

	"github.com/google/uuid"
)

type ScheduleUseCase interface {
	Create(ctx context.Context, roomID uuid.UUID, days []int, start, end string) (*domain.Schedule, error)
}
