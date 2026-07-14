package services

import (
	"context"
	"errors"
	"testing"

	"avito-task/internal/domain"

	"github.com/google/uuid"
)

type mockRoomRepo struct {
	CreateFn  func(ctx context.Context, room *domain.Room) error
	GetAllFn  func(ctx context.Context) ([]*domain.Room, error)
	GetByIDFn func(ctx context.Context, id uuid.UUID) (*domain.Room, error)
}

func (m *mockRoomRepo) Create(ctx context.Context, room *domain.Room) error {
	return m.CreateFn(ctx, room)
}

func (m *mockRoomRepo) GetAll(ctx context.Context) ([]*domain.Room, error) {
	return m.GetAllFn(ctx)
}

func (m *mockRoomRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Room, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}

	return nil, nil
}

func TestRoomService_Create(t *testing.T) {
	tests := []struct {
		name        string
		roomName    string
		mockCreate  func(ctx context.Context, room *domain.Room) error
		expectError error
	}{
		{
			name:     "Success",
			roomName: "Conference Hall",
			mockCreate: func(ctx context.Context, room *domain.Room) error {
				return nil
			},
			expectError: nil,
		},
		{
			name:        "Empty Name",
			roomName:    "",
			mockCreate:  nil,
			expectError: domain.ErrInvalidData,
		},
		{
			name:     "Repository Error",
			roomName: "Small Room",
			mockCreate: func(ctx context.Context, room *domain.Room) error {
				return errors.New("db connection lost")
			},
			expectError: domain.ErrDomainValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRoomRepo{CreateFn: tt.mockCreate}
			svc := NewRoomService(repo)

			room, err := svc.Create(context.Background(), tt.roomName)

			if !errors.Is(err, tt.expectError) {
				t.Errorf("expected error %v, got %v", tt.expectError, err)
			}

			if err == nil && room == nil {
				t.Error("expected room to be created, got nil")
			}
			if err == nil && room.Name != tt.roomName {
				t.Errorf("expected room name %s, got %s", tt.roomName, room.Name)
			}
		})
	}
}

func TestRoomService_GetList(t *testing.T) {
	tests := []struct {
		name        string
		mockGetAll  func(ctx context.Context) ([]*domain.Room, error)
		expectLen   int
		expectError error
	}{
		{
			name: "Success",
			mockGetAll: func(ctx context.Context) ([]*domain.Room, error) {
				return []*domain.Room{{}, {}}, nil
			},
			expectLen:   2,
			expectError: nil,
		},
		{
			name: "Repository Error",
			mockGetAll: func(ctx context.Context) ([]*domain.Room, error) {
				return nil, errors.New("db error")
			},
			expectLen:   0,
			expectError: domain.ErrDomainValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRoomRepo{GetAllFn: tt.mockGetAll}
			svc := NewRoomService(repo)

			rooms, err := svc.GetList(context.Background())

			if !errors.Is(err, tt.expectError) {
				t.Errorf("expected error %v, got %v", tt.expectError, err)
			}

			if len(rooms) != tt.expectLen {
				t.Errorf("expected %d rooms, got %d", tt.expectLen, len(rooms))
			}
		})
	}
}
