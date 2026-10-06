package postgres

import (
	"avito-task/internal/domain"
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
)

type SchedulePostgresRepository struct {
	db *sql.DB
}

func NewSchedulePostgresRepository(db *sql.DB) *SchedulePostgresRepository {
	return &SchedulePostgresRepository{
		db: db,
	}
}

func (r *SchedulePostgresRepository) Save(ctx context.Context, s *domain.Schedule) error {
	query := `
		INSERT INTO schedules (id, room_id, days_of_week, start_time, end_time)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.ExecContext(ctx, query,
		s.Id,
		s.RoomId,
		pq.Array(s.DaysOfWeek),
		s.StartTime,
		s.EndTime,
	)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return domain.ErrAlreadyExists
	}
	return err
}

func (r *SchedulePostgresRepository) GetByRoomID(ctx context.Context, roomID uuid.UUID) (*domain.Schedule, error) {
	s := &domain.Schedule{}
	var days []int64
	var startStr, endStr string

	query := `SELECT id, room_id, days_of_week, start_time, end_time FROM schedules WHERE room_id = $1`

	err := r.db.QueryRowContext(ctx, query, roomID).Scan(
		&s.Id,
		&s.RoomId,
		pq.Array(&days),
		&startStr,
		&endStr,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	layout := "15:04:05"
	s.StartTime, _ = time.Parse(layout, startStr)
	s.EndTime, _ = time.Parse(layout, endStr)

	s.DaysOfWeek = make([]int, len(days))
	for i, v := range days {
		s.DaysOfWeek[i] = int(v)
	}

	return s, nil
}
