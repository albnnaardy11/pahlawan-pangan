package domain

import (
	"context"
	"time"
)

// UserImpact represents a user's environmental and social impact metrics.
type UserImpact struct {
	UserID        string    `json:"user_id"`
	TotalSavedKgs float64   `json:"total_saved_kgs"`
	MealsRescued  int       `json:"meals_rescued"`
	CO2Prevented  float64   `json:"co2_prevented_kgs"`
	KarmaPoints   int64     `json:"karma_points"`
	CurrentRank   int       `json:"current_rank"`
	Badges        []string  `json:"badges"`
	LastUpdate    time.Time `json:"last_update"`
}

// GlobalLeaderboard represents the national or regional impact rankings.
type GlobalLeaderboard struct {
	Region   string       `json:"region"`
	Rankings []UserImpact `json:"rankings"`
	TopCity  string       `json:"top_city"`
	TotalCO2 float64      `json:"total_national_co2_saved"`
}

// ImpactUsecase defines the business logic interface for impact tracking and leaderboards.
type ImpactUsecase interface {
	GetUserImpact(ctx context.Context, userID string) (*UserImpact, error)
	GetNationalLeaderboard(ctx context.Context, region string) (*GlobalLeaderboard, error)
	GenerateShareCard(ctx context.Context, claimID string) (string, error) // Returns URL to beautiful image
}
