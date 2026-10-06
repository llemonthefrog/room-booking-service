package services

import (
	"avito-task/internal/domain"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type mockScheduleRepo struct {
	GetByRoomIDFn func(ctx context.Context, roomID uuid.UUID) (*domain.Schedule, error)
	SaveFn        func(ctx context.Context, sch *domain.Schedule) error
}

func (m *mockScheduleRepo) GetByRoomID(ctx context.Context, id uuid.UUID) (*domain.Schedule, error) {
	return m.GetByRoomIDFn(ctx, id)
}
func (m *mockScheduleRepo) Save(ctx context.Context, sch *domain.Schedule) error {
	return m.SaveFn(ctx, sch)
}

type mockRoomRepoFull struct {
	GetByIDFn func(ctx context.Context, id uuid.UUID) (*domain.Room, error)
}

func (m *mockRoomRepoFull) GetByID(ctx context.Context, id uuid.UUID) (*domain.Room, error) {
	return m.GetByIDFn(ctx, id)
}
func (m *mockRoomRepoFull) Create(ctx context.Context, r *domain.Room) error   { return nil }
func (m *mockRoomRepoFull) GetAll(ctx context.Context) ([]*domain.Room, error) { return nil, nil }

func TestScheduleService_Create(t *testing.T) {
	roomID := uuid.New()

	tests := []struct {
		name        string
		roomID      uuid.UUID
		days        []int
		start, end  string
		mockRoom    func(ctx context.Context, id uuid.UUID) (*domain.Room, error)
		mockSch     func(ctx context.Context, id uuid.UUID) (*domain.Schedule, error)
		expectError error
	}{
		{
			name:   "Success",
			roomID: roomID,
			days:   []int{1, 2, 3},
			start:  "09:00",
			end:    "18:00",
			mockRoom: func(ctx context.Context, id uuid.UUID) (*domain.Room, error) {
				return &domain.Room{Id: id}, nil
			},
			mockSch: func(ctx context.Context, id uuid.UUID) (*domain.Schedule, error) {
				return nil, domain.ErrNotFound
			},
			expectError: nil,
		},
		{
			name:   "Room Not Found",
			roomID: roomID,
			mockRoom: func(ctx context.Context, id uuid.UUID) (*domain.Room, error) {
				return nil, domain.ErrNotFound
			},
			expectError: domain.ErrNotFound,
		},
		{
			name:   "Schedule Already Exists",
			roomID: roomID,
			mockRoom: func(ctx context.Context, id uuid.UUID) (*domain.Room, error) {
				return &domain.Room{Id: id}, nil
			},
			mockSch: func(ctx context.Context, id uuid.UUID) (*domain.Schedule, error) {
				return &domain.Schedule{}, nil
			},
			expectError: domain.ErrAlreadyExists,
		},
		{
			name:   "Invalid Time Format",
			roomID: roomID,
			start:  "9 AM",
			end:    "18:00",
			mockRoom: func(ctx context.Context, id uuid.UUID) (*domain.Room, error) {
				return &domain.Room{Id: id}, nil
			},
			mockSch: func(ctx context.Context, id uuid.UUID) (*domain.Schedule, error) {
				return nil, domain.ErrNotFound
			},
			expectError: domain.ErrInvalidData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rRepo := &mockRoomRepoFull{GetByIDFn: tt.mockRoom}
			sRepo := &mockScheduleRepo{
				GetByRoomIDFn: tt.mockSch,
				SaveFn:        func(ctx context.Context, sch *domain.Schedule) error { return nil },
			}
			svc := NewScheduleService(sRepo, rRepo)

			_, err := svc.Create(context.Background(), tt.roomID, tt.days, tt.start, tt.end)

			if !errors.Is(err, tt.expectError) {
				t.Errorf("expected error %v, got %v", tt.expectError, err)
			}
		})
	}
}
