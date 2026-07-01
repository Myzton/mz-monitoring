package usecase

import (
	"context"
	"errors"
	"mz-monitoring/internal/domain"
	url2 "net/url"
)

type TargetUsecase struct {
	targetRepo  domain.TargetRepository
	statusCache domain.StatusCache
}

func NewTargetUsecase(r domain.TargetRepository, cache domain.StatusCache) *TargetUsecase {
	return &TargetUsecase{
		targetRepo:  r,
		statusCache: cache,
	}
}

func (t *TargetUsecase) Create(ctx context.Context, userID int, url string, intervalSec int) (*domain.Target, error) {
	target := &domain.Target{UserID: userID, URL: url, IntervalSec: intervalSec, IsActive: true}

	if intervalSec != 10 && intervalSec != 30 && intervalSec != 60 {
		return nil, errors.New("invalid value for interval")
	}

	_, err := url2.ParseRequestURI(url)
	if err != nil {
		return nil, err
	}

	err = t.targetRepo.Create(ctx, target)
	if err != nil {
		return nil, err
	}
	return target, nil
}

func (t *TargetUsecase) Delete(ctx context.Context, id int, userID int) error {
	_, err := t.targetRepo.GetById(ctx, id, userID)
	if err != nil {
		return err
	}

	err = t.targetRepo.Delete(ctx, id, userID)
	if err != nil {
		return err
	}
	return nil
}

func (t *TargetUsecase) GetList(ctx context.Context, userID int) ([]domain.Target, error) {
	targets, err := t.targetRepo.GetList(ctx, userID)
	if err != nil {
		return nil, err
	}

	for i := range targets {
		isOnline, err := t.statusCache.GetStatus(ctx, targets[i].ID)
		if err != nil {
			targets[i].IsOnline = false
			continue
		}
		targets[i].IsOnline = isOnline
	}

	return targets, nil
}
