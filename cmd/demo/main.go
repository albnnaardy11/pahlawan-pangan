package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"sync"
	"time"

	"github.com/albnnaardy11/pahlawan-pangan/internal/fintech"
	"github.com/albnnaardy11/pahlawan-pangan/internal/geo"
	"github.com/albnnaardy11/pahlawan-pangan/pkg/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

// In-Memory Storage
var (
	surplusDB = make(map[string]SurplusItem)
	mu        sync.RWMutex
)

type SurplusItem struct {
	ID          string    `json:"id"`
	ProviderID  string    `json:"provider_id"`
	FoodType    string    `json:"food_type"`
	QuantityKgs float64   `json:"quantity_kgs"`
	Status      string    `json:"status"` // available, claimed
	ExpiryTime  time.Time `json:"expiry_time"`
	CreatedAt   time.Time `json:"created_at"`
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// API Routes
	r.Post("/api/v1/surplus", PostSurplus)
	r.Get("/api/v1/marketplace", ListSurplus)
	r.Post("/api/v1/surplus/{id}/claim", ClaimSurplus)

	// Unicorn Phase 3 (Demo)
	r.Post("/api/v1/surplus/analyze-image", AnalyzeFoodImage)
	r.Get("/api/v1/impact/verify/{id}", VerifyBlockchainImpact)
	r.Post("/api/v1/marketplace/auction/bid", PlaceAuctionBid)

	// Viral Engine (New Phase 4)
	r.Get("/api/v1/impact/leaderboard", GetNationalLeaderboard)
	r.Get("/api/v1/impact/share/{id}", GenerateShareCard)

	// Phase 5: Reverse Logistics & Disputes
	r.Post("/api/v1/dispute", RaiseDispute)

	fmt.Println("🚀 Pahlawan Pangan - UNICORN VIRAL ENGINE READY")
	fmt.Println("-----------------------------------------------------")
	fmt.Println("✅ In-Memory Database Initialized")
	fmt.Println("✅ Geo-Spatial Engine (Mock) Ready")
	fmt.Println("✅ Serving on http://localhost:8080")
	fmt.Println("-----------------------------------------------------")
	fmt.Println("👉 Try: curl -X GET http://localhost:8080/api/v1/marketplace")

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}

func PostSurplus(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProviderID  string    `json:"provider_id"`
		FoodType    string    `json:"food_type"`
		QuantityKgs float64   `json:"quantity_kgs"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	item := SurplusItem{
		ID:          uuid.New().String(),
		ProviderID:  req.ProviderID,
		FoodType:    req.FoodType,
		QuantityKgs: req.QuantityKgs,
		Status:      "available",
		ExpiryTime:  time.Now().Add(24 * time.Hour),
		CreatedAt:   time.Now(),
	}

	mu.Lock()
	surplusDB[item.ID] = item
	mu.Unlock()

	s2Engine := geo.NewS2Engine()
	s2ID := s2Engine.GetCellID(-6.2, 106.8) // Jakarta Cell

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         item.ID,
		"status":     "posted",
		"geo_index":  fmt.Sprintf("S2-Cell-%d", s2ID), // Scalable Indexing
		"message":    "Surplus indexed using Google S2 Geometry",
	})
}

func ListSurplus(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	items := make([]SurplusItem, 0)
	for _, item := range surplusDB {
		if item.Status == "available" {
			items = append(items, item)
		}
	}
	
	// Seed some dummy data if empty
	if len(items) == 0 {
		items = append(items, SurplusItem{
			ID: uuid.New().String(), ProviderID: "Starbucks-Kemang", FoodType: "Pastry Box", QuantityKgs: 5.0, Status: "available", ExpiryTime: time.Now().Add(5*time.Hour),
		})
		items = append(items, SurplusItem{
			ID: uuid.New().String(), ProviderID: "Padang-Sederhana", FoodType: "Nasi Box", QuantityKgs: 15.0, Status: "available", ExpiryTime: time.Now().Add(2*time.Hour),
		})
	}

	json.NewEncoder(w).Encode(items)
}

func ClaimSurplus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	mu.Lock()
	defer mu.Unlock()

	item, exists := surplusDB[id]
	if !exists {
		// Mock claim for seeded data
		item = SurplusItem{ID: id, Status: "available"} 
	}

	if item.Status != "available" {
		http.Error(w, "Item not available", http.StatusConflict)
		return
	}

	item.Status = "claimed"
	surplusDB[id] = item

	// Simulate Matching Latency
	time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)

	// Fintech Layer: Lock Funds
	escrow := fintech.NewEscrowService()
	payment, _ := escrow.LockFunds(r.Context(), "NGO-User-001", 25000.0)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "claimed",
		"fulfillment": map[string]string{
			"method": "courier",
			"eta": "15 mins",
			"courier": "Pahlawan-Express Driver #402",
		},
		"escrow": map[string]interface{}{
			"payment_id": payment.ID,
			"status":     "FUNDS_LOCKED_IN_ESCROW",
			"amount":     "Rp25.000",
		},
		"note": "Uang aman di sistem. Akan dilepaskan ke resto setelah makanan diterima.",
	})
}

func AnalyzeFoodImage(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ai_model":          "Pahlawan-Vision-v3.0-Nutritionist-Pro",
		"overall_freshness": "98.2%",
		"nutrition": map[string]interface{}{
			"calories": "520 kcal",
			"protein":  "28g",
			"carbs":    "70g",
			"fat":      "15g",
		},
		"ahli_gizi_score": "Grade A (Highly Nutritious)",
		"advice":          "Tinggi protein dan rendah lemak jenuh. Sangat aman dan sehat untuk dikonsumsi.",
		"status":          "NUTRI_APPROVED",
	})
}

func VerifyBlockchainImpact(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"transaction_id": id,
		"status":         "verified_on_ledger",
		"chain":          "Pahlawan-Trust-Private",
		"proof_hash":     uuid.New().String(),
	})
}

func GenerateShareCard(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"share_url": fmt.Sprintf("https://pahlawanpangan.org/v/card-%s.png", id),
		"og_title":  "Saya baru saja menyelamatkan 2kg Makanan!",
		"og_description": "Bergabunglah menjadi pahlawan dan cegah 5kg Emisi CO2 hari ini.",
		"cta": "Download Apps Sekarang!",
	})
}

func GetNationalLeaderboard(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"top_cities": []map[string]interface{}{
			{"city": "Bandung", "total_saved_kgs": 45021, "rank": 1},
			{"city": "Jakarta", "total_saved_kgs": 43902, "rank": 2},
			{"city": "Surabaya", "total_saved_kgs": 38112, "rank": 3},
		},
		"top_merchants": []string{"Hotel Mulia", "Bakmi GM", "Starbucks Indo"},
		"national_impact": "1.2 Million Kgs of Food Saved in 2026",
	})
}

func PlaceAuctionBid(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"auction_status": "won",
		"message":        "Dutch Auction Win! Final Price locked.",
	})
}

func RaiseDispute(w http.ResponseWriter, r *http.Request) {
	traceID := uuid.New().String() // OpenTelemetry Propagation simulation
	w.Header().Set("X-Trace-ID", traceID)
	
	fmt.Printf("🔍 [OBSERVABILITY] Trace ID: %s - Processing Dispute...\n", traceID)
	
	// PII Masking Simulation
	email := "budi.pahlawan@gmail.com"
	maskedEmail := utils.MaskPII(email)
	fmt.Printf("🛡️  [PRIVACY] Logging masked user data: %s\n", maskedEmail)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"trace_id": traceID,
		"message":  "Dispute received. Automated Refund check started.",
		"policy":   "15-min auto-refund active",
	})
}

