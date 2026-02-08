package trust

import (
	"context"
	"math"
)

// TrustService calculates the comprehensive Pahlawan Score
type TrustService struct {
	// In real setup, dependencies: DB, TransactionHistoryRepo
}

func NewTrustService() *TrustService {
	return &TrustService{}
}

// ScoreFactors for calculating trust
type ScoreFactors struct {
	TotalPickups      int
	GhostingIncidents int // No-show count
	DisputeCount      int
	AccountAgeDays    int
	VerifiedIdentity  bool
}

// CalculateScore computes the Credit Score (0-850)
// 0-300: High Risk (Cash Only)
// 301-600: Moderate (Standard Access)
// 601-850: Pahlawan (Priority Access, Higher Discounts)
func (s *TrustService) CalculateScore(ctx context.Context, factors ScoreFactors) int {
	score := 300.0 // Base Score

	// Positive Factors
	score += float64(factors.TotalPickups) * 5.0 // +5 per successful rescue
	if factors.VerifiedIdentity {
		score += 50.0 // Instant boost for KYC
	}
	score += math.Min(float64(factors.AccountAgeDays)*0.5, 100.0) // Loyalty bonus

	// Negative Factors (Heavy penalty)
	score -= float64(factors.GhostingIncidents) * 100.0 // Huge penalty for ghosting
	score -= float64(factors.DisputeCount) * 50.0

	// Cap the score
	if score < 0 {
		return 0
	}
	if score > 850 {
		return 850
	}

	return int(score)
}

// GetTrustLevel translates score to badge
func (s *TrustService) GetTrustLevel(score int) string {
	switch {
	case score >= 750:
		return "UNICORN_SAVIOR" // Top Tier
	case score >= 600:
		return "PAHLAWAN"
	case score >= 400:
		return "WARGA_BAIK"
	default:
		return "PELUANG_KEDUA" // Needs improvement
	}
}
