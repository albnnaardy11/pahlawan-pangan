package loyalty

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// LoyaltyService handles XP and Leaderboard tracking
type LoyaltyService struct {
	redis *redis.Client
}

const (
	LeaderboardKey = "loyalty:leaderboard"
	UserStatsKey   = "loyalty:user:%s" // Stores detailed XP history
)

func NewLoyaltyService(redisClient *redis.Client) *LoyaltyService {
	return &LoyaltyService{redis: redisClient}
}

// AddXP updates the user's score in the global leaderboard (Real-time)
// Uses ZINCRBY which is O(log(N)) - extremely fast for 287M users
func (s *LoyaltyService) AddXP(ctx context.Context, userID string, xp float64) error {
	// 1. Update Leaderboard
	err := s.redis.ZIncrBy(ctx, LeaderboardKey, xp, userID).Err()
	if err != nil {
		return fmt.Errorf("failed to update leaderboard: %w", err)
	}

	// 2. Log Action (for history/audit)
	// In production, push to time-series DB or append to a list
	s.redis.RPush(ctx, fmt.Sprintf(UserStatsKey, userID), fmt.Sprintf("%d: +%.2f XP", time.Now().Unix(), xp))

	return nil
}

// GetTopUsers returns the top N users from the leaderboard
// Uses ZREVRANGE - O(log(N)+M) where M is count
func (s *LoyaltyService) GetTopUsers(ctx context.Context, count int64) ([]redis.Z, error) {
	return s.redis.ZRevRangeWithScores(ctx, LeaderboardKey, 0, count-1).Result()
}

// GetUserRank returns the user's current rank and score
// Uses ZREVRANK - O(log(N))
func (s *LoyaltyService) GetUserRank(ctx context.Context, userID string) (int64, float64, error) {
	pipe := s.redis.Pipeline()
	rankCmd := pipe.ZRevRank(ctx, LeaderboardKey, userID)
	scoreCmd := pipe.ZScore(ctx, LeaderboardKey, userID)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get rank: %w", err)
	}

	return rankCmd.Val(), scoreCmd.Val(), nil
}
