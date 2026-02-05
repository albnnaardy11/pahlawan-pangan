package matching

import (
	"context"
	"testing"
)

func TestAIEngine_PredictWaste(t *testing.T) {
	ai := NewAIEngine()
	ctx := context.Background()
	providerID := "test-provider-123"

	result, err := ai.PredictWaste(ctx, providerID)
	if err != nil {
		t.Fatalf("PredictWaste failed: %v", err)
	}

	if result.PredictedKgs < 5.0 || result.PredictedKgs > 25.0 {
		t.Errorf("PredictedKgs out of expected range (5-25), got %f", result.PredictedKgs)
	}

	if result.Confidence < 0.70 || result.Confidence > 1.0 {
		t.Errorf("Confidence out of expected range (0.7-1.0), got %f", result.Confidence)
	}

	if result.RecommendedAction == "" {
		t.Error("RecommendedAction should not be empty")
	}
}

func TestAIEngine_GenerateHeatmap(t *testing.T) {
	ai := NewAIEngine()
	ctx := context.Background()
	regionID := 1

	heatmap, err := ai.GenerateHeatmapData(ctx, regionID)
	if err != nil {
		t.Fatalf("GenerateHeatmapData failed: %v", err)
	}

	if len(heatmap) == 0 {
		t.Error("Heatmap should not be empty")
	}

	for _, point := range heatmap {
		if _, ok := point["lat"]; !ok {
			t.Error("Heatmap point should have 'lat'")
		}
		if _, ok := point["lon"]; !ok {
			t.Error("Heatmap point should have 'lon'")
		}
		if _, ok := point["intensity"]; !ok {
			t.Error("Heatmap point should have 'intensity'")
		}
	}
}
