package domain

import (
	"time"

	"github.com/google/uuid"
)

type Slot struct {
	Id       uuid.UUID
	RoomId   uuid.UUID
	StartAt  time.Time
	EndAt    time.Time
	IsBooked bool
}

func NewSlot(roomId uuid.UUID, start time.Time) (*Slot, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, ErrDomainValidation
	}

	return &Slot{
		Id:       id,
		RoomId:   roomId,
		StartAt:  start.UTC(),
		EndAt:    start.Add(30 * time.Minute).UTC(),
		IsBooked: false,
	}, nil
}

func (s *Slot) Book() error {
	if s.IsBooked {
		return ErrDomainValidation
	}
	s.IsBooked = true
	return nil
}

func (s *Slot) Release() {
	s.IsBooked = false
}
