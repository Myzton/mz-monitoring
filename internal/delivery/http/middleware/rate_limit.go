package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func RateLimitMiddleware(rdb *redis.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			ip := r.RemoteAddr
			if idx := strings.LastIndex(ip, ":"); idx != -1 {
				ip = ip[:idx]
			}

			key := fmt.Sprintf("rate_limit:%s", ip)

			count, err := rdb.Incr(ctx, key).Result()
			if err != nil {
				slog.Error("Redis rate limit error", "error", err)
				next.ServeHTTP(w, r)
				return
			}

			if count == 1 {
				rdb.Expire(ctx, key, time.Minute)
			}

			if count > 60 {
				slog.Warn("Rate limit exceeded", "ip", ip, "requests_count", count)
				http.Error(w, "429 Too Many Requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
