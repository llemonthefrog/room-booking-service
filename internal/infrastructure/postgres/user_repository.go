package postgres

import (
	"avito-task/internal/domain"
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type UserPostgresRepository struct {
	db *sql.DB
}

func NewUserPostgresRepository(db *sql.DB) *UserPostgresRepository {
	return &UserPostgresRepository{
		db: db,
	}
}

func (repo *UserPostgresRepository) GetById(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user := &domain.User{}

	query := `SELECT id, email, role, password_hash, created_at FROM users WHERE id = $1`

	err := repo.db.QueryRowContext(ctx, query, id).Scan(
		&user.Id,
		&user.Email,
		&user.Role,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return user, nil
}

func (repo *UserPostgresRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user := &domain.User{}

	query := `SELECT id, email, password_hash, role, created_at FROM users WHERE email = $1`

	err := repo.db.QueryRowContext(ctx, query, email).Scan(
		&user.Id,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return user, nil
}

func (repo *UserPostgresRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
       INSERT INTO users (id, email, password_hash, role, created_at)
       VALUES ($1, $2, $3, $4, $5)
    `
	_, err := repo.db.ExecContext(ctx, query,
		user.Id,
		user.Email,
		user.PasswordHash,
		user.Role,
		time.Now().UTC(),
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrAlreadyExists
		}
		return err
	}

	return nil
}
