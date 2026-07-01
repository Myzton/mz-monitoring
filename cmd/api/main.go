package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"time"

	delivery "mz-monitoring/internal/delivery/http"
	"mz-monitoring/internal/delivery/http/middleware"
	"mz-monitoring/internal/repository/postgres"
	redisRepo "mz-monitoring/internal/repository/redis"
	"mz-monitoring/internal/usecase"
	"mz-monitoring/pkg/env"

	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()

	env.Load(".env")

	jwtSecret := env.Get("JWT_SECRET", "")
	if jwtSecret == "" {
		jwtSecret = "super-secret-key-for-local-dev-12345"
		slog.Warn("JWT_SECRET env not found, using default secret key")
	}

	pgDSN := env.Get("DATABASE_URL", "postgres://postgres:mysecretpassword@localhost:5432/mz_monitoring")
	pool, err := pgxpool.New(ctx, pgDSN)
	if err != nil {
		log.Fatalf("Unable to connect to PostgreSQL: %v", err)
	}
	defer pool.Close()
	slog.Info("Successfully connected to PostgreSQL")

	redisAddr := env.Get("REDIS_ADDR", "localhost:6379")
	redisClient := goredis.NewClient(&goredis.Options{
		Addr: redisAddr,
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Unable to connect to Redis: %v", err)
	}
	defer redisClient.Close()
	slog.Info("Successfully connected to Redis")

	userRepo := postgres.NewUserRepository(pool)
	targetRepo := postgres.NewTargetRepository(pool)
	statusCache := redisRepo.NewStatusRepository(redisClient)

	userUsecase := usecase.NewUserUsecase(userRepo, []byte(jwtSecret))
	targetUsecase := usecase.NewTargetUsecase(targetRepo, statusCache)

	userHandler := delivery.NewUserHandler(userUsecase)
	targetHandler := delivery.NewTargetHandler(targetUsecase)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /auth/register", userHandler.CreateUserHandler)
	mux.HandleFunc("POST /auth/login", userHandler.LoginHandler)

	mux.Handle("POST /targets", middleware.AuthMiddleware(http.HandlerFunc(targetHandler.CreateTargetHandler), jwtSecret))
	mux.Handle("GET /targets", middleware.AuthMiddleware(http.HandlerFunc(targetHandler.GetListTargetHandler), jwtSecret))
	mux.Handle("DELETE /targets/{id}", middleware.AuthMiddleware(http.HandlerFunc(targetHandler.DeleteTargetHandler), jwtSecret))

	handlerWithRateLimit := middleware.RateLimitMiddleware(redisClient)(mux)

	serverAddr := ":8080"
	slog.Info("API Server is running", "address", serverAddr)

	server := &http.Server{
		Addr:         serverAddr,
		Handler:      handlerWithRateLimit,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
