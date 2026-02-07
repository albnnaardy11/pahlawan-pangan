package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

// DistributedRateLimiter uses Redis to limit requests across multiple server nodes.
// Essential for national scale (Tokopedia/Gojek level).
type DistributedRateLimiter struct {
	client *redis.Client
	limit  int
	window time.Duration
}

func NewDistributedRateLimiter(client *redis.Client, limit int, window time.Duration) *DistributedRateLimiter {
	return &DistributedRateLimiter{client: client, limit: limit, window: window}
}

func (l *DistributedRateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr // In prod: X-Forwarded-For
		key := "rate_limit:" + ip
		
		ctx := context.Background()
		count, err := l.client.Incr(ctx, key).Result()
		if err != nil {
			next.ServeHTTP(w, r) // Fail-open strategy to maintain availability
			return
		}

		if count == 1 {
			l.client.Expire(ctx, key, l.window)
		}

		if int(count) > l.limit {
			w.Header().Set("X-RateLimit-Limit", string(rune(l.limit)))
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("🚫 [SRE-DISTRIBUTED] Rate limit exceeded. Try again in a few seconds."))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// IdempotencyMiddleware ensures a request is only processed once.
func IdempotencyMiddleware(client *redis.Client) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-Idempotency-Key")
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.Background()
			set, err := client.SetNX(ctx, "idempotency:"+key, "processing", 1*time.Hour).Result()
			if err != nil || !set {
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte("⚠️ Request is already being processed or has been completed."))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
