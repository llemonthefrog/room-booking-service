package postgres

import (
	"avito-task/internal/domain"
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

type RoomPostgresRepository struct {
	db *sql.DB
}

func NewRoomPostgresRepository(db *sql.DB) *RoomPostgresRepository {
	return &RoomPostgresRepository{
		db: db,
	}
}

func (r *RoomPostgresRepository) Create(ctx context.Context, room *domain.Room) error {
	query := `INSERT INTO rooms (id, name, created_at) VALUES ($1, $2, $3)`

	_, err := r.db.ExecContext(ctx, query, room.Id, room.Name, room.CreatedAt)

	return err
}

func (r *RoomPostgresRepository) GetAll(ctx context.Context) ([]*domain.Room, error) {
	query := `SELECT id, name, created_at FROM rooms`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rooms := make([]*domain.Room, 0)

	for rows.Next() {
		room := &domain.Room{}
		if err := rows.Scan(&room.Id, &room.Name, &room.CreatedAt); err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rooms, nil
}

func (r *RoomPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Room, error) {
	room := &domain.Room{}

	query := `SELECT id, name, created_at FROM rooms WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(&room.Id, &room.Name, &room.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return room, nil
}
