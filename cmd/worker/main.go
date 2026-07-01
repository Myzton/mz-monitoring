package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"mz-monitoring/internal/repository/postgres"
	redisRepo "mz-monitoring/internal/repository/redis"
	"mz-monitoring/internal/worker"
	"mz-monitoring/pkg/env"
	pkgRabbit "mz-monitoring/pkg/rabbitmq"

	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	env.Load(".env")

	pgDSN := env.Get("DATABASE_URL", "postgres://postgres:mysecretpassword@localhost:5432/mz_monitoring")
	pool, err := pgxpool.New(ctx, pgDSN)
	if err != nil {
		log.Fatalf("Worker: unable to connect to PostgreSQL: %v", err)
	}
	defer pool.Close()
	slog.Info("Worker successfully connected to PostgreSQL")

	redisAddr := env.Get("REDIS_ADDR", "localhost:6379")
	redisClient := goredis.NewClient(&goredis.Options{
		Addr: redisAddr,
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Worker: unable to connect to Redis: %v", err)
	}
	defer redisClient.Close()
	slog.Info("Worker successfully connected to Redis")

	rabbitURL := env.Get("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	conn, ch, err := pkgRabbit.InitRabbitMQ(rabbitURL)
	if err != nil {
		log.Fatalf("Worker: unable to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	defer ch.Close()
	slog.Info("Worker successfully connected to RabbitMQ")

	logRepo := postgres.NewLogRepository(pool)
	statusCache := redisRepo.NewStatusRepository(redisClient)

	consumer := worker.NewConsumer(ch, logRepo, statusCache)

	slog.Info("Worker Consumer is listening for tasks...")
	if err := consumer.Start(ctx); err != nil {
		slog.Error("Worker stopped with error", "error", err)
	}
}
