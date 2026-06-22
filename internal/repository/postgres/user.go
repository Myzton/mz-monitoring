package postgres

import (
	"context"
	"errors"
	"log/slog"
	"mz-monitoring/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users ( name, password, email) VALUES ($1, $2, $3)`
	_, err := r.pool.Exec(ctx, query, user.Name, user.Password, user.Email)
	if err != nil {
		slog.Error("An error occurred.", "error", err)
		return err
	}
	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var id int
	var name, password string

	query := `SELECT id, name, password, email FROM users WHERE email = $1`
	err := r.pool.QueryRow(ctx, query, email).Scan(&id, &name, &password, &email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		slog.Error("An error occurred.", "error", err)
		return nil, err
	}
	user := domain.User{id, name, password, email}
	return &user, nil
}
