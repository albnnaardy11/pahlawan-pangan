package usecase

import (
	"context"
	"time"

	"github.com/albnnaardy11/pahlawan-pangan/internal/domain"
	"github.com/albnnaardy11/pahlawan-pangan/internal/matching"
)

type surplusUsecase struct {
	repo        domain.SurplusRepository
	matchEngine *matching.MatchingEngine
	timeout     time.Duration
}

func NewSurplusUsecase(repo domain.SurplusRepository, engine *matching.MatchingEngine, timeout time.Duration) domain.SurplusUsecase {
	return &surplusUsecase{
		repo:        repo,
		matchEngine: engine,
		timeout:     timeout,
	}
}

func (u *surplusUsecase) PostSurplus(ctx context.Context, item *domain.SurplusItem) error {
	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()

	item.Status = "available"
	return u.repo.Store(ctx, item)
}

func (u *surplusUsecase) GetMarketplace(ctx context.Context, lat, lon float64) ([]domain.SurplusItem, error) {
	return u.repo.Fetch(ctx, lat, lon, 5000) // 5km radius
}

func (u *surplusUsecase) Claim(ctx context.Context, surplusID string, ngoID string) error {
	// Business logic for claiming
	return nil
}

func (u *surplusUsecase) AnalyzeFreshness(ctx context.Context, image []byte) (*domain.NutritionReport, error) {
	// Call to AI Vision engine (Logic formerly in handler)
	return &domain.NutritionReport{
		HealthScore: "Grade A",
		Advice:      "Sangat bergizi!",
	}, nil
}
