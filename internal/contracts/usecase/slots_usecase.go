package usecase

import (
	"avito-task/internal/domain"
	"context"
	"time"

	"github.com/google/uuid"
)

type SlotsUseCase interface {
	GetSlotsByRoomId(ctx context.Context, room uuid.UUID, date time.Time) ([]*domain.Slot, error)
}
