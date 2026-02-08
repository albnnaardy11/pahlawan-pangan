package domain

import (
	"time"

	"github.com/golang/geo/s2"
)

// DeliverySLA defines the service level agreement urgency
type DeliverySLA string

const (
	SLA_EXPRESS  DeliverySLA = "EXPRESS"  // P2P, No batching
	SLA_STANDARD DeliverySLA = "STANDARD" // Max 1 extra stop
	SLA_HEMAT    DeliverySLA = "HEMAT"    // Max 4 stops (High Density)
	SLA_CRITICAL DeliverySLA = "CRITICAL" // Forced upgrade if expiry < 15m
)

// Order represents a delivery request
type Order struct {
	ID          string
	UserID      string
	ProviderID  string
	PickupLat   float64
	PickupLon   float64
	DropoffLat  float64
	DropoffLon  float64
	ExpiryTime  time.Time
	QuantityKg  float64
	SelectedSLA DeliverySLA
	CurrentSLA  DeliverySLA // Can be upgraded by the system
	Status      string
	BatchID     *string
	CreatedAt   time.Time
}

// Batch represents a grouped set of orders for optimal routing
type Batch struct {
	ID        string
	CourierID string
	Orders    []Order
	Route     []RoutePoint
	Score     float64
}

type RoutePoint struct {
	OrderID string
	Type    string // PICKUP or DROPOFF
	Lat     float64
	Lon     float64
	ETA     time.Time
}

// S2CellID returns the Level 13 cell ID for clustering
func (o *Order) S2CellID() s2.CellID {
	return s2.CellIDFromLatLng(s2.LatLngFromDegrees(o.PickupLat, o.PickupLon)).Parent(13)
}

// EnforceSLA upgrades priority if food is about to expire
func (o *Order) EnforceSLA() {
	timeLeft := time.Until(o.ExpiryTime)
	if timeLeft < 15*time.Minute {
		o.CurrentSLA = SLA_CRITICAL
	} else {
		o.CurrentSLA = o.SelectedSLA
	}
}
