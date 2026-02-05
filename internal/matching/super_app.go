package matching

import (
	"fmt"
	"math"
	"sort"
)

// RecommendationEngine implements weighted scoring for Super-App ranking
type RecommendationEngine struct{}

type SurplusCandidate struct {
	ID                string
	DistanceKm        float64
	Rating            float64
	DiscountPercent   float64
	TemperatureSafe   bool
	FinalScore        float64
}

// RankSurplus uses a multi-factor weighting algorithm similar to Gojek/Grab
func (e *RecommendationEngine) RankSurplus(candidates []SurplusCandidate) []SurplusCandidate {
	for i := range candidates {
		// Weighting: 40% Rating, 30% Proximity, 30% Price Value
		proximityScore := 1.0 / (candidates[i].DistanceKm + 1.0)
		
		score := (candidates[i].Rating * 0.4) +
			(proximityScore * 10.0 * 0.3) + // Normalize proximity
			((candidates[i].DiscountPercent / 100.0) * 5.0 * 0.3)

		// Hard filter: If not temperature safe, penalize heavily
		if !candidates[i].TemperatureSafe {
			score -= 10.0
		}

		candidates[i].FinalScore = score
	}

	// Sort by score descending
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].FinalScore > candidates[j].FinalScore
	})

	return candidates
}

// VoucherService handles the promo activation logic
type VoucherService struct{}

func (s *VoucherService) ValidateVoucher(code string, orderValue float64) (float64, bool) {
	// Standard Tokopedia logic: min-order check
	if code == "PAHLAWANBARU" && orderValue > 50000 {
		return 15000, true // Fixed 15k discount
	}
	if code == "ZEROWASTE" {
		return orderValue * 0.2, true // 20% off
	}
	return 0, false
}

// FlashSaleMonitor identifies deep-discount items for push notifications
func (e *RecommendationEngine) DetectFlashSale(candidates []SurplusCandidate) []string {
	flashIDs := []string{}
	for _, c := range candidates {
		if c.DiscountPercent >= 80.0 {
			flashIDs = append(flashIDs, c.ID)
		}
	}
	return flashIDs
}

// --- Hybrid Fulfillment Orchestrator ---

type FulfillmentOption string

const (
	FulfillmentCourier    FulfillmentOption = "courier"
	FulfillmentSelfPickup FulfillmentOption = "self_pickup"
)

type FulfillmentStatus struct {
	Method          FulfillmentOption `json:"method"`
	TrackingID      string            `json:"tracking_id,omitempty"`
	VerificationCode string           `json:"verification_code,omitempty"`
	DistanceToStore float64           `json:"distance_to_store_meters"`
}

// OrchestrateFulfillment handles the logic for choosing and validating fulfillment methods
func OrchestrateFulfillment(method FulfillmentOption, userLat, userLon, storeLat, storeLon float64) (FulfillmentStatus, error) {
	// Calculate actual distance to store for self-pickup validation
	// Haversine formula already available in engine.go
	dist := calculateHaversine(userLat, userLon, storeLat, storeLon)

	if method == FulfillmentSelfPickup {
		// Complex Logic: Only allow self-pickup if user is within 5km radius
		if dist > 5.0 {
			return FulfillmentStatus{}, fmt.Errorf("distance too far for self-pickup: %.2f km", dist)
		}
		
		return FulfillmentStatus{
			Method:           FulfillmentSelfPickup,
			VerificationCode: "PAH-PICK-77", // Would be generated in real DB
			DistanceToStore:  dist * 1000,
		}, nil
	}

	// Logistics Method: Hook into Gojek/Grab Mock API
	return FulfillmentStatus{
		Method:     FulfillmentCourier,
		TrackingID: "GK-123-RESCUE",
	}, nil
}

// Helper for haversine (assuming exists in engine.go, otherwise redefined)
func calculateHaversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0 // Earth radius in km
	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*(math.Pi/180.0))*math.Cos(lat2*(math.Pi/180.0))*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

