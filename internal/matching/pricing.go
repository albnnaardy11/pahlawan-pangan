package matching

import (
	"math"
	"time"
)

// PricingEngine handles unicorn-level dynamic pricing for B2C market
type PricingEngine struct {
	MinPriceRatio float64 // Minimal price (e.g., 0.1 for 10% of original)
}

func NewPricingEngine() *PricingEngine {
	return &PricingEngine{
		MinPriceRatio: 0.1, // Default to 90% discount max
	}
}

// CalculatePrice implements Exponential Decay Pricing
// Formula: Price = Original * e^(-k * t)
// where t is the percentage of time elapsed toward expiry
func (p *PricingEngine) CalculatePrice(originalPrice float64, postedAt, expiryAt time.Time) float64 {
	now := time.Now()
	if now.After(expiryAt) {
		return 0
	}

	totalDuration := expiryAt.Sub(postedAt).Seconds()
	elapsedDuration := now.Sub(postedAt).Seconds()

	if elapsedDuration <= 0 {
		return originalPrice
	}

	// Progress from 0.0 to 1.0
	progress := elapsedDuration / totalDuration

	// Exponential decay: faster price drop as we approach expiry
	// At progress = 1.0, multiplier will be around 0.13
	multiplier := math.Exp(-2.0 * progress)

	finalPrice := originalPrice * multiplier

	// Ensure it doesn't go below floor
	floor := originalPrice * p.MinPriceRatio
	if finalPrice < floor {
		return floor
	}

	return math.Round(finalPrice*100) / 100
}

// CalculateImpactPoints rewards providers based on quantity and speed of rescue
func (p *PricingEngine) CalculateImpactPoints(quantityKgs float64, savedMinutesBeforeExpiry float64) int {
	// Base points from quantity
	base := int(quantityKgs * 10)

	// Bonus points for saving it early (Strava-style competition)
	bonus := int(savedMinutesBeforeExpiry / 10)

	return base + bonus
}
