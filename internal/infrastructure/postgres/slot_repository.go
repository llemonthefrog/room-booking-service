package postgres

import (
	"avito-task/internal/domain"
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type SlotPostgresRepository struct {
	db *sql.DB
}

func NewSlotPostgresRepository(db *sql.DB) *SlotPostgresRepository {
	return &SlotPostgresRepository{
		db: db,
	}
}

func (r *SlotPostgresRepository) GetById(ctx context.Context, id uuid.UUID) (*domain.Slot, error) {
	s := &domain.Slot{}

	query := `
		SELECT id, room_id, start_at, end_at, is_booked 
		FROM slots 
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&s.Id, &s.RoomId, &s.StartAt, &s.EndAt, &s.IsBooked,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return s, nil
}

func (r *SlotPostgresRepository) GetAvailable(ctx context.Context, roomID uuid.UUID, from, to time.Time) ([]*domain.Slot, error) {
	query := `
		SELECT id, room_id, start_at, end_at, is_booked 
		FROM slots 
		WHERE room_id = $1 
		  AND start_at >= $2 
		  AND start_at < $3 
		  AND is_booked = false
		ORDER BY start_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, roomID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	slots := make([]*domain.Slot, 0)

	for rows.Next() {
		s := &domain.Slot{}
		err := rows.Scan(&s.Id, &s.RoomId, &s.StartAt, &s.EndAt, &s.IsBooked)
		if err != nil {
			return nil, err
		}
		slots = append(slots, s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return slots, nil
}

func (r *SlotPostgresRepository) SaveBatch(ctx context.Context, slots []*domain.Slot) error {
	if len(slots) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO slots (id, room_id, start_at, end_at, is_booked) 
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (room_id, start_at) DO NOTHING
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, s := range slots {
		_, err := stmt.ExecContext(ctx, s.Id, s.RoomId, s.StartAt, s.EndAt, s.IsBooked)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *SlotPostgresRepository) UpdateStatus(ctx context.Context, id uuid.UUID, isBooked bool) error {

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.ErrConflict
	}

	defer tx.Rollback()

	query := `
       UPDATE slots 
       SET is_booked = $1 
       WHERE id = $2 AND is_booked != $1
    `

	_, err = tx.ExecContext(ctx, query, isBooked, id)
	if err != nil {
		return domain.ErrConflict
	}

	if err := tx.Commit(); err != nil {
		return domain.ErrConflict
	}

	return nil
}
