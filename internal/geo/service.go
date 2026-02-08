package geo

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// GeoService handles high-throughput location updates and radius queries
type GeoService struct {
	redis *redis.Client
}

const (
	UserLocationsKey = "geo:user_locations"
	GeoExpiry        = 24 * time.Hour // User location expires after 24h
)

func NewGeoService(redisClient *redis.Client) *GeoService {
	return &GeoService{redis: redisClient}
}

// UpdateUserLocation updates user's current GPS coordinates (called by mobile app every 5 min)
func (s *GeoService) UpdateUserLocation(ctx context.Context, userID string, lat, lon float64) error {
	// Add to GEO index
	err := s.redis.GeoAdd(ctx, UserLocationsKey, &redis.GeoLocation{
		Name:      userID,
		Latitude:  lat,
		Longitude: lon,
	}).Err()

	// Also update a standard key to track "last active" time if needed
	// s.redis.Set(ctx, "last_active:"+userID, time.Now().Unix(), GeoExpiry)

	return err
}

// FindUsersNearby finds all users within `radius` meters of (lat, lon)
// Cost: O(N+log(M)) where N is number of nearby users, M is total users. Extremely fast.
func (s *GeoService) FindUsersNearby(ctx context.Context, lat, lon, radiusMeters float64) ([]string, error) {
	// Use GEORADIUS (or GEOSEARCH in newer Redis)
	locations, err := s.redis.GeoRadius(ctx, UserLocationsKey, lon, lat, &redis.GeoRadiusQuery{
		Radius:      radiusMeters,
		Unit:        "m",
		WithCoord:   false,
		WithDist:    false,
		WithGeoHash: false,
		Count:       10000, // Limit to 10k users per blast to prevent overloading notification service
		Sort:        "ASC", // Send to closest users first
	}).Result()

	if err != nil {
		return nil, fmt.Errorf("redis geo error: %w", err)
	}

	userIDs := make([]string, len(locations))
	for i, loc := range locations {
		userIDs[i] = loc.Name
	}

	return userIDs, nil
}
