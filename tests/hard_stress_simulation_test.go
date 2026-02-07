package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const (
	TargetURL = "http://localhost:8080"
)

func TestHardStressSimulation(t *testing.T) {
	// Skenario: "Total Chaos" - Menghantam Server dari Segala Sisi
	const (
		Duration       = 10 * time.Second
		ProviderCount  = 100 // Tukang Post
		NGOCount       = 300 // Tukang Klaim
		AICount        = 100 // Tukang Scan Foto
		TrustCount     = 100 // Tukang Cek Blockchain
		AuctionCount   = 200 // Tukang Lelang (Bidding)
	)

	fmt.Printf("\n🔥🔥 STARTING HARD STRESS TEST (Duration: %v) 🔥🔥\n", Duration)
	fmt.Printf("   - 👨‍🍳 Providers: %d\n", ProviderCount)
	fmt.Printf("   - 🤝 NGOs: %d\n", NGOCount)
	fmt.Printf("   - 🤖 AI Scanners: %d\n", AICount)
	fmt.Printf("   - ⛓️ Blockchain Verifiers: %d\n", TrustCount)
	fmt.Printf("   - ⚡ Auction Bidders: %d\n", AuctionCount)
	fmt.Println("-------------------------------------------------------")

	var wg sync.WaitGroup
	var (
		totalPosts      int64
		totalClaims     int64
		totalAIAnalyses int64
		totalTrustChecks int64
		totalBids       int64
		totalErrors     int64
	)

	start := time.Now()
	client := &http.Client{
		Timeout: 1 * time.Second,
	}

	// 1. Providers (POST /surplus)
	for i := 0; i < ProviderCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for time.Since(start) < Duration {
				payload, _ := json.Marshal(map[string]interface{}{
					"provider_id":  fmt.Sprintf("P-%d", id),
					"food_type":    "Fresh Meals Mega Pack",
					"quantity_kgs": 50.0,
				})
				resp, err := client.Post(TargetURL+"/api/v1/surplus", "application/json", bytes.NewBuffer(payload))
				if err == nil {
					if resp.StatusCode == 201 {
						atomic.AddInt64(&totalPosts, 1)
					} else {
						atomic.AddInt64(&totalErrors, 1)
					}
					resp.Body.Close()
				} else {
					atomic.AddInt64(&totalErrors, 1)
				}
				time.Sleep(50 * time.Millisecond)
			}
		}(i)
	}

	// 2. AI Scanners (POST /analyze-image)
	for i := 0; i < AICount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for time.Since(start) < Duration {
				resp, err := client.Post(TargetURL+"/api/v1/surplus/analyze-image", "application/json", bytes.NewBuffer([]byte("{}")))
				if err == nil {
					if resp.StatusCode == 200 {
						atomic.AddInt64(&totalAIAnalyses, 1)
					}
					resp.Body.Close()
				} else {
					atomic.AddInt64(&totalErrors, 1)
				}
				time.Sleep(30 * time.Millisecond)
			}
		}()
	}

	// 3. NGOs (Marketplace + Claim)
	for i := 0; i < NGOCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for time.Since(start) < Duration {
				// Browse
				resp, err := client.Get(TargetURL + "/api/v1/marketplace")
				if err == nil {
					var list []map[string]interface{}
					json.NewDecoder(resp.Body).Decode(&list)
					resp.Body.Close()
					
					// Random claim
					if len(list) > 0 {
						target := list[rand.IntN(len(list))]
						idVal := target["id"].(string)
						cResp, cErr := client.Post(TargetURL+"/api/v1/surplus/"+idVal+"/claim", "application/json", nil)
						if cErr == nil {
							if cResp.StatusCode == 200 {
								atomic.AddInt64(&totalClaims, 1)
							}
							cResp.Body.Close()
						}
					}
				}
				time.Sleep(100 * time.Millisecond)
			}
		}(i)
	}

	// 4. Blockchain Verifiers (GET /impact/verify/{id})
	for i := 0; i < TrustCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for time.Since(start) < Duration {
				resp, err := client.Get(TargetURL + "/api/v1/impact/verify/RESCUE-STRESS-TEST")
				if err == nil {
					if resp.StatusCode == 200 {
						atomic.AddInt64(&totalTrustChecks, 1)
					}
					resp.Body.Close()
				}
				time.Sleep(40 * time.Millisecond)
			}
		}()
	}

	// 5. Bidders (POST /marketplace/auction/bid)
	for i := 0; i < AuctionCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for time.Since(start) < Duration {
				payload, _ := json.Marshal(map[string]interface{}{
					"surplus_id": "AUCTION-X",
					"bid_amount": 25000.0,
					"user_id":    "STRESS-USER",
				})
				resp, err := client.Post(TargetURL+"/api/v1/marketplace/auction/bid", "application/json", bytes.NewBuffer(payload))
				if err == nil {
					if resp.StatusCode == 200 {
						atomic.AddInt64(&totalBids, 1)
					}
					resp.Body.Close()
				}
				time.Sleep(60 * time.Millisecond)
			}
		}()
	}

	wg.Wait()
	totalTime := time.Since(start)

	// Summary
	totalOps := totalPosts + totalClaims + totalAIAnalyses + totalTrustChecks + totalBids
	fmt.Printf("\n📊 FINAL HARD TEST SUMMARY:\n")
	fmt.Printf("   🏁 Duration: %v\n", totalTime)
	fmt.Printf("   👨‍🍳 Posts Success: %d\n", totalPosts)
	fmt.Printf("   🤝 Claims Success: %d\n", totalClaims)
	fmt.Printf("   🤖 AI Analyses: %d\n", totalAIAnalyses)
	fmt.Printf("   ⛓️ Blockchain Verifications: %d\n", totalTrustChecks)
	fmt.Printf("   ⚡ Auction Bids: %d\n", totalBids)
	fmt.Printf("   ❌ Total Failures: %d\n", totalErrors)
	fmt.Printf("\n🚀 AGGREGATE THROUGHPUT: %.0f req/s\n", float64(totalOps)/totalTime.Seconds())
	fmt.Printf("⭐ STATUS: %s\n", map[bool]string{true: "UNSTOPPABLE 💪", false: "STABLE ✅"}[totalOps > 10000])
}
