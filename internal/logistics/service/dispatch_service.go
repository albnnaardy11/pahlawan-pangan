package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/albnnaardy11/pahlawan-pangan/internal/logistics/domain"
)

type DispatchService struct {
	batchEngine *BatchingEngine
	// escrow logic is separated
	pendingOrders []domain.Order // Queue
}

func NewDispatchService(engine *BatchingEngine) *DispatchService {
	return &DispatchService{
		batchEngine:   engine,
		pendingOrders: make([]domain.Order, 0),
	}
}

// CreateOrder initiates the delivery process
func (s *DispatchService) CreateOrder(ctx context.Context, order domain.Order) (*domain.Order, error) {
	// 1. SLA Logic: Only "The Brain" decides urgency
	order.EnforceSLA() // Checks expiry < 15m -> CRITICAL

	// 2. Validate Constraints
	if order.QuantityKg <= 0 {
		return nil, fmt.Errorf("invalid quantity")
	}

	order.ID = uuid.New().String()
	order.Status = "PENDING_MATCHING"
	order.CreatedAt = time.Now()

	// 3. Queue for Batching
	s.pendingOrders = append(s.pendingOrders, order)

	// 4. Trigger Instant Dispatch for EXPRESS/CRITICAL
	if order.CurrentSLA == domain.SLA_EXPRESS || order.CurrentSLA == domain.SLA_CRITICAL {
		go s.ForceDispatch(order.ID)
	} else {
		// 5. Anti-Stuck Timer (The "HEMAT" Guarantee)
		go func(oid string) {
			time.Sleep(5 * time.Minute)
			s.ForceDispatch(oid) // If still pending after 5m, force assign
		}(order.ID)
	}

	return &order, nil
}

// ForceDispatch bypasses batching wait time
func (s *DispatchService) ForceDispatch(orderID string) {
	// Logic to convert Pending Order -> Active Batch (Size 1)
	// In production: Remove from Redis Queue, Create Batch, assign to Courier
	fmt.Printf("[Dispatch] Forcing order %s to dispatch immediately (SLA Enforcement)\n", orderID)
}

// ProcessBatch runs periodically (e.g., every 30s)
func (s *DispatchService) RunBatchProcessor(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Calculate Optimal Batches
			// Mock courier location for simplicity
			// courierLoc := s2.LatLngFromDegrees(-6.2, 106.8)

			// result, _ := s.batchEngine.CalculateOptimalBatch(ctx, s.pendingOrders, courierLoc)
			// Assign batches...
		}
	}
}
