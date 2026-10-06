package services

import (
	"avito-task/internal/domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type customClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

type JWTService struct {
	secretKey string
	timeLimit time.Duration
}

func NewJWTService(secretKey string, timeLimit time.Duration) *JWTService {
	return &JWTService{
		secretKey: secretKey,
		timeLimit: timeLimit,
	}
}

func (j *JWTService) GenerateToken(userID string, role domain.Role) (string, error) {
	claims := customClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.timeLimit)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: userID,
		Role:   string(role),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

func (j *JWTService) VerifyToken(tokenString string) (*domain.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &customClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(j.secretKey), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())

	if err != nil {
		return nil, domain.ErrInvalidData
	}

	claims, ok := token.Claims.(*customClaims)
	if !ok || !token.Valid {
		return nil, domain.ErrInvalidData
	}
	userID, err := uuid.Parse(claims.UserID)
	if err != nil || userID == uuid.Nil || (claims.Role != string(domain.RoleUser) && claims.Role != string(domain.RoleAdmin)) {
		return nil, domain.ErrInvalidData
	}

	return &domain.Claims{
		UserID: claims.UserID,
		Role:   domain.Role(claims.Role),
	}, nil
}
