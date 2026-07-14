package usecase

import "avito-task/internal/domain"

type JWTUseCase interface {
	GenerateToken(userID string, role domain.Role) (string, error)
	VerifyToken(tokenString string) (*domain.Claims, error)
}
