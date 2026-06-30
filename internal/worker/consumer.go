package worker

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"mz-monitoring/internal/domain"
	"net/http"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	ch *amqp091.Channel
}

func NewConsumer(ch *amqp091.Channel) *Consumer {
	return &Consumer{ch: ch}
}

func (c *Consumer) Start(ctx context.Context) error {
	msgs, err := c.ch.Consume(
		"monitoring_tasks",
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				return errors.New("rabbitmq channel closed")
			}
			// Убрали лишний slog.Info отсюда
			c.processTask(ctx, msg)
		}
	}
}

func (c *Consumer) processTask(ctx context.Context, msg amqp091.Delivery) {
	var task domain.CheckTask
	if err := json.Unmarshal(msg.Body, &task); err != nil {
		slog.Error("Failed to unmarshal task", "error", err)
		return
	}

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	startTime := time.Now()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, task.URL, nil)
	if err != nil {
		return // Если не смогли создать запрос, просто выходим
	}

	resp, err := http.DefaultClient.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		slog.Warn("Target is OFFLINE", "url", task.URL, "duration_ms", duration.Milliseconds(), "error", err)
		return
	}
	defer resp.Body.Close()

	slog.Info("Target response", "url", task.URL, "status", resp.StatusCode, "duration_ms", duration.Milliseconds())
}
