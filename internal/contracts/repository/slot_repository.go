package repository

import (
	"avito-task/internal/domain"
	"context"
	"time"

	"github.com/google/uuid"
)

type SlotRepository interface {
	GetById(ctx context.Context, id uuid.UUID) (*domain.Slot, error)
	GetAvailable(ctx context.Context, roomID uuid.UUID, from, to time.Time) ([]*domain.Slot, error)
	SaveBatch(ctx context.Context, slots []*domain.Slot) error
	UpdateStatus(ctx context.Context, id uuid.UUID, isBooked bool) error
}
