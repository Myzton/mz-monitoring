package redis

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type StatusRepository struct {
	client *redis.Client
}

func NewStatusRepository(client *redis.Client) *StatusRepository {
	return &StatusRepository{client: client}
}

func (r *StatusRepository) SetStatus(ctx context.Context, targetID int, isOnline bool) error {
	key := fmt.Sprintf("target:status:%d", targetID)

	err := r.client.Set(ctx, key, isOnline, 0).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *StatusRepository) GetStatus(ctx context.Context, targetID int) (bool, error) {
	key := fmt.Sprintf("target:status:%d", targetID)

	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}

	isOnline, err := strconv.ParseBool(val)
	if err != nil {
		return false, err
	}

	return isOnline, nil
}
