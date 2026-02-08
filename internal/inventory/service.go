package inventory

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/albnnaardy11/pahlawan-pangan/internal/outbox"
)

// POSWebhookPayload represents the JSON payload from a restaurant POS system
type POSWebhookPayload struct {
	ProviderID    string  `json:"provider_id"`
	SKU           string  `json:"sku"`
	ItemName      string  `json:"item_name"`
	CurrentStock  int     `json:"current_stock"`
	OriginalPrice float64 `json:"original_price"`
}

// InventoryService handles real-time stock updates from merchants
type InventoryService struct {
	outboxRepo outbox.Repository // Using Outbox to trigger surplus creation async
	logger     *zap.Logger
}

func NewInventoryService(outboxRepo outbox.Repository, logger *zap.Logger) *InventoryService {
	return &InventoryService{outboxRepo: outboxRepo, logger: logger}
}

// ProcessStockUpdate handles the "Flash Ludes" logic
func (s *InventoryService) ProcessStockUpdate(ctx context.Context, payload POSWebhookPayload) error {
	// Rule: Creating "Flash Ludes" if stock is low (e.g. < 5) and late at night (e.g. > 9PM)
	// For demo: Always trigger if stock < 5
	const LowStockThreshold = 5

	if payload.CurrentStock > 0 && payload.CurrentStock < LowStockThreshold {
		s.logger.Info("🔥 Flash Ludes Condition Met!",
			zap.String("item", payload.ItemName),
			zap.Int("stock", payload.CurrentStock))

		// Trigger automated surplus creation via Event-Driven Architecture
		// We emit an internal event that another worker picks up to act as the "Auto-Poster"
		eventID := uuid.New().String()
		eventPayload, _ := json.Marshal(map[string]interface{}{
			"action":       "auto_create_surplus",
			"provider_id":  payload.ProviderID,
			"item_name":    payload.ItemName,
			"quantity_qty": payload.CurrentStock,
			"discount":     0.70,                          // 70% off for Flash Ludes
			"expiry":       time.Now().Add(3 * time.Hour), // Quick expiry
		})

		// This decouples the high-speed POS webhook from our heavy database logic
		// Ideally, push to Kafka/NATS. Here we reuse the Outbox table for simplicity.
		// Note: Using nil transaction for simplicity in this example handler
		// In production, pass the DB/Tx context properly.
		return s.outboxRepo.Save(ctx, nil, outbox.Event{
			ID:          eventID,
			AggregateID: payload.ProviderID,
			EventType:   outbox.SurplusPosted, // Reuse existing event or new internal type
			Payload:     eventPayload,
		})
	}

	return nil
}
