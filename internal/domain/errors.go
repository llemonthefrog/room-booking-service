package domain

import "errors"

var (
	ErrInvalidData        = errors.New("invalid data")
	ErrDomainValidation   = errors.New("domain validation error")
	ErrNotFound           = errors.New("not found")
	ErrConflict           = errors.New("conflict")
	ErrAlreadyExists      = errors.New("already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
