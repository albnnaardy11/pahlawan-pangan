package matching

import (
	"context"
	"testing"
	"time"
)

func TestMatchingEngine_MatchNGO(t *testing.T) {
	router := &MockRouter{}
	engine := NewMatchingEngine(router)

	surplus := Surplus{
		ID:          "surplus-1",
		ProviderID:  "provider-1",
		Lat:         -6.2088,
		Lon:         106.8456,
		ExpiryTime:  time.Now().Add(2 * time.Hour),
		QuantityKgs: 50.0,
	}

	candidates := []NGO{
		{ID: "ngo-1", Lat: -6.2100, Lon: 106.8460},
		{ID: "ngo-2", Lat: -6.2200, Lon: 106.8500},
		{ID: "ngo-3", Lat: -6.1900, Lon: 106.8400},
	}

	ctx := context.Background()
	bestNGO, err := engine.MatchNGO(ctx, surplus, candidates)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if bestNGO == nil {
		t.Fatal("Expected a matched NGO, got nil")
	}

	// NGO-1 should be closest
	if bestNGO.ID != "ngo-1" {
		t.Errorf("Expected ngo-1, got %s", bestNGO.ID)
	}
}

func TestMatchingEngine_ContextCancellation(t *testing.T) {
	router := &SlowRouter{}
	engine := NewMatchingEngine(router)

	surplus := Surplus{
		ID:          "surplus-1",
		Lat:         -6.2088,
		Lon:         106.8456,
		ExpiryTime:  time.Now().Add(2 * time.Hour),
		QuantityKgs: 50.0,
	}

	candidates := []NGO{
		{ID: "ngo-1", Lat: -6.2100, Lon: 106.8460},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := engine.MatchNGO(ctx, surplus, candidates)

	if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded, got %v", err)
	}
}

func TestCircuitBreaker_OpenState(t *testing.T) {
	cb := NewCircuitBreaker(3, 1*time.Second)

	// Trigger failures
	for i := 0; i < 3; i++ {
		_ = cb.Execute(func() error {
			return context.DeadlineExceeded
		})
	}

	// Circuit should be open
	err := cb.Execute(func() error {
		return nil
	})

	if err == nil || err.Error() != "circuit breaker open" {
		t.Errorf("Expected circuit breaker open error, got %v", err)
	}
}

func TestHaversine(t *testing.T) {
	// Jakarta to Bandung (~120km)
	distance := haversine(-6.2088, 106.8456, -6.9175, 107.6191)

	if distance < 100 || distance > 150 {
		t.Errorf("Expected distance ~120km, got %.2f", distance)
	}
}

// MockRouter for testing
type MockRouter struct{}

func (m *MockRouter) GetTravelTime(ctx context.Context, startLat, startLon, endLat, endLon float64) (time.Duration, error) {
	// Return distance-based time
	dist := haversine(startLat, startLon, endLat, endLon)
	return time.Duration(dist*60) * time.Second, nil
}

// SlowRouter simulates slow API
type SlowRouter struct{}

func (m *SlowRouter) GetTravelTime(ctx context.Context, startLat, startLon, endLat, endLon float64) (time.Duration, error) {
	time.Sleep(500 * time.Millisecond)
	return 15 * time.Minute, nil
}

func BenchmarkMatchNGO(b *testing.B) {
	router := &MockRouter{}
	engine := NewMatchingEngine(router)

	surplus := Surplus{
		ID:          "surplus-1",
		Lat:         -6.2088,
		Lon:         106.8456,
		ExpiryTime:  time.Now().Add(2 * time.Hour),
		QuantityKgs: 50.0,
	}

	candidates := make([]NGO, 100)
	for i := 0; i < 100; i++ {
		candidates[i] = NGO{
			ID:  "ngo-" + string(rune(i)),
			Lat: -6.2 + float64(i)*0.01,
			Lon: 106.8 + float64(i)*0.01,
		}
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.MatchNGO(ctx, surplus, candidates)
	}
}
