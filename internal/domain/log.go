package domain

import (
	"context"
	"time"
)

type CheckLog struct {
	ID           int
	TargetID     int
	Status       int
	ResponseTime time.Duration
	Flag         bool
	CreatedAt    time.Time
}

type LogRepository interface {
	SaveLog(ctx context.Context, check *CheckLog) error
}

type StatusCache interface {
	SetStatus(ctx context.Context, targetID int, isOnline bool) error
	GetStatus(ctx context.Context, targetID int) (bool, error)
}
