package middleware

import (
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// ChaosMiddleware simulates network faults and errors to test system resilience.
// Inspired by Chaos Mesh and Netflix Chaos Monkey.
func ChaosMiddleware(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check for Chaos Headers (In prod, restrict this to internal IPs or Test Users)
			chaos := r.Header.Get("X-Chaos-Simulate")
			if chaos != "" {
				// 1. Latency Injection
				if latencyStr := r.Header.Get("X-Chaos-Latency-Ms"); latencyStr != "" {
					latency, _ := strconv.Atoi(latencyStr)
					if latency > 0 {
						logger.Warn("🔥 Injecting Chaos Latency", zap.Int("ms", latency))
						time.Sleep(time.Duration(latency) * time.Millisecond)
					}
				}

				// 2. Random Failure Injection (500 Error)
				if failureRateStr := r.Header.Get("X-Chaos-Failure-Rate"); failureRateStr != "" {
					rate, _ := strconv.ParseFloat(failureRateStr, 64) // e.g., 0.1 for 10%
					if rand.Float64() < rate {
						logger.Error("🔥 Injecting Chaos Failure (500)")
						http.Error(w, "Chaos Monkey Striked!", http.StatusInternalServerError)
						return
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
