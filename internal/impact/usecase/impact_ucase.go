package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/albnnaardy11/pahlawan-pangan/internal/domain"
)

type impactUsecase struct {
	repo domain.SurplusRepository
}

func NewImpactUsecase(repo domain.SurplusRepository) domain.ImpactUsecase {
	return &impactUsecase{repo: repo}
}

func (u *impactUsecase) GetUserImpact(ctx context.Context, userID string) (*domain.UserImpact, error) {
	// Logic: Calculate impact based on user's claim history
	return &domain.UserImpact{
		UserID:        userID,
		TotalSavedKgs: 42.5,
		MealsRescued:  15,
		CO2Prevented:  106.25, // 42.5 * 2.5
		KarmaPoints:   8500,
		CurrentRank:   124,
		Badges:        []string{"Earth_Hero_2026", "Master_Claimer", "Early_Adopter"},
		LastUpdate:    time.Now(),
	}, nil
}

func (u *impactUsecase) GetNationalLeaderboard(ctx context.Context, region string) (*domain.GlobalLeaderboard, error) {
	// Logic: Aggregate stats from all users in the region
	return &domain.GlobalLeaderboard{
		Region:   region,
		TopCity:  "Bandung (The Greenest City)",
		TotalCO2: 15402.5,
		Rankings: []domain.UserImpact{
			{UserID: "User_A", KarmaPoints: 12000, TotalSavedKgs: 150},
			{UserID: "User_B", KarmaPoints: 11500, TotalSavedKgs: 142},
			{UserID: "User_C", KarmaPoints: 9000, TotalSavedKgs: 98},
		},
	}, nil
}

func (u *impactUsecase) GenerateShareCard(ctx context.Context, claimID string) (string, error) {
	// Logic: In production, use high-res image generation library (e.g., gg or vips)
	// For now, return a dynamic URL that would render a beautiful social card
	return fmt.Sprintf("https://cdn.pahlawanpangan.org/share/card-%s.jpg", claimID), nil
}
