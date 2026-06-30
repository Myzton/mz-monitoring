package domain

import "context"

type CheckTask struct {
	ID  int
	URL string
}

type QueuePublisher interface {
	Publish(ctx context.Context, check *CheckTask) error
}
