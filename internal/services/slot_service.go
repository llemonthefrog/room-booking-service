package services

import (
	"avito-task/internal/contracts/repository"
	"avito-task/internal/domain"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type SlotsService struct {
	slotsRepo    repository.SlotRepository
	scheduleRepo repository.ScheduleRepository
	roomRepo     repository.RoomRepository
}

func NewSlotsService(
	slotsRepo repository.SlotRepository,
	scheduleRepo repository.ScheduleRepository,
	roomRepo repository.RoomRepository,
) *SlotsService {
	return &SlotsService{
		slotsRepo:    slotsRepo,
		scheduleRepo: scheduleRepo,
		roomRepo:     roomRepo,
	}
}

func (s *SlotsService) GetSlotsByRoomId(ctx context.Context, room uuid.UUID, date time.Time) ([]*domain.Slot, error) {
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.AddDate(0, 0, 1)

	if _, err := s.roomRepo.GetByID(ctx, room); err != nil {
		return nil, domain.ErrNotFound
	}

	sch, err := s.scheduleRepo.GetByRoomID(ctx, room)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return []*domain.Slot{}, nil
		}

		return nil, err
	}

	if !sch.IsDayInSchedule(dayStart) {
		return []*domain.Slot{}, nil
	}

	existingSlots, err := s.slotsRepo.GetAvailable(ctx, room, dayStart, dayEnd)
	if err != nil {
		return nil, err
	}

	if len(existingSlots) == 0 {
		generated, err := s.generateAndSaveSlots(ctx, room, dayStart, sch)
		if err != nil {
			return nil, err
		}

		return generated, nil
	}

	available := make([]*domain.Slot, 0)
	for _, sl := range existingSlots {
		if !sl.IsBooked {
			available = append(available, sl)
		}
	}

	return available, nil
}

func (s *SlotsService) generateAndSaveSlots(
	ctx context.Context,
	roomID uuid.UUID,
	day time.Time,
	sch *domain.Schedule,
) ([]*domain.Slot, error) {
	var newSlots []*domain.Slot

	current := time.Date(
		day.Year(),
		day.Month(),
		day.Day(),
		sch.StartTime.Hour(),
		sch.StartTime.Minute(),
		0,
		0,
		time.UTC,
	)

	endTime := time.Date(
		day.Year(),
		day.Month(),
		day.Day(),
		sch.EndTime.Hour(),
		sch.EndTime.Minute(),
		0,
		0,
		time.UTC,
	)

	for current.Before(endTime) {
		slotEnd := current.Add(30 * time.Minute)

		slotID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("%s-%d", roomID, current.Unix())))

		newSlots = append(newSlots, &domain.Slot{
			Id:       slotID,
			RoomId:   roomID,
			StartAt:  current,
			EndAt:    slotEnd,
			IsBooked: false,
		})
		current = slotEnd
	}

	if err := s.slotsRepo.SaveBatch(ctx, newSlots); err != nil {
		return nil, err
	}

	return newSlots, nil
}
