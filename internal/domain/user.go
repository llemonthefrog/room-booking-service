package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type User struct {
	Id           uuid.UUID
	Email        string
	Role         Role
	PasswordHash string
	CreatedAt    time.Time
}

func NewUser(email string, passwordHash string, role Role) (*User, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, ErrDomainValidation
	}

	if !strings.Contains(email, "@") {
		return nil, ErrInvalidData
	}

	return &User{
		Id:           id,
		Email:        email,
		Role:         role,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now().UTC(),
	}, nil
}
