package postgres

import (
	"context"
	"mz-monitoring/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type LogRepository struct {
	pool *pgxpool.Pool
}

func NewLogRepository(pool *pgxpool.Pool) *LogRepository {
	return &LogRepository{pool: pool}
}

func (r *LogRepository) SaveLog(ctx context.Context, check *domain.CheckLog) error {
	query := `INSERT INTO check_logs (target_id, status_code, response_time_ms, is_up, checked_at) 
              VALUES ($1, $2, $3, $4, $5) RETURNING id`

	err := r.pool.QueryRow(ctx, query,
		check.TargetID,
		check.Status,
		check.ResponseTime.Milliseconds(),
		check.Flag,
		check.CreatedAt,
	).Scan(&check.ID)

	if err != nil {
		return err
	}
	return nil
}
