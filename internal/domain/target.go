package domain

import "context"

type Target struct {
	ID          int
	UserID      int
	URL         string
	IntervalSec int
	IsActive    bool
}

type TargetRepository interface {
	Create(ctx context.Context, target *Target) error
	GetList(ctx context.Context, userID int) ([]Target, error)
	Delete(ctx context.Context, id int, userID int) error
	GetById(ctx context.Context, id int, userID int) (*Target, error)
	GetActive(ctx context.Context) ([]Target, error)
}
