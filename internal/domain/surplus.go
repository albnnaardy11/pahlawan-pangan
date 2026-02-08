package domain

import (
	"context"
	"time"
)

// SurplusItem represents the core entity
type SurplusItem struct {
	ID              string           `json:"id" validate:"required,uuid"`
	ProviderID      string           `json:"provider_id" validate:"required"`
	FoodType        string           `json:"food_type" validate:"required"`
	QuantityKgs     float64          `json:"quantity_kgs" validate:"required,gt=0"`
	OriginalPrice   float64          `json:"original_price" validate:"required,gte=0"`
	DiscountPrice   float64          `json:"discount_price" validate:"required,gte=0"`
	Status          string           `json:"status" validate:"required,oneof=available claimed expired"`
	ExpiryTime      time.Time        `json:"expiry_time" validate:"required,gt"`
	Latitude        float64          `json:"lat" validate:"required,latitude"`
	Longitude       float64          `json:"lon" validate:"required,longitude"`
	S2CellID        uint64           `json:"s2_cell_id"`    // Google S2 Index
	Version         int64            `json:"version"`       // Optimistic Locking
	EscrowStatus    string           `json:"escrow_status"` // pending, locked, released
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	NutritionReport *NutritionReport `json:"nutrition_report,omitempty"`
}

// NutritionReport contains AI-generated nutritional analysis of food items.
type NutritionReport struct {
	Calories       string            `json:"calories"`
	Macronutrients map[string]string `json:"macronutrients"`
	HealthScore    string            `json:"health_score"`
	Advice         string            `json:"advice"`
}

// SurplusRepository defines the data store contract
type SurplusRepository interface {
	GetByID(ctx context.Context, id string) (*SurplusItem, error)
	Fetch(ctx context.Context, lat, lon float64, radius int) ([]SurplusItem, error)
	Store(ctx context.Context, item *SurplusItem) error
	Update(ctx context.Context, item *SurplusItem) error
}

// SurplusUsecase defines the business logic contract
type SurplusUsecase interface {
	PostSurplus(ctx context.Context, item *SurplusItem) error
	GetMarketplace(ctx context.Context, lat, lon float64) ([]SurplusItem, error)
	Claim(ctx context.Context, surplusID string, ngoID string) error
	AnalyzeFreshness(ctx context.Context, image []byte) (*NutritionReport, error)
}
