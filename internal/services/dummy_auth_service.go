package services

import (
	"avito-task/internal/contracts/repository"
	"avito-task/internal/domain"
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	DummyAdminId = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	DummyUserId  = uuid.MustParse("22222222-2222-2222-2222-222222222222")
)

type DummyAuthService struct {
	secretKey []byte
	userRepo  repository.UserRepository
}

func NewDummyAuthService(secret string, userRepository repository.UserRepository) *DummyAuthService {
	return &DummyAuthService{
		secretKey: []byte(secret),
		userRepo:  userRepository,
	}
}

func (s *DummyAuthService) Login(ctx context.Context, roleStr string) (string, error) {
	role := domain.Role(roleStr)
	var userID uuid.UUID
	var email string

	switch role {
	case domain.RoleAdmin:
		userID = DummyAdminId
		email = "admin@itmo.ru"
	case domain.RoleUser:
		userID = DummyUserId
		email = "user@itmo.ru"
	default:
		return "", domain.ErrNotFound
	}

	testUser := &domain.User{
		Id:           userID,
		Email:        email,
		Role:         role,
		PasswordHash: "some_dummy_password",
		CreatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, testUser); err != nil {
		if !errors.Is(err, domain.ErrAlreadyExists) {
			return "", err
		}
		existing, err := s.userRepo.GetById(ctx, userID)
		if err != nil {
			return "", err
		}
		if existing.Role != role {
			return "", domain.ErrInvalidData
		}
	}

	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"role":    string(role),
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}
