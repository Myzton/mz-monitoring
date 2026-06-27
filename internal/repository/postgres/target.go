package postgres

import (
	"context"
	"log/slog"
	"mz-monitoring/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TargetRepository struct {
	pool *pgxpool.Pool
}

func NewTargetRepository(pool *pgxpool.Pool) *TargetRepository {
	return &TargetRepository{pool: pool}
}

func (r *TargetRepository) Create(ctx context.Context, target *domain.Target) error {
	query := `INSERT INTO targets (user_id, url, interval_sec, is_active) VALUES ($1, $2, $3, $4)`
	_, err := r.pool.Exec(ctx, query, target.UserID, target.URL, target.IntervalSec, target.IsActive)
	if err != nil {
		slog.Error("An error occurred.", "error", err)
		return err
	}
	return nil
}

func (r *TargetRepository) GetList(ctx context.Context, userID int) ([]domain.Target, error) {
	query := `SELECT id, user_id, url, interval_sec, is_active FROM targets WHERE user_id = $1`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		slog.Error("An error occurred.", "error", err)
		return nil, err
	}
	var targets []domain.Target
	defer rows.Close()

	for rows.Next() {
		var t domain.Target

		err := rows.Scan(&t.ID, &t.UserID, &t.URL, &t.IntervalSec, &t.IsActive)
		if err != nil {
			slog.Error("failed to scan node row", "error", err)
			return nil, err
		}
		targets = append(targets, t)
	}
	if err = rows.Err(); err != nil {
		slog.Error("error during rows iteration", "error", err)
		return nil, err
	}
	return targets, nil
}

func (r *TargetRepository) Delete(ctx context.Context, id int, userID int) error {
	query := `DELETE FROM targets WHERE id = $1 AND user_id= $2`
	_, err := r.pool.Exec(ctx, query, id, userID)
	if err != nil {
		slog.Error("An error occurred.", "error", err)
		return err
	}
	return nil
}
func (r *TargetRepository) GetById(ctx context.Context, id int, userID int) (*domain.Target, error) {
	var target domain.Target
	query := `SELECT id, user_id, url, interval_sec, is_active FROM targets WHERE id= $1 AND user_id= $2`
	err := r.pool.QueryRow(ctx, query, id, userID).Scan(&target.ID, &target.UserID, &target.URL, &target.IntervalSec, &target.IsActive)
	if err != nil {
		slog.Error("error during rows iteration", "error", err)
		return nil, err
	}
	return &target, nil

}
func (r *TargetRepository) GetActive(ctx context.Context) ([]domain.Target, error) {
	query := `SELECT id, user_id, url, interval_sec, is_active FROM targets WHERE is_active = $1`
	rows, err := r.pool.Query(ctx, query, true)
	if err != nil {
		slog.Error("error during rows iteration", "error", err)
		return nil, err
	}
	var targets []domain.Target
	defer rows.Close()

	for rows.Next() {
		var t domain.Target
		err := rows.Scan(&t.ID, &t.UserID, &t.URL, &t.IntervalSec, &t.IsActive)
		if err != nil {
			slog.Error("failed to scan node row", "error", err)
			return nil, err
		}
		targets = append(targets, t)
	}
	if err = rows.Err(); err != nil {
		slog.Error("error during rows iteration", "error", err)
		return nil, err
	}
	return targets, nil
}
