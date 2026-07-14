package services

import (
	"avito-task/internal/domain"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWTService_Flow(t *testing.T) {
	secret := "test_secret_key"
	ttl := time.Hour
	svc := NewJWTService(secret, ttl)

	userID := uuid.New().String()
	role := domain.RoleUser

	token, err := svc.GenerateToken(userID, role)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	if token == "" {
		t.Fatal("token is empty")
	}

	claims, err := svc.VerifyToken(token)
	if err != nil {
		t.Errorf("failed to verify valid token: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("expected userID %s, got %s", userID, claims.UserID)
	}
	if claims.Role != role {
		t.Errorf("expected role %s, got %s", role, claims.Role)
	}

	wrongSvc := NewJWTService("wrong_secret", ttl)
	_, err = wrongSvc.VerifyToken(token)
	if err == nil {
		t.Error("expected error for token with wrong secret, got nil")
	}

	shortSvc := NewJWTService(secret, -time.Second)
	expiredToken, _ := shortSvc.GenerateToken(userID, role)
	_, err = svc.VerifyToken(expiredToken)
	if err == nil {
		t.Error("expected error for expired token, got nil")
	}

	_, err = svc.VerifyToken("not.a.token.string")
	if err == nil {
		t.Error("expected error for invalid token format, got nil")
	}
}
