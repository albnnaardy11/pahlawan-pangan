package recommendation

import (
	"context"
	"math/rand"
	"time"

	"github.com/golang/geo/s2"
)

// RecommendationService provides personalized smart nudges
type RecommendationService struct {
	// In production: Connect to Feature Store (Feast) and Model Registry (MLflow)
}

func NewRecommendationService() *RecommendationService {
	return &RecommendationService{}
}

type Recommendation struct {
	SurplusID string  `json:"surplus_id"`
	Reason    string  `json:"reason"` // e.g., "Usually ordered at 4 PM"
	Score     float64 `json:"score"`
	DistanceM float64 `json:"distance_m"`
}

// GetSmartNudges returns personalized recommendations using S2 cells and history
func (s *RecommendationService) GetSmartNudges(ctx context.Context, userID string, lat, lon float64) ([]Recommendation, error) {
	// 1. Convert User Location to S2 Cell (Level 13 ~1km radius)
	cellID := s2.CellIDFromLatLng(s2.LatLngFromDegrees(lat, lon)).Parent(13)

	// 2. Logic: "Data Lake" Analysis (Simulated)
	// Query historical patterns for this user in this cell
	// e.g., "User buys bread in Cell X between 16:00-18:00"

	hour := time.Now().Hour()
	var recs []Recommendation

	// Mock Inference Engine response
	if hour >= 16 && hour <= 19 {
		recs = append(recs, Recommendation{
			SurplusID: "surplus-bakery-123",
			Reason:    "It's tea time! 🍵 Your favorite bakery nearby has surplus.",
			Score:     0.95,
			DistanceM: 350,
		})
	}

	// 3. Proximity Bias (S2 Locality)
	// If user is in a high-density food zone (Cell Token match)
	if cellID.ToToken() == "123456" { // Mock token
		recs = append(recs, Recommendation{
			SurplusID: "surplus-pizza-999",
			Reason:    "Hot Pizza just 500m away! 🍕",
			Score:     0.88,
			DistanceM: 500,
		})
	}

	// Add randomization for "Serendipity" (Discovery)
	// #nosec G404
	if rand.Float64() > 0.7 {
		recs = append(recs, Recommendation{
			SurplusID: "surplus-mystery-box",
			Reason:    "Try something new? 🎁 Mystery Box available.",
			Score:     0.75,
			DistanceM: 1200,
		})
	}

	return recs, nil
}
