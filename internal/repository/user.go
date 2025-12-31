package repository

import (
	"context"
	"database/sql"

	"github.com/eduardovfaleiro/gatekeeper/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
}

type postgresUserRepository struct {
	db *sql.DB
}

func (r *postgresUserRepository) Create(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, created_at`

	row := r.db.QueryRowContext(ctx, query, user.Email, user.PasswordHash).Scan(&user.ID, &user.CreatedAt)
	return row
}

func NewPostgresUserRepository(db *sql.DB) UserRepository {
	return &postgresUserRepository{db}
}
