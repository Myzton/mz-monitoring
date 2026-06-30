package rabbitmq

import (
	"context"
	"encoding/json"
	"mz-monitoring/internal/domain"

	"github.com/rabbitmq/amqp091-go"
)

type RabbitPublisher struct {
	ch *amqp091.Channel
}

func NewRabbitPublisher(ch *amqp091.Channel) *RabbitPublisher {
	return &RabbitPublisher{ch: ch}
}

func (r *RabbitPublisher) Publish(ctx context.Context, check *domain.CheckTask) error {
	byteCheck, err := json.Marshal(check)
	if err != nil {

		return err
	}
	err = r.ch.PublishWithContext(ctx,
		"",
		"monitoring_tasks",
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        byteCheck,
		},
	)
	if err != nil {

		return err
	}
	return nil
}
