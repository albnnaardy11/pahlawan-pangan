package tests

import (
	"context"
	"fmt"
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/albnnaardy11/pahlawan-pangan/internal/matching"
)

// ScaleSimulationMockRouter simulates a routing service
type ScaleSimulationMockRouter struct{}

func (m *ScaleSimulationMockRouter) GetTravelTime(ctx context.Context, startLat, startLon, endLat, endLon float64) (time.Duration, error) {
	// Simulate network latency (50-200ms)
	//nolint:gosec // G404: Using math/rand for test simulation, not cryptographic purposes
	latency := time.Duration(50+rand.IntN(150)) * time.Millisecond
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case <-time.After(latency):
		return latency, nil // Return latency as duration
	}
}

func TestScaleSimulation(t *testing.T) {
	// Setup
	router := &ScaleSimulationMockRouter{}
	engine := matching.NewMatchingEngine(router)

	// Simulation Parameters
	const (
		ConcurrentUsers = 1000 // Number of concurrent users trying to post surplus
		CandidatesPerUser = 20   // Number of NGOs to match against per user
	)

	fmt.Printf("🚀 Starting Load Simulation\n")
	fmt.Printf("   - Concurrent Users: %d\n", ConcurrentUsers)
	fmt.Printf("   - Candidates per User: %d\n", CandidatesPerUser)
	fmt.Printf("   - Total Goroutines (Potential): %d\n", ConcurrentUsers*CandidatesPerUser)

	var (
		successCount int64
		errorCount   int64
		totalLatency int64
	)

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(ConcurrentUsers)

	for i := 0; i < ConcurrentUsers; i++ {
		go func(id int) {
			defer wg.Done()
			
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // 2s global timeout
			defer cancel()

			surplus := matching.Surplus{
				ID:          fmt.Sprintf("surplus-%d", id),
				ProviderID:  fmt.Sprintf("provider-%d", id),
				//nolint:gosec // G404: Using math/rand for test mock coordinates, not cryptographic purposes
				Lat:         -6.200000 + (rand.Float64() * 0.01),
				//nolint:gosec // G404: Using math/rand for test mock coordinates, not cryptographic purposes
				Lon:         106.816666 + (rand.Float64() * 0.01),
				ExpiryTime:  time.Now().Add(2 * time.Hour),
				QuantityKgs: 10.0,
			}

			candidates := make([]matching.NGO, CandidatesPerUser)
			for j := 0; j < CandidatesPerUser; j++ {
				candidates[j] = matching.NGO{
					ID:  fmt.Sprintf("ngo-%d-%d", id, j),
					//nolint:gosec // G404: Using math/rand for test mock coordinates, not cryptographic purposes
					Lat: -6.200000 + (rand.Float64() * 0.01),
					//nolint:gosec // G404: Using math/rand for test mock coordinates, not cryptographic purposes
					Lon: 106.816666 + (rand.Float64() * 0.01),
				}
			}

			reqStart := time.Now()
			_, err := engine.MatchNGO(ctx, surplus, candidates)
			duration := time.Since(reqStart)

			if err != nil {
				atomic.AddInt64(&errorCount, 1)
			} else {
				atomic.AddInt64(&successCount, 1)
				atomic.AddInt64(&totalLatency, int64(duration))
			}
		}(i)
	}

	wg.Wait()
	totalTime := time.Since(start)

	// Results
	avgLatency := time.Duration(0)
	if successCount > 0 {
		avgLatency = time.Duration(totalLatency / successCount)
	}

	fmt.Printf("\n📊 Simulation Results:\n")
	fmt.Printf("   - Total Time: %v\n", totalTime)
	fmt.Printf("   - Successful Matches: %d\n", successCount)
	fmt.Printf("   - Failed Matches: %d\n", errorCount)
	fmt.Printf("   - Average Latency (Success): %v\n", avgLatency)
	fmt.Printf("   - Throughput: %.2f req/s\n", float64(ConcurrentUsers)/totalTime.Seconds())

	// extrapolate for 100M users
	p99Latency := avgLatency * 2 // Estimate
	fmt.Printf("\n🔮 Projection for 100M Users (Concurrent):\n")
	fmt.Printf("   - P99 Latency (Est): %v\n", p99Latency)
	fmt.Printf("   - Required Throughput: ~100,000,000 req/s (Impossible on single node)\n")
	fmt.Printf("   - Estimated Cluster Size (1000 req/s per node): 100,000 nodes\n")
	
	if errorCount > 0 {
		t.Logf("⚠️ Warning: %d requests failed due to timeouts or errors.", errorCount)
	}
}
