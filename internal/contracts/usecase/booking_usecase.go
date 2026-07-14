package usecase

import (
	"avito-task/internal/domain"
	"context"

	"github.com/google/uuid"
)

type BookingUseCase interface {
	Create(ctx context.Context, userId, slotId uuid.UUID) (*domain.Book, error)
	Cancel(ctx context.Context, userId, bookingId uuid.UUID) (*domain.Book, error)
	GetUserBookings(ctx context.Context, userId uuid.UUID) ([]*domain.Book, error)
	GetAll(ctx context.Context, page, pageSize int) ([]*domain.Book, int, error)
}
