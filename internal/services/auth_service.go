package services

import (
	"avito-task/internal/contracts/repository"
	"avito-task/internal/domain"
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	secretKey []byte
	userRepo  repository.UserRepository
}

func NewAuthService(secretKey string, userRepo repository.UserRepository) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		secretKey: []byte(secretKey),
	}
}

func (s *AuthService) Register(
	ctx context.Context,
	email,
	password string,
	role domain.Role,
) (*domain.User, string, error) {
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	newUser, err := domain.NewUser(email, string(hashBytes), role)
	if err != nil {
		return nil, "", err
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, "", err
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", err
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return "", domain.ErrInvalidCredentials
		}
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", domain.ErrInvalidCredentials
	}

	return s.generateToken(user)
}

func (s *AuthService) generateToken(user *domain.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.Id.String(),
		"role":    string(user.Role),
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}
