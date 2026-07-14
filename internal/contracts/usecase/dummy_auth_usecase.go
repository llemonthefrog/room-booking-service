package usecase

import "context"

type DummyAuthUseCase interface {
	Login(ctx context.Context, roleStr string) (string, error)
}
