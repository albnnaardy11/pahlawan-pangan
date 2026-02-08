package service

import (
	"context"
	"fmt"
	"time"

	"github.com/albnnaardy11/pahlawan-pangan/internal/escrow/domain"
	"github.com/google/uuid"
)

type EscrowService struct {
	// In production: Postgres + Kafka
	events []domain.EscrowEvent // In-memory store for demo
}

func NewEscrowService() *EscrowService {
	return &EscrowService{
		events: make([]domain.EscrowEvent, 0),
	}
}

// SecurePayment locks funds using Append-Only Log
func (s *EscrowService) SecurePayment(ctx context.Context, orderID string, amount float64) error {
	event := domain.EscrowEvent{
		ID:        uuid.New().String(),
		OrderID:   orderID,
		Type:      domain.PaymentCollected, // Money is SAFE in Escrow
		Amount:    amount,
		Timestamp: time.Now(),
	}
	s.events = append(s.events, event)
	return nil
}

// ReleaseFunds transfers money to Courier/Provider after delivery confirmation
func (s *EscrowService) ReleaseFunds(ctx context.Context, orderID string) error {
	// Rehydrate to check if eligible
	state := s.getAggregatedState(orderID)
	if state.Status != "CONFIRMED_PENDING_RELEASE" {
		return fmt.Errorf("cannot release funds: state is %s", state.Status)
	}

	event := domain.EscrowEvent{
		ID:        uuid.New().String(),
		OrderID:   orderID,
		Type:      domain.FundsReleased,
		Timestamp: time.Now(),
	}
	s.events = append(s.events, event)
	return nil
}

func (s *EscrowService) FoodDelivered(ctx context.Context, orderID string) {
	s.events = append(s.events, domain.EscrowEvent{
		ID:        uuid.New().String(),
		OrderID:   orderID,
		Type:      domain.FoodDelivered,
		Timestamp: time.Now(),
	})
}

// Private: Rehydrate State on the fly
func (s *EscrowService) getAggregatedState(orderID string) *domain.EscrowState {
	var relevantEvents []domain.EscrowEvent
	for _, e := range s.events {
		if e.OrderID == orderID {
			relevantEvents = append(relevantEvents, e)
		}
	}
	
	// Default state
	state := &domain.EscrowState{Status: "PENDING", OrderID: orderID}
	
	// Apply events in order
	for _, e := range relevantEvents {
		switch e.Type {
		case domain.PaymentCollected:
			state.Status = "LOCKED"
			state.TotalLocked = e.Amount
		case domain.CourierAssigned:
			state.Status = "PICKUP_IN_PROGRESS"
		case domain.FoodPickedUp:
			state.Status = "DELIVERY_IN_PROGRESS"
		case domain.FoodDelivered:
			state.Status = "CONFIRMED_PENDING_RELEASE"
		case domain.FundsReleased:
			state.Status = "CLOSED"
			state.TotalLocked = 0
		case domain.DisputeRaised:
			state.Status = "DISPUTED"
		}
		state.LastUpdated = e.Timestamp
	}
	return state
}
