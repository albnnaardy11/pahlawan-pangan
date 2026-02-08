package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// DistributedRateLimiter uses Redis to limit requests across multiple server nodes.
// Essential for national scale (Tokopedia/Gojek level).
type DistributedRateLimiter struct {
	redis  *redis.Client
	limit  int
	window time.Duration
}

func NewDistributedRateLimiter(redisClient *redis.Client, limit int, window time.Duration) *DistributedRateLimiter {
	return &DistributedRateLimiter{redis: redisClient, limit: limit, window: window}
}

// Lua script for atomic increment and expire
var rateLimitScript = redis.NewScript(`
	local count = redis.call("INCR", KEYS[1])
	if count == 1 then
		redis.call("EXPIRE", KEYS[1], ARGV[1])
	end
	return count
`)

func (l *DistributedRateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Resolve Real IP (Standard SRE practice for national-scale apps behind LB/CDN)
		ip := r.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = r.Header.Get("X-Real-IP")
		}
		if ip == "" {
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err == nil {
				ip = host
			} else {
				ip = r.RemoteAddr
			}
		} else {
			// X-Forwarded-For can be a comma-separated list
			if strings.Contains(ip, ",") {
				ip = strings.TrimSpace(strings.Split(ip, ",")[0])
			}
		}

		key := "rate_limit:" + ip
		ctx := r.Context()

		// 2. Atomic Increment & Expire using Lua (Ensures window is set even if process crashes)
		result, err := rateLimitScript.Run(ctx, l.redis, []string{key}, int(l.window.Seconds())).Int64()
		if err != nil {
			// Fail-open strategy: better to allow extra requests than block legit users when Redis is flaky
			next.ServeHTTP(w, r)
			return
		}

		if result > int64(l.limit) {
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", l.limit))
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte("🚫 [SRE-DISTRIBUTED] Rate limit exceeded. Try again in a few seconds."))
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

			ctx := r.Context()
			set, err := client.SetNX(ctx, "idempotency:"+key, "processing", 1*time.Hour).Result()
			if err != nil || !set {
				w.WriteHeader(http.StatusConflict)
				_, _ = w.Write([]byte("⚠️ Request is already being processed or has been completed."))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
