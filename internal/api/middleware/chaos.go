package middleware

import (
	"crypto/rand"
	"encoding/binary"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// secureRandomFloat64 returns a float64 in [0, 1) using crypto/rand to satisfy SAST (G404).
func secureRandomFloat64() float64 {
	var b [8]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return 0 // Fallback to safe value
	}
	// Use 53 bits for precision
	val := binary.LittleEndian.Uint64(b[:]) & ((1 << 53) - 1)
	return float64(val) / (1 << 53)
}

// ChaosMiddleware simulates network faults and errors to test system resilience.
// Inspired by Chaos Mesh and Netflix Chaos Monkey.
func ChaosMiddleware(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check for Chaos Headers (In prod, restrict this to internal IPs or Test Users)
			// Standard practice: check for internal headers or specific auth roles
			chaos := r.Header.Get("X-Chaos-Simulate")
			if chaos != "" {
				// 1. Latency Injection (Capped at 5s to prevent intentional DoS)
				if latencyStr := r.Header.Get("X-Chaos-Latency-Ms"); latencyStr != "" {
					latency, _ := strconv.Atoi(latencyStr)
					if latency > 0 {
						if latency > 5000 {
							latency = 5000 // Safety cap
						}
						logger.Warn("🔥 Injecting Chaos Latency", zap.Int("ms", latency))
						time.Sleep(time.Duration(latency) * time.Millisecond)
					}
				}

				// 2. Random Failure Injection (500 Error)
				if failureRateStr := r.Header.Get("X-Chaos-Failure-Rate"); failureRateStr != "" {
					rate, _ := strconv.ParseFloat(failureRateStr, 64) // e.g., 0.1 for 10%
					if secureRandomFloat64() < rate {
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
