package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	delivery "mz-monitoring/internal/delivery/http"
	"mz-monitoring/internal/delivery/http/middleware"
	"mz-monitoring/internal/repository/postgres"
	redisRepo "mz-monitoring/internal/repository/redis"
	"mz-monitoring/internal/usecase"

	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "super-secret-key-for-local-dev-12345"
		slog.Warn("JWT_SECRET env not found, using default secret key")
	}

	pgDSN := "postgres://postgres:mysecretpassword@localhost:5432/mz_monitoring"
	pool, err := pgxpool.New(ctx, pgDSN)
	if err != nil {
		log.Fatalf("Unable to connect to PostgreSQL: %v", err)
	}
	defer pool.Close()
	slog.Info("Successfully connected to PostgreSQL")

	redisClient := goredis.NewClient(&goredis.Options{
		Addr: "localhost:6379",
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
