package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/albnnaardy11/pahlawan-pangan/internal/community/domain"
)

type reviewRepository struct {
	db *sql.DB
}

func NewReviewRepository(db *sql.DB) domain.ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) Create(ctx context.Context, review domain.Review) error {
	query := `
		INSERT INTO reviews (id, user_id, target_id, type, rating, comment, created_at, is_verified)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		review.ID,
		review.UserID,
		review.TargetID,
		review.Type,
		review.Rating,
		review.Comment,
		review.CreatedAt,
		review.IsVerified,
	)
	return err
}

func (r *reviewRepository) GetByTargetID(ctx context.Context, targetID string) ([]domain.Review, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, user_id, rating, comment, type, created_at FROM reviews WHERE target_id = $1 ORDER BY created_at DESC LIMIT 50", targetID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var reviews []domain.Review
	for rows.Next() {
		var rev domain.Review
		var createdAt time.Time
		if err := rows.Scan(&rev.ID, &rev.UserID, &rev.Rating, &rev.Comment, &rev.Type, &createdAt); err != nil {
			continue
		}
		rev.CreatedAt = createdAt
		reviews = append(reviews, rev)
	}
	return reviews, nil
}

func (r *reviewRepository) GetSummary(ctx context.Context, targetID string) (domain.CommunitySummary, error) {
	var summary domain.CommunitySummary
	err := r.db.QueryRowContext(ctx, `
		SELECT 
			COALESCE(AVG(rating), 0), 
			COUNT(*) 
		FROM reviews 
		WHERE target_id = $1
	`, targetID).Scan(&summary.AverageRating, &summary.TotalReviews)

	summary.TargetID = targetID
	summary.TopTags = []string{"Trusted", "Quick Response"} // Mock tags specific to vendor (can be dynamic later)

	return summary, err
}
