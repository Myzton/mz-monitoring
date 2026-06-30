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
