package services

import (
	"avito-task/internal/domain"
	"context"
	"testing"

	"github.com/google/uuid"
)

type mockBookingRepo struct {
	CreateFn      func(ctx context.Context, b *domain.Book) error
	GetByIdFn     func(ctx context.Context, id uuid.UUID) (*domain.Book, error)
	CancelFn      func(ctx context.Context, id uuid.UUID) error
	GetByUserIdFn func(ctx context.Context, id uuid.UUID) ([]*domain.Book, error)
	GetAllFn      func(ctx context.Context, limit, offset int) ([]*domain.Book, int, error)
}

func (m *mockBookingRepo) Create(ctx context.Context, b *domain.Book) error {
	return m.CreateFn(ctx, b)
}
func (m *mockBookingRepo) GetById(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
	return m.GetByIdFn(ctx, id)
}
func (m *mockBookingRepo) Cancel(ctx context.Context, id uuid.UUID) error { return m.CancelFn(ctx, id) }
func (m *mockBookingRepo) GetByUserId(ctx context.Context, id uuid.UUID) ([]*domain.Book, error) {
	return m.GetByUserIdFn(ctx, id)
}
func (m *mockBookingRepo) GetAll(ctx context.Context, l, o int) ([]*domain.Book, int, error) {
	return m.GetAllFn(ctx, l, o)
}

func TestBookingService_Cancel(t *testing.T) {
	userID := uuid.New()
	bookingID := uuid.New()

	tests := []struct {
		name           string
		setupMock      func() *mockBookingRepo
		wantErr        bool
		expectedStatus domain.BookStatus
	}{
		{
			name: "Success Cancel",
			setupMock: func() *mockBookingRepo {
				return &mockBookingRepo{
					GetByIdFn: func(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
						return &domain.Book{Id: id, UserId: userID, Status: domain.StatusCreated}, nil
					},
					CancelFn: func(ctx context.Context, id uuid.UUID) error { return nil },
				}
			},
			wantErr:        false,
			expectedStatus: domain.StatusCancelled,
		},
		{
			name: "Idempotent Cancel (Already Cancelled)",
			setupMock: func() *mockBookingRepo {
				return &mockBookingRepo{
					GetByIdFn: func(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
						return &domain.Book{Id: id, UserId: userID, Status: domain.StatusCancelled}, nil
					},
				}
			},
			wantErr:        false,
			expectedStatus: domain.StatusCancelled,
		},
		{
			name: "Forbidden (Wrong User)",
			setupMock: func() *mockBookingRepo {
				return &mockBookingRepo{
					GetByIdFn: func(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
						return &domain.Book{Id: id, UserId: uuid.New(), Status: domain.StatusCreated}, nil
					},
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewBookingService(tt.setupMock(), &mockSlotRepo{})
			res, err := s.Cancel(context.Background(), userID, bookingID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Cancel() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && res.Status != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, res.Status)
			}
		})
	}
}

func TestBookingService_Create(t *testing.T) {
	userID := uuid.New()
	slotID := uuid.New()

	bookingRepo := &mockBookingRepo{
		CreateFn: func(ctx context.Context, b *domain.Book) error { return nil },
	}
	slotRepo := &mockSlotRepo{
		GetByIdFn: func(ctx context.Context, id uuid.UUID) (*domain.Slot, error) {
			return &domain.Slot{Id: id}, nil
		},
	}

	s := NewBookingService(bookingRepo, slotRepo)
	res, err := s.Create(context.Background(), userID, slotID)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if res.UserId != userID || res.SlotId != slotID {
		t.Error("booking data mismatch")
	}
}
