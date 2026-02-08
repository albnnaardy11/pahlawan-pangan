package worker

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/albnnaardy11/pahlawan-pangan/internal/carbon/service"
	"github.com/albnnaardy11/pahlawan-pangan/internal/outbox"
)

// CarbonWorker listens for completed deliveries and records carbon impact
type CarbonWorker struct {
	carbonSvc *service.CarbonService
	nc        *nats.Conn
	logger    *zap.Logger
}

func NewCarbonWorker(carbonSvc *service.CarbonService, nc *nats.Conn, logger *zap.Logger) *CarbonWorker {
	return &CarbonWorker{
		carbonSvc: carbonSvc,
		nc:        nc,
		logger:    logger,
	}
}

func (w *CarbonWorker) Start(ctx context.Context) error {
	w.logger.Info("Starting Carbon Ledger Worker")

	// Subscribe to delivery completed events
	_, err := w.nc.Subscribe("delivery.completed", func(m *nats.Msg) {
		var event outbox.Event
		if err := json.Unmarshal(m.Data, &event); err != nil {
			w.logger.Error("Failed to unmarshal carbon event", zap.Error(err))
			return
		}

		// Payload extraction (Mocking payload structure for demo)
		// In production, the payload would contain Category and Weight
		var payload struct {
			VendorID string  `json:"vendor_id"`
			Category string  `json:"category"`
			WeightKg float64 `json:"weight_kg"`
		}
		_ = json.Unmarshal(event.Payload, &payload)

		// Defaults for demo if payload empty
		if payload.WeightKg == 0 {
			payload.WeightKg = 5.0
			payload.Category = "PRODUCE"
		}

		hash, err := w.carbonSvc.RecordSavings(ctx, payload.VendorID, event.AggregateID, payload.Category, payload.WeightKg)
		if err != nil {
			w.logger.Error("Failed to record carbon savings", zap.Error(err))
			return
		}

		w.logger.Info("✅ Carbon Footprint Recorded",
			zap.String("order_id", event.AggregateID),
			zap.String("hash", hash),
			zap.Float64("savings_kg", payload.WeightKg*2.5),
		)
	})

	return err
}
