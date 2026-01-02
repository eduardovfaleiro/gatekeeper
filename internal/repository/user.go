package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/eduardovfaleiro/gatekeeper/internal/model"
	"github.com/lib/pq"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	UpdatePassword(ctx context.Context, userID model.ID, hashedPassword string) error
}

type postgresUserRepository struct {
	db *sql.DB
}

func (r *postgresUserRepository) Create(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users (id, email, password_hash, created_at) VALUES ($1, $2, $3, $4)`

	_, err := r.db.ExecContext(ctx, query, user.ID, user.Email, user.PasswordHash, user.CreatedAt)

	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			if pgErr.Code == "23505" {
				return ErrUniqueConstraint
			}
		}

		return fmt.Errorf("postgresUserRepository.Create (exec): %w", err)
	}

	return nil
}

func NewPostgresUserRepository(db *sql.DB) UserRepository {
	return &postgresUserRepository{db}
}

func (r *postgresUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `SELECT id, email, password_hash, created_at FROM users WHERE email = $1`

	var user model.User

	err := r.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *postgresUserRepository) UpdatePassword(ctx context.Context, userID model.ID, hashedPassword string) error {
	query := `UPDATE users SET password_hash = $1 WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, hashedPassword, userID)
	if err != nil {
		return fmt.Errorf("repository.UpdatePassword (exec): %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository.UpdatePassword (rows_affected): %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
