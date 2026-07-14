package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewBook(t *testing.T) {
	userID := uuid.New()
	slotID := uuid.New()

	book, err := NewBook(userID, slotID)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if book.UserId != userID {
		t.Errorf("expected userID %v, got %v", userID, book.UserId)
	}

	if book.SlotId != slotID {
		t.Errorf("expected slotID %v, got %v", slotID, book.SlotId)
	}

	if book.Status != StatusCreated {
		t.Errorf("expected status %s, got %s", StatusCreated, book.Status)
	}

	if book.Id == uuid.Nil {
		t.Error("expected generated UUID, got nil")
	}
}

func TestBook_IsAllowed(t *testing.T) {
	ownerID := uuid.New()
	strangerID := uuid.New()

	book := &Book{
		UserId: ownerID,
	}

	tests := []struct {
		name    string
		userID  uuid.UUID
		wantErr bool
	}{
		{
			name:    "Access allowed for owner",
			userID:  ownerID,
			wantErr: false,
		},
		{
			name:    "Access denied for stranger",
			userID:  strangerID,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := book.IsAllowed(tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsAllowed() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != ErrDomainValidation {
				t.Errorf("expected ErrDomainValidation, got %v", err)
			}
		})
	}
}
