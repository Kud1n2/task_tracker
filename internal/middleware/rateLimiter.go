package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"task_tracker/internal/config"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	redis  *redis.Client
	limit  int           // например 100
	window time.Duration // 1 минута
}

func NewRateLimiter(cfg config.CacheConfig) *RateLimiter {
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})
	return &RateLimiter{
		redis:  rdb,
		limit:  100,
		window: time.Minute,
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		userIDRaw := r.Context().Value(UserIDKey)
		if userIDRaw == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userID := userIDRaw.(int64)
		key := fmt.Sprintf("rate:%d", userID)

		ctx := r.Context()

		count, err := rl.redis.Incr(ctx, key).Result()
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		rl.redis.Do(ctx, "EXPIRE", key, int(rl.window.Seconds()), "NX")

		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.limit))

		remaining := rl.limit - int(count)
		if remaining < 0 {
			remaining = 0
		}

		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))

		if int(count) > rl.limit {
			w.Header().Set("Retry-After", strconv.Itoa(int(rl.window.Seconds())))
			http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
