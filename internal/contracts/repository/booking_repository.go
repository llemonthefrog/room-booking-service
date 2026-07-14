package repository

import (
	"avito-task/internal/domain"
	"context"

	"github.com/google/uuid"
)

type BookingRepository interface {
	Create(ctx context.Context, b *domain.Book) error
	Cancel(ctx context.Context, bookingID uuid.UUID) error
	GetByUserId(ctx context.Context, userID uuid.UUID) ([]*domain.Book, error)
	GetById(ctx context.Context, id uuid.UUID) (*domain.Book, error)
	GetAll(ctx context.Context, limit, offset int) ([]*domain.Book, int, error)
}
