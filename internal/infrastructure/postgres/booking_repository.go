package postgres

import (
	"avito-task/internal/domain"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type BookingPostgresRepository struct {
	db *sql.DB
}

func NewBookingPostgresRepository(db *sql.DB) *BookingPostgresRepository {
	return &BookingPostgresRepository{
		db: db,
	}
}

func (r *BookingPostgresRepository) Create(ctx context.Context, b *domain.Book) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}

	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `
		UPDATE slots 
		SET is_booked = true 
		WHERE id = $1 AND is_booked = false`,
		b.SlotId,
	)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			if pgErr.Code == "23505" {
				return domain.ErrAlreadyExists
			}
		}
		return fmt.Errorf("failed to insert booking: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrConflict
	}

	query := `
		INSERT INTO bookings (id, user_id, slot_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = tx.ExecContext(ctx, query,
		b.Id,
		b.UserId,
		b.SlotId,
		b.Status,
		b.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert booking: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}

	return nil
}

func (r *BookingPostgresRepository) Cancel(ctx context.Context, bookingID uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var slotID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		UPDATE bookings 
		SET status = 'cancelled' 
		WHERE id = $1 AND status = 'created'
		RETURNING slot_id`,
		bookingID,
	).Scan(&slotID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}

	_, err = tx.ExecContext(ctx, "UPDATE slots SET is_booked = false WHERE id = $1", slotID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *BookingPostgresRepository) GetByUserId(ctx context.Context, userID uuid.UUID) ([]*domain.Book, error) {
	query := `
       SELECT b.id, b.user_id, b.slot_id, b.status, b.created_at
       FROM bookings b
       JOIN slots s ON b.slot_id = s.id
       WHERE b.user_id = $1 
         AND b.status = 'created'
         AND s.start_at >= $2
       ORDER BY s.start_at ASC
    `

	rows, err := r.db.QueryContext(ctx, query, userID, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanBookings(rows)
}

func (r *BookingPostgresRepository) GetById(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
	b := &domain.Book{}
	query := `SELECT id, user_id, slot_id, status, created_at FROM bookings WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&b.Id, &b.UserId, &b.SlotId, &b.Status, &b.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return b, nil
}

func (r *BookingPostgresRepository) GetAll(ctx context.Context, limit, offset int) ([]*domain.Book, int, error) {
	var total int
	countQuery := "SELECT COUNT(*) FROM bookings"
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
       SELECT id, user_id, slot_id, status, created_at
       FROM bookings
       ORDER BY created_at DESC
       LIMIT $1 OFFSET $2
    `

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	bookings, err := r.scanBookings(rows)
	return bookings, total, err
}

func (r *BookingPostgresRepository) scanBookings(rows *sql.Rows) ([]*domain.Book, error) {
	bookings := make([]*domain.Book, 0)
	for rows.Next() {
		b := &domain.Book{}
		err := rows.Scan(&b.Id, &b.UserId, &b.SlotId, &b.Status, &b.CreatedAt)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, b)
	}
	return bookings, nil
}
