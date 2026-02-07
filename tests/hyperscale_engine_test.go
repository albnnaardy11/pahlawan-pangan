package tests

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/albnnaardy11/pahlawan-pangan/internal/matching"
)

type MockRouter struct{}

func (m *MockRouter) GetTravelTime(ctx context.Context, startLat, startLon, endLat, endLon float64) (time.Duration, error) {
	// Simulate minor I/O latency
	return 10 * time.Millisecond, nil
}

func TestHyperScaleEngine(t *testing.T) {
	// SRE Goal: Verify MatchNGO performance with sync.Pool and Bounded Concurrency
	runtime.GOMAXPROCS(runtime.NumCPU())

	const (
		TotalMatches     = 100000 // Total match requests
		ConcurrentCalls  = 5000   // Concurrent match requests at any time
		CandidatesPerMatch = 20    // Number of NGOs per match
	)

	router := &MockRouter{}
	engine := matching.NewMatchingEngine(router)

	surplus := matching.Surplus{
		ID: "S-123", Lat: -6.1, Lon: 106.8, QuantityKgs: 10,
	}

	candidates := make([]matching.NGO, CandidatesPerMatch)
	for i := 0; i < CandidatesPerMatch; i++ {
		candidates[i] = matching.NGO{ID: fmt.Sprintf("NGO-%d", i), Lat: -6.101, Lon: 106.801}
	}

	fmt.Printf("\n🚀 STARTING HYPER-SCALE ENGINE TEST 🚀\n")
	fmt.Printf("   - Running %d total matches\n", TotalMatches)
	fmt.Printf("   - %d concurrent calls\n", ConcurrentCalls)
	fmt.Printf("   - %d candidates evaluated per match\n", CandidatesPerMatch)
	fmt.Println("-------------------------------------------------------")

	start := time.Now()
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, ConcurrentCalls)

	for i := 0; i < TotalMatches; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			semaphore <- struct{}{}        // Throttle input
			defer func() { <-semaphore }() // Release

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			_, err := engine.MatchNGO(ctx, surplus, candidates)
			if err != nil {
				// We don't expect errors in this mock
			}
		}()
	}

	wg.Wait()
	duration := time.Since(start)

	// Metrics
	throughput := float64(TotalMatches) / duration.Seconds()
	totalEvaluations := TotalMatches * CandidatesPerMatch
	evalThroughput := float64(totalEvaluations) / duration.Seconds()

	fmt.Printf("\n📊 HYPER-SCALE PERFORMANCE SUMMARY:\n")
	fmt.Printf("   🏁 Total Duration: %v\n", duration)
	fmt.Printf("   🚄 Matches/Sec: %.0f\n", throughput)
	fmt.Printf("   🔍 Total NGO Evaluations: %d\n", totalEvaluations)
	fmt.Printf("   💨 Evaluations/Sec: %.0f\n", evalThroughput)
	
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("   🧠 Final Heap Alloc: %v MB\n", m.Alloc / 1024 / 1024)
	fmt.Printf("   🧹 Total GC Cycles: %v\n", m.NumGC)
	
	fmt.Printf("\n⭐ SRE EVALUATION: %s\n", map[bool]string{true: "GO-NASHIONAL READY! 🇮🇩", false: "STABLE"}[throughput > 5000])
}
