package domain

import (
	"context"
	"time"
)

type EventType string

const (
	PaymentCollected EventType = "PaymentCollected"
	CourierAssigned  EventType = "CourierAssigned"
	FoodPickedUp     EventType = "FoodPickedUp"
	FoodDelivered    EventType = "FoodDelivered"
	FundsReleased    EventType = "FundsReleased"
	OrderCancelled   EventType = "OrderCancelled"
	DisputeRaised    EventType = "DisputeRaise"
)

// EscrowEvent represents an immutable fact in the financial ledger
type EscrowEvent struct {
	ID        string    `json:"id"`
	OrderID   string    `json:"order_id"`
	Amount    float64   `json:"amount"`
	Type      EventType `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Payload   string    `json:"payload"`
}

type EscrowState struct {
	OrderID     string
	TotalLocked float64
	Status      string
	LastUpdated time.Time
}

// Repository defines the contract for storing events (Event Store)
type Repository interface {
	Save(ctx context.Context, event EscrowEvent) error
	GetEventsByOrder(ctx context.Context, orderID string) ([]EscrowEvent, error)
}

// ApplyEvents reconstructs the current state from history (Rehydration)
func RehydrateState(events []EscrowEvent) *EscrowState {
	state := &EscrowState{Status: "PENDING"}
	for _, e := range events {
		switch e.Type {
		case PaymentCollected:
			state.TotalLocked = e.Amount
			state.Status = "LOCKED"
		case FoodDelivered:
			state.Status = "CONFIRMED_PENDING_RELEASE"
		case FundsReleased:
			state.TotalLocked = 0
			state.Status = "CLOSED"
		case DisputeRaised:
			state.Status = "DISPUTED"
		}
		state.LastUpdated = e.Timestamp
	}
	return state
}
