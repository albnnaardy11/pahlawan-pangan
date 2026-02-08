package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/albnnaardy11/pahlawan-pangan/internal/logistics/domain"
	"github.com/albnnaardy11/pahlawan-pangan/internal/logistics/service"
)

type LogisticsHandler struct {
	dispatchSvc *service.DispatchService
}

func NewLogisticsHandler(svc *service.DispatchService) *LogisticsHandler {
	return &LogisticsHandler{dispatchSvc: svc}
}

// POST /api/v1/orders
// Payload with SLA and Expiry
func (h *LogisticsHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PickupLat  float64            `json:"pickup_lat"`
		PickupLon  float64            `json:"pickup_lon"`
		DropoffLat float64            `json:"dropoff_lat"`
		DropoffLon float64            `json:"dropoff_lon"`
		Expiry     time.Time          `json:"expiry_timestamp"`
		SLA        domain.DeliverySLA `json:"service_level"`
		Quantity   float64            `json:"quantity_kg"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	order := domain.Order{
		PickupLat:   req.PickupLat,
		PickupLon:   req.PickupLon,
		DropoffLat:  req.DropoffLat,
		DropoffLon:  req.DropoffLon,
		ExpiryTime:  req.Expiry,
		SelectedSLA: req.SLA,
		QuantityKg:  req.Quantity,
	}

	// Dispatch logic handles SLA enforcement automatically (The "15 Minute Rule")
	createdOrder, err := h.dispatchSvc.CreateOrder(r.Context(), order)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(createdOrder)
}

// GET /api/v1/courier/itinerary
// Returns logical route sequence
func (h *LogisticsHandler) GetCourierItinerary(w http.ResponseWriter, r *http.Request) {
	// Mock implementation for demo
	// In production: Query assigned Batch from Redis
	courierID := chi.URLParam(r, "id")

	itinerary := []map[string]interface{}{
		{"seq": 1, "type": "PICKUP", "loc": "Mall A", "eta": "10:00"},
		{"seq": 2, "type": "PICKUP", "loc": "Bakery B", "eta": "10:15"}, // Optimization: 2 Pickups
		{"seq": 3, "type": "DROPOFF", "loc": "User B", "eta": "10:30"},  // Drop B first (Optimization)
		{"seq": 4, "type": "DROPOFF", "loc": "User A", "eta": "10:45"},
	}

	res := map[string]interface{}{
		"courier_id": courierID,
		"route":      itinerary,
		"status":     "OPTIMIZED",
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

func (h *LogisticsHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/orders", h.CreateOrder)
	r.Get("/courier/{id}/itinerary", h.GetCourierItinerary)
	return r
}
