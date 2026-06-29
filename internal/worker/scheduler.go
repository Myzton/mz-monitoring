package worker

import (
	"context"
	"log/slog"
	"mz-monitoring/internal/domain"
	"time"
)

type Scheduler struct {
	TargetRep domain.TargetRepository
}

func NewScheduler(d domain.TargetRepository) *Scheduler {
	return &Scheduler{d}
}

func (s *Scheduler) Start(ctx context.Context) error {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			targets, err := s.TargetRep.GetActive(ctx)
			if err != nil {
				slog.Error("Error get active sites", "error", err)
				continue
			}

			for _, target := range targets {
				slog.Info("The scheduler has locked the site for verification.", "ID", target.ID, "URL", target.URL)
			}

		}
	}

}
