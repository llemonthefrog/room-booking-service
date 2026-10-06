package services

import (
	"avito-task/internal/domain"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestCancelOtherUsersCancelledBooking(t *testing.T) {
	repo := &mockBookingRepo{GetByIdFn: func(context.Context, uuid.UUID) (*domain.Book, error) {
		return &domain.Book{UserId: uuid.New(), Status: domain.StatusCancelled}, nil
	}}
	booking, err := NewBookingService(repo, nil).Cancel(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, domain.ErrDomainValidation) || booking != nil {
		t.Fatalf("got booking %v, error %v; want access denied", booking, err)
	}
}

func TestServicesPreserveDatabaseErrors(t *testing.T) {
	dbErr := errors.New("database unavailable")
	roomRepo := &mockRoomRepoFull{GetByIDFn: func(context.Context, uuid.UUID) (*domain.Room, error) {
		return nil, dbErr
	}}
	slotRepo := &mockSlotRepo{GetByIdFn: func(context.Context, uuid.UUID) (*domain.Slot, error) {
		return nil, dbErr
	}}
	_, err := NewBookingService(nil, slotRepo).Create(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, dbErr) {
		t.Fatalf("booking error = %v", err)
	}
	_, err = NewSlotsService(nil, nil, roomRepo).GetSlotsByRoomId(context.Background(), uuid.New(), time.Now())
	if !errors.Is(err, dbErr) {
		t.Fatalf("slots error = %v", err)
	}
	_, err = NewScheduleService(nil, roomRepo).Create(context.Background(), uuid.New(), []int{1}, "09:00", "10:00")
	if !errors.Is(err, dbErr) {
		t.Fatalf("room lookup error = %v", err)
	}
	roomRepo.GetByIDFn = func(context.Context, uuid.UUID) (*domain.Room, error) { return &domain.Room{}, nil }
	schRepo := &mockScheduleRepo{GetByRoomIDFn: func(context.Context, uuid.UUID) (*domain.Schedule, error) {
		return nil, dbErr
	}}
	_, err = NewScheduleService(schRepo, roomRepo).Create(context.Background(), uuid.New(), []int{1}, "09:00", "10:00")
	if !errors.Is(err, dbErr) {
		t.Fatalf("schedule lookup error = %v", err)
	}
}

func TestFullyBookedDayReturnsNoSlots(t *testing.T) {
	roomID := uuid.New()
	date := time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)
	roomRepo := &mockRoomRepoFull{GetByIDFn: func(context.Context, uuid.UUID) (*domain.Room, error) {
		return &domain.Room{Id: roomID}, nil
	}}
	schRepo := &mockScheduleRepo{GetByRoomIDFn: func(context.Context, uuid.UUID) (*domain.Schedule, error) {
		return &domain.Schedule{DaysOfWeek: []int{1}, StartTime: date.Add(9 * time.Hour), EndTime: date.Add(10 * time.Hour)}, nil
	}}
	slotRepo := &mockSlotRepo{
		GetAvailableFn: func(context.Context, uuid.UUID, time.Time, time.Time) ([]*domain.Slot, error) {
			return []*domain.Slot{}, nil
		},
		SaveBatchFn: func(context.Context, []*domain.Slot) error { return nil },
	}
	slots, err := NewSlotsService(slotRepo, schRepo, roomRepo).GetSlotsByRoomId(context.Background(), roomID, date)
	if err != nil || len(slots) != 0 {
		t.Fatalf("slots = %v, error = %v", slots, err)
	}
}

func TestRegisterRejectsInvalidData(t *testing.T) {
	for _, tt := range []struct {
		email, password string
		role            domain.Role
	}{
		{"user@example.com", "", domain.RoleUser},
		{"user@example.com", strings.Repeat("a", 73), domain.RoleUser},
		{"invalid", "password", domain.RoleUser},
		{"user@example.com", "password", "unknown"},
	} {
		_, _, err := NewAuthService("secret", nil).Register(context.Background(), tt.email, tt.password, tt.role)
		if !errors.Is(err, domain.ErrInvalidData) {
			t.Fatalf("error = %v, want invalid data", err)
		}
	}
}

func TestJWTRejectsInvalidClaims(t *testing.T) {
	for _, tt := range []struct {
		name         string
		method       jwt.SigningMethod
		userID, role string
		expires      bool
	}{
		{"missing expiry", jwt.SigningMethodHS256, uuid.NewString(), "user", false},
		{"wrong algorithm", jwt.SigningMethodHS384, uuid.NewString(), "user", true},
		{"invalid user", jwt.SigningMethodHS256, "bad-id", "user", true},
		{"nil user", jwt.SigningMethodHS256, uuid.Nil.String(), "user", true},
		{"invalid role", jwt.SigningMethodHS256, uuid.NewString(), "unknown", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			claims := jwt.MapClaims{"user_id": tt.userID, "role": tt.role}
			if tt.expires {
				claims["exp"] = time.Now().Add(time.Hour).Unix()
			}
			token, err := jwt.NewWithClaims(tt.method, claims).SignedString([]byte("secret"))
			if err != nil {
				t.Fatal(err)
			}
			if _, err = NewJWTService("secret", time.Hour).VerifyToken(token); !errors.Is(err, domain.ErrInvalidData) {
				t.Fatalf("error = %v, want invalid data", err)
			}
		})
	}
}
