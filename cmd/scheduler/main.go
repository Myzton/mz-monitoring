package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"mz-monitoring/internal/repository/postgres"
	"mz-monitoring/internal/repository/rabbitmq"
	"mz-monitoring/internal/worker"
	"mz-monitoring/pkg/env"
	pkgRabbit "mz-monitoring/pkg/rabbitmq"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	env.Load(".env")

	pgDSN := env.Get("DATABASE_URL", "postgres://postgres:mysecretpassword@localhost:5432/mz_monitoring")
	pool, err := pgxpool.New(ctx, pgDSN)
	if err != nil {
		log.Fatalf("Scheduler: unable to connect to PostgreSQL: %v", err)
	}
	defer pool.Close()
	slog.Info("Scheduler successfully connected to PostgreSQL")

	rabbitURL := env.Get("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	conn, ch, err := pkgRabbit.InitRabbitMQ(rabbitURL)
	if err != nil {
		log.Fatalf("Scheduler: unable to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	defer ch.Close()
	slog.Info("Scheduler successfully connected to RabbitMQ and declared queue")

	targetRepo := postgres.NewTargetRepository(pool)
	queuePub := rabbitmq.NewRabbitPublisher(ch)

	sched := worker.NewScheduler(targetRepo, queuePub)

	slog.Info("Scheduler is starting...")
	if err := sched.Start(ctx); err != nil {
		slog.Error("Scheduler stopped with error", "error", err)
	}
}
