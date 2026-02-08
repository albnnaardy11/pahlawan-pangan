package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Simplified structures for demo
type DemoSurplus struct {
	ID          string  `json:"id"`
	ProviderID  string  `json:"provider_id"`
	FoodType    string  `json:"food_type"`
	QuantityKgs float64 `json:"quantity_kgs"`
	Status      string  `json:"status"`
}

var (
	baseURL = "http://localhost:8080"
)

func TestFullWorkflowSimulation(t *testing.T) {
	// Wait for server to be ready (assuming running in background via cmd/demo/main.go)
	// In a real test, we would start the server here in a goroutine

	const (
		NumProviders = 50  // 50 Restaurants posting food
		NumNGOs      = 100 // 100 NGOs trying to claim
		NumAdmins    = 5   // 5 Admins monitoring
		testDuration = 5 * time.Second
	)

	fmt.Printf("\n🚀 STARTING CHAOS SIMULATION: ALL ROLES ACTIVE\n")
	fmt.Printf("   - 👨‍🍳 %d Providers posting surplus\n", NumProviders)
	fmt.Printf("   - 🤝 %d NGOs claiming concurrently\n", NumNGOs)
	fmt.Printf("   - 👮 %d Admins monitoring dashboard\n", NumAdmins)
	fmt.Println("---------------------------------------------------")

	var wg sync.WaitGroup
	var (
		totalPosted  int64
		totalClaimed int64
		totalErrors  int64
		totalReads   int64
	)

	start := time.Now()

	// 1. Providers: Post Surplus continuously
	for i := 0; i < NumProviders; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			providerID := fmt.Sprintf("P-%d", id)

			for time.Since(start) < testDuration {
				surplus := DemoSurplus{
					ProviderID:  providerID,
					FoodType:    "Nasi Box Premium",
					QuantityKgs: 10.0,
				}

				payload, _ := json.Marshal(surplus)
				resp, err := http.Post(baseURL+"/api/v1/surplus", "application/json", bytes.NewBuffer(payload))
				if err != nil {
					atomic.AddInt64(&totalErrors, 1)
					continue
				}
				if resp.StatusCode == http.StatusCreated {
					atomic.AddInt64(&totalPosted, 1)
				} else {
					atomic.AddInt64(&totalErrors, 1)
				}
				_ = resp.Body.Close()
				time.Sleep(100 * time.Millisecond) // Human delay
			}
		}(i)
	}

	// 2. NGOs: Scan and Claim continuously
	for i := 0; i < NumNGOs; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ngoID := fmt.Sprintf("NGO-%d", id)

			for time.Since(start) < testDuration {
				// Step A: Browse Marketplace
				resp, err := http.Get(baseURL + "/api/v1/marketplace")
				if err != nil {
					atomic.AddInt64(&totalErrors, 1)
					continue
				}

				var items []DemoSurplus
				_ = json.NewDecoder(resp.Body).Decode(&items)
				_ = resp.Body.Close()
				atomic.AddInt64(&totalReads, 1)

				// Step B: Claim first available item
				if len(items) > 0 {
					target := items[0] // Race condition bait!
					claimURL := fmt.Sprintf("%s/api/v1/surplus/%s/claim", baseURL, target.ID)

					claimPayload, _ := json.Marshal(map[string]string{
						"ngo_id": ngoID,
						"method": "courier",
					})

					req, _ := http.NewRequest("POST", claimURL, bytes.NewBuffer(claimPayload))
					cResp, err := http.DefaultClient.Do(req)

					if err == nil {
						if cResp.StatusCode == 200 {
							atomic.AddInt64(&totalClaimed, 1)
						}
						_ = cResp.Body.Close()
					}
				}
				time.Sleep(50 * time.Millisecond)
			}
		}(i)
	}

	// 3. Admins: Passive Monitoring (High Read Load)
	var wgAdmins sync.WaitGroup
	wgAdmins.Add(NumAdmins)
	for i := 0; i < NumAdmins; i++ {
		go func() {
			defer wgAdmins.Done()
			for {
				if time.Since(start) >= testDuration {
					return
				}
				resp, err := http.Get(baseURL + "/api/v1/marketplace")
				if err == nil {
					_, _ = io.Copy(io.Discard, resp.Body)
					_ = resp.Body.Close()
					atomic.AddInt64(&totalReads, 1)
				}
				time.Sleep(200 * time.Millisecond)
			}
		}()
	}
	wgAdmins.Wait()

	wg.Wait()

	duration := time.Since(start)

	fmt.Printf("\n📊 SIMULATION RESULTS (Duration: %v)\n", duration)
	fmt.Printf("   ✅ Total Surplus Posted: %d\n", totalPosted)
	fmt.Printf("   ✅ Total Successfully Claimed: %d\n", totalClaimed)
	fmt.Printf("   👀 Total Marketplace Views: %d\n", totalReads)
	fmt.Printf("   ❌ Errors (Network/Timeout): %d\n", totalErrors)
	fmt.Printf("   ⚡ Approx Throughput: %.0f req/s\n", float64(totalPosted+totalClaimed+totalReads)/duration.Seconds())

	if totalErrors == 0 {
		fmt.Println("\n🏆 SYSTEM STABILITY: PERFECT (0 Errors)")
	} else {
		fmt.Printf("\n⚠️ SYSTEM STABILITY: %d Errors detected (Review logs)\n", totalErrors)
	}
}

// Wrapper for simple execution
func TestRun(t *testing.T) {
	// Only run if server is up
	resp, err := http.Get(baseURL + "/api/v1/marketplace")
	if err != nil {
		t.Skip("Skipping integration test: server not running")
	}
	defer resp.Body.Close()
	TestFullWorkflowSimulation(t)
}
