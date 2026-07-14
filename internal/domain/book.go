package domain

import (
	"time"

	"github.com/google/uuid"
)

type BookStatus string

const (
	StatusCreated   BookStatus = "created"
	StatusCancelled BookStatus = "cancelled"
)

type Book struct {
	Id        uuid.UUID
	UserId    uuid.UUID
	SlotId    uuid.UUID
	Status    BookStatus
	CreatedAt time.Time
}

func NewBook(userId, slotId uuid.UUID) (*Book, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, ErrDomainValidation
	}

	return &Book{
		Id:        id,
		UserId:    userId,
		SlotId:    slotId,
		Status:    StatusCreated,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (b *Book) IsAllowed(userId uuid.UUID) error {
	if b.UserId != userId {
		return ErrDomainValidation
	}

	return nil
}
