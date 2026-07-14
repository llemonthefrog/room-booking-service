package domain

import (
	"time"

	"github.com/google/uuid"
)

type Room struct {
	Id        uuid.UUID
	Name      string
	CreatedAt time.Time
}

func NewRoom(name string) (*Room, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, ErrDomainValidation
	}

	return &Room{
		Id:        id,
		Name:      name,
		CreatedAt: time.Now().UTC(),
	}, nil
}
