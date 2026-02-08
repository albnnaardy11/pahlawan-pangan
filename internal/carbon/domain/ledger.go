package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// ImpactFactor defines emission savings per kg by category (kgCO2e/kgFood)
type ImpactFactor float64

const (
	FactorMeat    ImpactFactor = 27.0 // High impact
	FactorDairy   ImpactFactor = 12.0
	FactorProduce ImpactFactor = 2.5
	FactorBread   ImpactFactor = 1.2
	FactorMixed   ImpactFactor = 3.5 // Default average
)

// CarbonEntry represents an immutable record of emission savings
type CarbonEntry struct {
	ID            string    `json:"id"`
	VendorID      string    `json:"vendor_id"`
	OrderID       string    `json:"order_id"`
	FoodCategory  string    `json:"category"`
	WeightKg      float64   `json:"weight_kg"`
	CarbonSavedKg float64   `json:"carbon_saved_kg"`
	Timestamp     time.Time `json:"timestamp"`
	PreviousHash  string    `json:"prev_hash"` // Blockchain Link
	Hash          string    `json:"hash"`      // Current Hash
}

// ESGReport represents a formal certificate for corporate partners
type ESGReport struct {
	VendorID         string    `json:"vendor_id"`
	PeriodStart      time.Time `json:"period_start"`
	PeriodEnd        time.Time `json:"period_end"`
	TotalFoodSaved   float64   `json:"total_food_saved_kg"`
	TotalCarbonSaved float64   `json:"total_carbon_saved_kg"`
	TransactionCount int       `json:"transaction_count"`
	VerificationHash string    `json:"verification_hash"` // Digital Signature
}

// ComputeHash generates a SHA-256 hash for the entry to ensure integrity
func (e *CarbonEntry) ComputeHash() string {
	data := fmt.Sprintf("%s|%s|%s|%.2f|%.2f|%s|%s",
		e.ID, e.VendorID, e.OrderID, e.WeightKg, e.CarbonSavedKg, e.Timestamp, e.PreviousHash)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// CalculateSavings computes CO2e based on category
func CalculateSavings(weight float64, category string) float64 {
	factor := FactorMixed
	switch category {
	case "MEAT":
		factor = FactorMeat
	case "DAIRY":
		factor = FactorDairy
	case "PRODUCE":
		factor = FactorProduce
	case "BREAD":
		factor = FactorBread
	}
	return weight * float64(factor)
}
