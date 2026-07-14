package services

import (
	"context"
	"testing"
	"time"

	"avito-task/internal/domain"

	"github.com/google/uuid"
)

type mockSlotRepo struct {
	GetAvailableFn func(ctx context.Context, roomID uuid.UUID, start, end time.Time) ([]*domain.Slot, error)
	SaveBatchFn    func(ctx context.Context, slots []*domain.Slot) error
	GetByIdFn      func(ctx context.Context, id uuid.UUID) (*domain.Slot, error)
	UpdateStatusFn func(ctx context.Context, id uuid.UUID, isBooked bool) error
}

func (m *mockSlotRepo) GetAvailable(ctx context.Context, rid uuid.UUID, s, e time.Time) ([]*domain.Slot, error) {
	return m.GetAvailableFn(ctx, rid, s, e)
}

func (m *mockSlotRepo) SaveBatch(ctx context.Context, slots []*domain.Slot) error {
	return m.SaveBatchFn(ctx, slots)
}

func (m *mockSlotRepo) GetById(ctx context.Context, id uuid.UUID) (*domain.Slot, error) {
	if m.GetByIdFn != nil {
		return m.GetByIdFn(ctx, id)
	}
	return nil, nil
}

func (m *mockSlotRepo) UpdateStatus(ctx context.Context, id uuid.UUID, isBooked bool) error {
	if m.UpdateStatusFn != nil {
		return m.UpdateStatusFn(ctx, id, isBooked)
	}
	return nil
}

func TestSlotsService_GetSlotsByRoomId(t *testing.T) {
	roomID := uuid.New()
	testDate := time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)

	sch := &domain.Schedule{
		RoomId:     roomID,
		DaysOfWeek: []int{1, 2, 3, 4, 5}, // Пн-Пт
		StartTime:  time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC),
		EndTime:    time.Date(0, 1, 1, 10, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name          string
		mockRoom      func(ctx context.Context, id uuid.UUID) (*domain.Room, error)
		mockSch       func(ctx context.Context, id uuid.UUID) (*domain.Schedule, error)
		mockSlots     func(ctx context.Context, rid uuid.UUID, s, e time.Time) ([]*domain.Slot, error)
		expectedCount int
		expectError   error
	}{
		{
			name: "Return Existing Available Slots",
			mockRoom: func(ctx context.Context, id uuid.UUID) (*domain.Room, error) {
				return &domain.Room{Id: id}, nil
			},
			mockSch: func(ctx context.Context, id uuid.UUID) (*domain.Schedule, error) {
				return sch, nil
			},
			mockSlots: func(ctx context.Context, rid uuid.UUID, s, e time.Time) ([]*domain.Slot, error) {
				return []*domain.Slot{
					{Id: uuid.New(), IsBooked: false},
					{Id: uuid.New(), IsBooked: true},
				}, nil
			},
			expectedCount: 1,
			expectError:   nil,
		},
		{
			name: "Generate New Slots If None Exist",
			mockRoom: func(ctx context.Context, id uuid.UUID) (*domain.Room, error) {
				return &domain.Room{Id: id}, nil
			},
			mockSch: func(ctx context.Context, id uuid.UUID) (*domain.Schedule, error) {
				return sch, nil
			},
			mockSlots: func(ctx context.Context, rid uuid.UUID, s, e time.Time) ([]*domain.Slot, error) {
				return []*domain.Slot{}, nil
			},

			expectedCount: 2,
			expectError:   nil,
		},
		{
			name: "Empty List If Day Not In Schedule",
			mockRoom: func(ctx context.Context, id uuid.UUID) (*domain.Room, error) {
				return &domain.Room{Id: id}, nil
			},
			mockSch: func(ctx context.Context, id uuid.UUID) (*domain.Schedule, error) {
				return sch, nil
			},
			mockSlots: func(ctx context.Context, rid uuid.UUID, s, e time.Time) ([]*domain.Slot, error) {
				return []*domain.Slot{}, nil
			},

			expectedCount: 0,
			expectError:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rRepo := &mockRoomRepoFull{GetByIDFn: tt.mockRoom}
			sRepo := &mockScheduleRepo{GetByRoomIDFn: tt.mockSch}
			slRepo := &mockSlotRepo{
				GetAvailableFn: tt.mockSlots,
				SaveBatchFn:    func(ctx context.Context, slots []*domain.Slot) error { return nil },
			}

			svc := NewSlotsService(slRepo, sRepo, rRepo)

			currentDate := testDate
			if tt.name == "Empty List If Day Not In Schedule" {
				currentDate = time.Date(2026, 4, 12, 0, 0, 0, 0, time.UTC)
			}

			slots, err := svc.GetSlotsByRoomId(context.Background(), roomID, currentDate)

			if err != tt.expectError {
				t.Errorf("expected error %v, got %v", tt.expectError, err)
			}
			if len(slots) != tt.expectedCount {
				t.Errorf("expected %d slots, got %d", tt.expectedCount, len(slots))
			}
		})
	}
}
