package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// AdaptiveLoadShedder monitors system latency and sheds load if threshold is breached.
// Standard in Google (SRE Book) and Netflix.
type AdaptiveLoadShedder struct {
	mu           sync.RWMutex
	latencySum   time.Duration
	requestCount int
	threshold    time.Duration
	isShedding   bool
}

func NewAdaptiveLoadShedder(threshold time.Duration) *AdaptiveLoadShedder {
	return &AdaptiveLoadShedder{
		threshold: threshold,
	}
}

func (ls *AdaptiveLoadShedder) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ls.mu.RLock()
		shedding := ls.isShedding
		ls.mu.RUnlock()

		if shedding {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("⚠️  [SRE-LOAD-SHEDDING] System is under heavy load. Request rejected to maintain stability."))
			return
		}

		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)

		ls.recordLatency(duration)
	})
}

func (ls *AdaptiveLoadShedder) recordLatency(d time.Duration) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	ls.latencySum += d
	ls.requestCount++

	// Analyze every 100 requests (or a time window)
	if ls.requestCount >= 100 {
		avg := ls.latencySum / time.Duration(ls.requestCount)
		if avg > ls.threshold {
			if !ls.isShedding {
				fmt.Printf("🚨 [SRE-ALERT] Latency (%.2fms) exceeded threshold (%v). SHEDDING LOAD START.\n", float64(avg.Milliseconds()), ls.threshold)
			}
			ls.isShedding = true
		} else {
			if ls.isShedding {
				fmt.Printf("✅ [SRE-INFO] Latency (%.2fms) recovered. SHEDDING LOAD STOP.\n", float64(avg.Milliseconds()))
			}
			ls.isShedding = false
		}
		// Reset counters
		ls.latencySum = 0
		ls.requestCount = 0
	}
}
