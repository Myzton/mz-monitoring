package worker

import (
	"context"
	"log/slog"
	"mz-monitoring/internal/domain"
	"time"
)

type Scheduler struct {
	TargetRep domain.TargetRepository
	QueuePub  domain.QueuePublisher
}

func NewScheduler(t domain.TargetRepository, q domain.QueuePublisher) *Scheduler {
	return &Scheduler{t, q}
}

func (s *Scheduler) Start(ctx context.Context) error {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	var elapsedSeconds int

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			elapsedSeconds += 10
			if elapsedSeconds > 60 {
				elapsedSeconds = 10
			}

			targets, err := s.TargetRep.GetActive(ctx)
			if err != nil {
				slog.Error("Error get active sites", "error", err)
				continue
			}

			for _, target := range targets {
				if elapsedSeconds%target.IntervalSec == 0 {
					check := domain.CheckTask{target.ID, target.URL}
					slog.Info("Time to check target!", "ID", target.ID, "Interval", target.IntervalSec, "URL", target.URL)
					err := s.QueuePub.Publish(ctx, &check)
					if err != nil {
						slog.Error("failed to publish check task", "error", err)
					}

				}
			}
		}
	}
}
