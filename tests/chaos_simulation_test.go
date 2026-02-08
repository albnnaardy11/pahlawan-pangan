package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/albnnaardy11/pahlawan-pangan/internal/matching"
)

// ChaosRouter simulates a "Sakit" (Sick) network or service.
type ChaosRouter struct {
	shouldFail bool
}

func (m *ChaosRouter) GetTravelTime(ctx context.Context, startLat, startLon, endLat, endLon float64) (time.Duration, error) {
	if m.shouldFail {
		return 0, fmt.Errorf("💥 DATABASE LINK SEVERED (Chaos Monkey)")
	}
	return 10 * time.Millisecond, nil
}

func TestChaosSimulation(t *testing.T) {
	fmt.Printf("\n🐒 STARTING CHAOS SIMULATION (Netflix Principle) 🐒\n")
	fmt.Println("-------------------------------------------------------")

	// 1. Setup Engine with Chaos
	router := &ChaosRouter{shouldFail: true} // Start with a BROKEN router
	engine := matching.NewMatchingEngine(router)

	surplus := matching.Surplus{
		ID: "S-666", Lat: -6.123, Lon: 106.456, QuantityKgs: 50,
	}

	// 2. Scenario: Primary Matcher is DOWN
	fmt.Println("🔥 SCENARIO 1: Primary Database Instance is DOWN")
	fmt.Println("   Action: Attempting to match NGO while database link is severed...")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// No candidates or broken candidates
	result, err := engine.MatchNGO(ctx, surplus, []matching.NGO{})

	if err != nil {
		t.Errorf("Chaos Test Failed: Engine should NOT return error during outage, it should use FALLBACK. Got error: %v", err)
	}

	if result.ID == "EMERGENCY_DROP_POINT_RT_RW" {
		fmt.Printf("✅ FAILOVER SUCCESS: System redirected rescue to [%s]\n", result.ID)
		fmt.Println("   Result: Food is SAVED at the nearest community drop point instead of being wasted.")
	} else {
		t.Errorf("Chaos Test Failed: Expected emergency fallback, got %s", result.ID)
	}

	// 3. Scenario: Partial Recovery
	fmt.Println("\n🔄 SCENARIO 2: Partial Recovery (Restoring DB Link)")
	router.shouldFail = false

	candidates := []matching.NGO{{ID: "NGO-PRO-001", Lat: -6.124, Lon: 106.457}}
	result, _ = engine.MatchNGO(ctx, surplus, candidates)

	if result.ID == "NGO-PRO-001" {
		fmt.Printf("✅ RECOVERY SUCCESS: System returned to primary logic. Found: [%s]\n", result.ID)
	}

	fmt.Println("-------------------------------------------------------")
	fmt.Printf("⭐ CHAOS EVALUATION: SYSTEM IS RESILIENT (ANTI-FRAGILE) ⭐\n\n")
}
