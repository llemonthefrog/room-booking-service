package services

import (
	"avito-task/internal/contracts/repository"
	"avito-task/internal/domain"
	"context"

	"github.com/google/uuid"
)

type BookingService struct {
	bookingRepo repository.BookingRepository
	slotRepo    repository.SlotRepository
}

func NewBookingService(
	bookingRepo repository.BookingRepository,
	slotRepo repository.SlotRepository,
) *BookingService {
	return &BookingService{
		bookingRepo: bookingRepo,
		slotRepo:    slotRepo,
	}
}

func (s *BookingService) Create(ctx context.Context, userId, slotId uuid.UUID) (*domain.Book, error) {
	_, err := s.slotRepo.GetById(ctx, slotId)
	if err != nil {
		return nil, err
	}

	booking, err := domain.NewBook(userId, slotId)
	if err != nil {
		return nil, err
	}

	if err := s.bookingRepo.Create(ctx, booking); err != nil {
		return nil, err
	}

	return booking, nil
}

func (s *BookingService) Cancel(ctx context.Context, userId, bookingId uuid.UUID) (*domain.Book, error) {
	booking, err := s.bookingRepo.GetById(ctx, bookingId)
	if err != nil {
		return nil, err
	}

	if err := booking.IsAllowed(userId); err != nil {
		return nil, err
	}
	if booking.Status == domain.StatusCancelled {
		return booking, nil
	}

	if err := s.bookingRepo.Cancel(ctx, bookingId); err != nil {
		return nil, err
	}

	booking.Status = domain.StatusCancelled

	return booking, nil
}

func (s *BookingService) GetUserBookings(ctx context.Context, userID uuid.UUID) ([]*domain.Book, error) {
	return s.bookingRepo.GetByUserId(ctx, userID)
}

func (s *BookingService) GetAll(ctx context.Context, limit, offset int) ([]*domain.Book, int, error) {
	return s.bookingRepo.GetAll(ctx, limit, offset)
}
