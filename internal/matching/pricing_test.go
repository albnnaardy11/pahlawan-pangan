package matching

import (
	"testing"
	"time"
)

func TestCalculatePrice(t *testing.T) {
	engine := NewPricingEngine()
	originalPrice := 100.0
	postedAt := time.Now().Add(-1 * time.Hour)
	expiryAt := time.Now().Add(1 * time.Hour)

	// Mid point (1 hour elapsed out of 2 hours total)
	// Progress = 0.5
	// Multiplier = e^(-2 * 0.5) = e^-1 approx 0.36
	price := engine.CalculatePrice(originalPrice, postedAt, expiryAt)
	
	if price <= 0 || price >= originalPrice {
		t.Errorf("Price should be between 0 and original price, got %f", price)
	}

	if price > 40 || price < 30 {
		t.Errorf("Price at midpoint should be around 36, got %f", price)
	}
}

func TestImpactPoints(t *testing.T) {
	engine := NewPricingEngine()
	points := engine.CalculateImpactPoints(10.0, 60.0) // 10kg, 60 mins before expiry
	
	expected := 100 + 6 // 10*10 + 60/10
	if points != expected {
		t.Errorf("Expected %d points, got %d", expected, points)
	}
}
