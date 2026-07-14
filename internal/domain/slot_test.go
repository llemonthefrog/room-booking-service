package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewSlot(t *testing.T) {
	roomID := uuid.New()
	startTime := time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC)

	slot, err := NewSlot(roomID, startTime)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if slot.RoomId != roomID {
		t.Errorf("expected roomID %v, got %v", roomID, slot.RoomId)
	}

	expectedEnd := startTime.Add(30 * time.Minute)
	if !slot.EndAt.Equal(expectedEnd) {
		t.Errorf("expected end time %v, got %v", expectedEnd, slot.EndAt)
	}

	if slot.IsBooked {
		t.Error("new slot should not be booked")
	}

	if slot.Id == uuid.Nil {
		t.Error("expected generated UUID, got nil")
	}
}

func TestSlot_BookAndRelease(t *testing.T) {
	slot, _ := NewSlot(uuid.New(), time.Now())

	err := slot.Book()
	if err != nil {
		t.Errorf("expected no error on first booking, got %v", err)
	}
	if !slot.IsBooked {
		t.Error("slot should be marked as booked")
	}

	err = slot.Book()
	if err == nil {
		t.Error("expected error when booking already booked slot, got nil")
	}

	slot.Release()
	if slot.IsBooked {
		t.Error("slot should be free after release")
	}

	err = slot.Book()
	if err != nil {
		t.Errorf("expected no error after release, got %v", err)
	}
}
