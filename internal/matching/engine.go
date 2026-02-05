package matching

import (
	"context"
	"errors"
	"math"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
)

// Constants for performance and logic
const (
	RoutingTimeout    = 200 * time.Millisecond
	MaxWorkerPoolSize = 1000
)

var (
	tracer = otel.Tracer("matching-engine")
	meter  = otel.GetMeterProvider().Meter("matching-engine")

	// Custom Metrics
	claimLatency, _ = meter.Float64Histogram("surplus_claim_latency_seconds")
	wastePrevented, _ = meter.Float64Counter("food_waste_prevented_tons_total")
	// engineSaturation, _ = meter.Float64Gauge("matching_engine_saturation_ratio")
)

// Surplus represents a food donation offer
type Surplus struct {
	ID          string    `json:"id"`
	ProviderID  string    `json:"provider_id"`
	Lat         float64   `json:"lat"`
	Lon         float64   `json:"lon"`
	ExpiryTime  time.Time `json:"expiry_time"`
	QuantityKgs float64   `json:"quantity_kgs"`
}

// NGO represents a receiving entity
type NGO struct {
	ID   string  `json:"id"`
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
}

// Router interface for external routing engines
type Router interface {
	GetTravelTime(ctx context.Context, startLat, startLon, endLat, endLon float64) (time.Duration, error)
}

// MatchingEngine handles the assignment of surplus to NGOs
type MatchingEngine struct {
	router         Router
	circuitBreaker *CircuitBreaker
	workerPool     chan struct{}
}

func NewMatchingEngine(router Router) *MatchingEngine {
	return &MatchingEngine{
		router:         router,
		circuitBreaker: NewCircuitBreaker(3, 10*time.Second),
		workerPool:     make(chan struct{}, MaxWorkerPoolSize),
	}
}

// MatchNGO finds the optimal NGO for a surplus post
func (e *MatchingEngine) MatchNGO(ctx context.Context, surplus Surplus, candidates []NGO) (*NGO, error) {
	ctx, span := tracer.Start(ctx, "MatchNGO")
	defer span.End()

	start := time.Now()
	defer func() {
		claimLatency.Record(ctx, time.Since(start).Seconds())
	}()

	// Update saturation metric
	e.updateSaturation()

	var bestNGO *NGO
	minDistance := math.MaxFloat64

	// Concurrency: Use a worker pool or simple goroutine with context
	// In a real actor model, this would be handled within a Shard Actor.
	// Here we show a robust concurrent selection pattern.

	type result struct {
		ngo      *NGO
		distance float64
		err      error
	}

	resChan := make(chan result, len(candidates))

	for _, ngo := range candidates {
		go func(n NGO) {
			dist, err := e.getDistance(ctx, surplus.Lat, surplus.Lon, n.Lat, n.Lon)
			resChan <- result{&n, dist, err}
		}(ngo)
	}

	for i := 0; i < len(candidates); i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case res := <-resChan:
			if res.err == nil && res.distance < minDistance {
				minDistance = res.distance
				bestNGO = res.ngo
			}
		}
	}

	if bestNGO == nil {
		return nil, errors.New("no suitable NGO found")
	}

	// Update success metric
	wastePrevented.Add(ctx, surplus.QuantityKgs/1000.0)
	
	return bestNGO, nil
}

// getDistance wraps the routing engine with a circuit breaker and fallback
func (e *MatchingEngine) getDistance(ctx context.Context, lat1, lon1, lat2, lon2 float64) (float64, error) {
	// Anti-Fragile Pattern: Circuit Breaker + Fallback
	var distance float64
	err := e.circuitBreaker.Execute(func() error {
		// Try external routing engine (OSRM/Google Maps)
		childCtx, cancel := context.WithTimeout(ctx, RoutingTimeout)
		defer cancel()

		duration, err := e.router.GetTravelTime(childCtx, lat1, lon1, lat2, lon2)
		if err != nil {
			return err
		}
		distance = duration.Seconds() // Use time as distance metric
		return nil
	})

	if err != nil {
		// Fallback to Haversine calculation (Non-blocking / Local)
		_, span := otel.Tracer("matching-engine").Start(ctx, "HaversineFallback")
		distance = haversine(lat1, lon1, lat2, lon2)
		span.End()
		return distance, nil
	}

	return distance, nil
}

func (e *MatchingEngine) updateSaturation() {
	// saturation := float64(len(e.workerPool)) / float64(MaxWorkerPoolSize)
	// engineSaturation.Record(context.Background(), saturation)
}

// haversine calculates the great-circle distance between two points
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth radius in km
	dLat := (lat2 - lat1) * (math.Pi / 180)
	dLon := (lon2 - lon1) * (math.Pi / 180)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*(math.Pi/180))*math.Cos(lat2*(math.Pi/180))*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

// CircuitBreaker - Simple State Machine implementation
type CircuitBreaker struct {
	mu           sync.Mutex
	status       int // 0: Closed, 1: Open, 2: Half-Open
	failures     int
	threshold    int
	lastFailTime time.Time
	timeout      time.Duration
}

func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{threshold: threshold, timeout: timeout}
}

func (cb *CircuitBreaker) Execute(f func() error) error {
	cb.mu.Lock()
	if cb.status == 1 {
		if time.Since(cb.lastFailTime) > cb.timeout {
			cb.status = 2 // Half-Open
		} else {
			cb.mu.Unlock()
			return errors.New("circuit breaker open")
		}
	}
	cb.mu.Unlock()

	err := f()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		if cb.failures >= cb.threshold {
			cb.status = 1
			cb.lastFailTime = time.Now()
		}
		return err
	}

	// Success
	cb.failures = 0
	cb.status = 0 // Close
	return nil
}
