package domain

import (
	"context"
	"time"
)

type ReviewType string

const (
	ReviewTypeFood   ReviewType = "FOOD"
	ReviewTypeVendor ReviewType = "VENDOR"
	ReviewTypeCourier ReviewType = "COURIER"
)

// Review represents a user's feedback
type Review struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	TargetID  string    `json:"target_id"` // SurplusID, VendorID, or CourierID
	Type      ReviewType `json:"type"`
	Rating    int       `json:"rating"` // 1-5
	Comment   string    `json:"comment"`
	Tags      []string  `json:"tags"`   // e.g., ["Fresh", "Big Portion", "Fast"]
	Images    []string  `json:"images"`
	CreatedAt time.Time `json:"created_at"`
	IsVerified bool     `json:"is_verified"` // True if transaction confirmed
}

// CommunitySummary aggregates ratings
type CommunitySummary struct {
	TargetID      string  `json:"target_id"`
	AverageRating float64 `json:"average_rating"`
	TotalReviews  int     `json:"total_reviews"`
	TopTags       []string `json:"top_tags"`
}

// Repository defines data access
type ReviewRepository interface {
	Create(ctx context.Context, review Review) error
	GetByTargetID(ctx context.Context, targetID string) ([]Review, error)
	GetSummary(ctx context.Context, targetID string) (CommunitySummary, error)
}

// Usecase defines business logic
type CommunityUsecase interface {
	SubmitReview(ctx context.Context, review Review) error
	GetReviews(ctx context.Context, targetID string) ([]Review, error)
	GetVendorReputation(ctx context.Context, vendorID string) (CommunitySummary, error)
}
