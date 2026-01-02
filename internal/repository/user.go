package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/eduardovfaleiro/gatekeeper/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	UpdatePassword(ctx context.Context, userID model.ID, hashedPassword string) error
}

type postgresUserRepository struct {
	db *sql.DB
}

func (r *postgresUserRepository) Create(ctx context.Context, user *model.User) (*model.User, error) {
	query := `INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, created_at`

	var u model.User

	err := r.db.QueryRowContext(ctx, query, user.Email, user.PasswordHash).Scan(&u.ID, &u.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	u.Email = user.Email
	u.PasswordHash = user.PasswordHash

	return &u, nil
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

	result, err := r.db.ExecContext(ctx, query, hashedPassword, userID[:])
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
