package matching

import (
	"context"
	"fmt"
	"log"

	"github.com/albnnaardy11/pahlawan-pangan/internal/outbox"
	"github.com/segmentio/encoding/json"
)

// RematchWorker orchestrates the re-routing of surplus
// when the primary match fails to respond or is offline.
type RematchWorker struct {
	engine     *MatchingEngine
	repository *SurplusRepository // Mock or abstract DB access
}

func NewRematchWorker(engine *MatchingEngine, repo *SurplusRepository) *RematchWorker {
	return &RematchWorker{
		engine:     engine,
		repository: repo,
	}
}

// SurplusRepository abstracting DB interactions
type SurplusRepository struct {
	// In reality, this would wrap sql.DB
}

// HandleRematchEvent processes a RematchRequired event
func (w *RematchWorker) HandleRematchEvent(ctx context.Context, event outbox.OutboxEvent) error {
	var payload struct {
		SurplusID     string   `json:"surplus_id"`
		ExcludedNGOs []string `json:"excluded_ngos"`
	}

	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal rematch payload: %w", err)
	}

	log.Printf("[Rematch] Triggering rematch for surplus %s. Excluded: %v", payload.SurplusID, payload.ExcludedNGOs)

	// 1. Fetch surplus details
	surplus, err := w.repository.GetSurplus(ctx, payload.SurplusID)
	if err != nil {
		return err
	}

	// 2. Fetch nearby NGOs excluding previous ones
	candidates, err := w.repository.FindNearbyNGOs(ctx, surplus.Lat, surplus.Lon, payload.ExcludedNGOs)
	if err != nil {
		return err
	}

	if len(candidates) == 0 {
		return fmt.Errorf("no more candidates for surplus %s", payload.SurplusID)
	}

	// 3. Use MatchingEngine to find optimal successor
	nextNGO, err := w.engine.MatchNGO(ctx, *surplus, candidates)
	if err != nil {
		return err
	}

	// 4. Update Surplus status and notify
	// Implementation would involve a transaction to update DB and outbox
	log.Printf("[Rematch] Succesfully re-routed surplus %s to NGO %s", payload.SurplusID, nextNGO.ID)

	return nil
}

// Mock methods to satisfy implementation
func (r *SurplusRepository) GetSurplus(ctx context.Context, id string) (*Surplus, error) {
	return &Surplus{ID: id, Lat: -6.2, Lon: 106.8}, nil
}

func (r *SurplusRepository) FindNearbyNGOs(ctx context.Context, lat, lon float64, excluded []string) ([]NGO, error) {
	return []NGO{{ID: "ngo-next", Lat: -6.21, Lon: 106.81}}, nil
}
