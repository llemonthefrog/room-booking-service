package domain

import (
	"time"

	"github.com/google/uuid"
)

type Schedule struct {
	Id         uuid.UUID
	RoomId     uuid.UUID
	DaysOfWeek []int
	StartTime  time.Time
	EndTime    time.Time
}

func NewSchedule(roomId uuid.UUID, days []int, start, end time.Time) (*Schedule, error) {
	if len(days) == 0 {
		return nil, ErrInvalidData
	}

	daysMap := make(map[int]struct{})
	for _, day := range days {
		if day < 1 || day > 7 {
			return nil, ErrInvalidData
		}

		if _, exists := daysMap[day]; exists {
			return nil, ErrInvalidData
		}

		daysMap[day] = struct{}{}
	}

	startMin := start.Hour()*60 + start.Minute()
	endMin := end.Hour()*60 + end.Minute()

	if startMin >= endMin {
		return nil, ErrInvalidData
	}

	if startMin%30 != 0 || endMin%30 != 0 {
		return nil, ErrInvalidData
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, ErrDomainValidation
	}

	return &Schedule{
		Id:         id,
		RoomId:     roomId,
		DaysOfWeek: days,
		StartTime:  start.UTC(),
		EndTime:    end.UTC(),
	}, nil
}

func (s *Schedule) IsDayInSchedule(t time.Time) bool {
	wd := int(t.Weekday())
	if wd == 0 {
		wd = 7
	}

	for _, d := range s.DaysOfWeek {
		if d == wd {
			return true
		}
	}

	return false
}
