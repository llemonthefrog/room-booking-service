package services

import (
	"avito-task/internal/contracts/repository"
	"avito-task/internal/domain"
	"context"
	"time"

	"github.com/google/uuid"
)

type ScheduleService struct {
	scheduleRepo repository.ScheduleRepository
	roomRepo     repository.RoomRepository
}

func NewScheduleService(
	scheduleRepo repository.ScheduleRepository,
	roomRepo repository.RoomRepository,
) *ScheduleService {
	return &ScheduleService{
		scheduleRepo: scheduleRepo,
		roomRepo:     roomRepo,
	}
}

func (s *ScheduleService) Create(
	ctx context.Context,
	roomID uuid.UUID,
	days []int,
	start, end string,
) (*domain.Schedule, error) {
	if _, err := s.roomRepo.GetByID(ctx, roomID); err != nil {
		return nil, domain.ErrNotFound
	}

	existing, _ := s.scheduleRepo.GetByRoomID(ctx, roomID)
	if existing != nil {
		return nil, domain.ErrAlreadyExists
	}

	layout := "15:04"
	startTime, err := time.Parse(layout, start)
	if err != nil {
		return nil, domain.ErrInvalidData
	}

	endTime, err := time.Parse(layout, end)
	if err != nil {
		return nil, domain.ErrInvalidData
	}

	sch, err := domain.NewSchedule(roomID, days, startTime, endTime)
	if err != nil {
		return nil, err
	}

	if err := s.scheduleRepo.Save(ctx, sch); err != nil {
		return nil, err
	}

	return sch, nil
}
