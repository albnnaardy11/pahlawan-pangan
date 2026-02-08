package service

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/albnnaardy11/pahlawan-pangan/internal/logistics/domain"
	"github.com/golang/geo/s2"
)

// BatchingEngine handles the complex clustering logic
type BatchingEngine struct {
	// In production: Connect to Redis Geospatial
}

func NewBatchingEngine() *BatchingEngine {
	return &BatchingEngine{}
}

// Haversine calculates distance between two points on Earth (in km)
func Haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth radius in km
	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)
	lat1 = lat1 * (math.Pi / 180.0)
	lat2 = lat2 * (math.Pi / 180.0)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1)*math.Cos(lat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

// CalculateOptimalBatch groups orders based on S2 Cell and SLA constraints
func (e *BatchingEngine) CalculateOptimalBatch(ctx context.Context, pendingOrders []domain.Order, courierLoc s2.LatLng) ([]domain.Batch, error) {
	batches := make(map[s2.CellID]*domain.Batch)

	for _, order := range pendingOrders {
		// 1. HARD CONSTRAINT: EXPRESS or CRITICAL orders are never batched
		if order.CurrentSLA == domain.SLA_EXPRESS || order.CurrentSLA == domain.SLA_CRITICAL {
			// Direct Dispatch (Single batch)
			singleBatch := domain.Batch{
				ID:     "batch-" + order.ID,
				Orders: []domain.Order{order},
				Score:  999.0, // Highest priority
			}
			batches[s2.CellID(randInt())] = &singleBatch // Unique key
			continue
		}

		// 2. Clustering (S2 Level 13 ~1km radius)
		regionID := order.S2CellID()

		// 3. Batch Construction Logic
		if batch, exists := batches[regionID]; exists {
			// Check capacity (HEMAT max 4, STANDARD max 2)
			maxSize := 4
			if order.CurrentSLA == domain.SLA_STANDARD {
				maxSize = 2
			}

			if len(batch.Orders) < maxSize {
				batch.Orders = append(batch.Orders, order)
			} else {
				// Current batch full, create new one (simplified logic)
				// In production: check neighboring cells or create overflow batch
			}
		} else {
			batches[regionID] = &domain.Batch{
				ID:     "batch-" + regionID.ToToken(),
				Orders: []domain.Order{order},
			}
		}
	}


	// 4. Scoring Algorithm
	var prioritizedBatches []domain.Batch
	for _, b := range batches {
		if b == nil { continue }
		score := e.calculateBatchScore(*b, courierLoc)
		b.Score = score
		prioritizedBatches = append(prioritizedBatches, *b)
	}

	// Sort high score first (Pahlawan Priority)
	sort.Slice(prioritizedBatches, func(i, j int) bool {
		return prioritizedBatches[i].Score > prioritizedBatches[j].Score // Descending
	})

	return prioritizedBatches, nil
}

// internal helper to score importance
func (e *BatchingEngine) calculateBatchScore(batch domain.Batch, courierLoc s2.LatLng) float64 {
	var maxExpiryPenalty float64 = 0
	var minSLAUrgency float64 = 0 // Express = 100, Hemat = 10

	// Calculate centroid of batch
	var totalLat, totalLon float64
	for _, o := range batch.Orders {
		totalLat += o.PickupLat
		totalLon += o.PickupLon

		// Expiry penalty (closer to expiry = higher score to pickup)
		timeLeft := time.Until(o.ExpiryTime).Minutes()
		penalty := 1000.0 / (timeLeft + 1) // +1 to avoid div by zero
		if penalty > maxExpiryPenalty {
			maxExpiryPenalty = penalty
		}

		// SLA Urgency
		urgency := 10.0
		if o.CurrentSLA == domain.SLA_EXPRESS { urgency = 100.0 }
		if o.CurrentSLA == domain.SLA_STANDARD { urgency = 50.0 }
		if urgency > minSLAUrgency {
			minSLAUrgency = urgency
		}
	}
	
	avgLat := totalLat / float64(len(batch.Orders))
	avgLon := totalLon / float64(len(batch.Orders))

	// Distance from courier (closer is better)
	distKm := Haversine(courierLoc.Lat.Degrees(), courierLoc.Lng.Degrees(), avgLat, avgLon)
	distanceScore := 100.0 / (distKm + 0.1) // Avoid div zero

	// Weighted Formula: (Distance * 0.3) + (SLA * 0.4) + (Expiry * 0.3)
	finalScore := (distanceScore * 0.3) + (minSLAUrgency * 0.4) + (maxExpiryPenalty * 0.3)
	
	return finalScore
}

func randInt() int64 {
	return time.Now().UnixNano() // Simple unique ID generator for demo
}
