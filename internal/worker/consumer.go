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
	ch          *amqp091.Channel
	logRepo     domain.LogRepository
	statusCache domain.StatusCache
}

func NewConsumer(ch *amqp091.Channel, logRepo domain.LogRepository, statusCache domain.StatusCache) *Consumer {
	return &Consumer{
		ch:          ch,
		logRepo:     logRepo,
		statusCache: statusCache,
	}
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
		slog.Error("Failed to create request", "error", err)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	duration := time.Since(startTime)

	var statusCode int
	var isUp bool

	if err != nil {
		slog.Warn("Target is OFFLINE", "url", task.URL, "duration_ms", duration.Milliseconds(), "error", err)
		statusCode = 0
		isUp = false
	} else {
		defer resp.Body.Close()
		slog.Info("Target response", "url", task.URL, "status", resp.StatusCode, "duration_ms", duration.Milliseconds())
		statusCode = resp.StatusCode
		isUp = (resp.StatusCode == http.StatusOK)
	}

	checkLog := &domain.CheckLog{
		TargetID:     task.ID,
		Status:       statusCode,
		ResponseTime: duration,
		Flag:         isUp,
		CreatedAt:    time.Now(),
	}

	if err := c.logRepo.SaveLog(ctx, checkLog); err != nil {
		slog.Error("Failed to save check log to Postgres", "error", err)
	}

	if err := c.statusCache.SetStatus(ctx, task.ID, isUp); err != nil {
		slog.Error("Failed to save status to Redis", "error", err)
	}
}
