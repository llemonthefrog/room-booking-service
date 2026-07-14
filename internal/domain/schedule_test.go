package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewSchedule(t *testing.T) {
	roomID := uuid.New()
	validStart := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	validEnd := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		days    []int
		start   time.Time
		end     time.Time
		wantErr bool
	}{
		{
			name:    "Valid schedule",
			days:    []int{1, 2, 3},
			start:   validStart,
			end:     validEnd,
			wantErr: false,
		},
		{
			name:    "Empty days",
			days:    []int{},
			start:   validStart,
			end:     validEnd,
			wantErr: true,
		},
		{
			name:    "Invalid day range (0)",
			days:    []int{0},
			start:   validStart,
			end:     validEnd,
			wantErr: true,
		},
		{
			name:    "Duplicate days",
			days:    []int{1, 1},
			start:   validStart,
			end:     validEnd,
			wantErr: true,
		},
		{
			name:    "Start >= End",
			days:    []int{1},
			start:   validEnd,
			end:     validStart,
			wantErr: true,
		},
		{
			name:    "Not 30 min aligned",
			days:    []int{1},
			start:   time.Date(2026, 1, 1, 10, 15, 0, 0, time.UTC),
			end:     validEnd,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := NewSchedule(roomID, tt.days, tt.start, tt.end)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSchedule() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && s == nil {
				t.Error("expected schedule, got nil")
			}
		})
	}
}

func TestSchedule_IsDayInSchedule(t *testing.T) {
	start := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	s, _ := NewSchedule(uuid.New(), []int{1, 3, 5}, start, end)

	tests := []struct {
		name string
		date time.Time // Год-Месяц-День
		want bool
	}{
		{
			name: "Monday is in",
			date: time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "Tuesday is not in",
			date: time.Date(2026, 4, 7, 0, 0, 0, 0, time.UTC),
			want: false,
		},
		{
			name: "Sunday is not in (boundary check)",
			date: time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.IsDayInSchedule(tt.date); got != tt.want {
				t.Errorf("IsDayInSchedule() = %v, want %v", got, tt.want)
			}
		})
	}
}
