package services

import (
	"avito-task/internal/contracts/repository"
	"avito-task/internal/domain"
	"context"
)

type RoomService struct {
	repo repository.RoomRepository
}

func NewRoomService(repo repository.RoomRepository) *RoomService {
	return &RoomService{
		repo: repo,
	}
}

func (s *RoomService) Create(ctx context.Context, name string) (*domain.Room, error) {
	if name == "" {
		return nil, domain.ErrInvalidData
	}

	newRoom, err := domain.NewRoom(name)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, newRoom); err != nil {
		return nil, domain.ErrDomainValidation
	}

	return newRoom, nil
}

func (s *RoomService) GetList(ctx context.Context) ([]*domain.Room, error) {
	rooms, err := s.repo.GetAll(ctx)

	if err != nil {
		return nil, domain.ErrDomainValidation
	}

	return rooms, nil
}
