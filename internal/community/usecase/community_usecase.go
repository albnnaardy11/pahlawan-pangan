package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/albnnaardy11/pahlawan-pangan/internal/community/domain"
)

type communityUsecase struct {
	repo domain.ReviewRepository
	// In production: NLP Sentiment Analysis
}

func NewCommunityUsecase(repo domain.ReviewRepository) domain.CommunityUsecase {
	return &communityUsecase{repo: repo}
}

func (u *communityUsecase) SubmitReview(ctx context.Context, review domain.Review) error {
	// 1. Validation (Profanity Filter Demo)
	badWords := []string{"scam", "cheat", "bad"} // Simple list
	for _, word := range badWords {
		if strings.Contains(strings.ToLower(review.Comment), word) {
			// Instead of error, flag for moderation but save
			review.Comment = "[Flagged for Review] " + review.Comment
		}
	}

	review.ID = uuid.New().String()
	review.CreatedAt = time.Now()

	// 2. Persist
	return u.repo.Create(ctx, review)
}

func (u *communityUsecase) GetReviews(ctx context.Context, targetID string) ([]domain.Review, error) {
	return u.repo.GetByTargetID(ctx, targetID)
}

func (u *communityUsecase) GetVendorReputation(ctx context.Context, vendorID string) (domain.CommunitySummary, error) {
	summary, err := u.repo.GetSummary(ctx, vendorID)
	if err != nil {
		return domain.CommunitySummary{}, err
	}

	// Add "Unicorn Badge" logic
	if summary.AverageRating > 4.8 && summary.TotalReviews > 100 {
		summary.TopTags = append(summary.TopTags, "🌟 Pahlawan Super")
	}

	return summary, nil
}
