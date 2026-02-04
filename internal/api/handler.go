package api

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/antigravity/pahlawan-pangan/internal/matching"
	"github.com/antigravity/pahlawan-pangan/internal/outbox"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/segmentio/encoding/json"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var tracer = otel.Tracer("api-handler")

type Handler struct {
	db            *sql.DB
	matchEngine   *matching.MatchingEngine
	outboxService *outbox.OutboxService
}

func NewHandler(db *sql.DB, engine *matching.MatchingEngine, outbox *outbox.OutboxService) *Handler {
	return &Handler{
		db:            db,
		matchEngine:   engine,
		outboxService: outbox,
	}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Health checks
	r.Get("/health/live", h.LivenessCheck)
	r.Get("/health/ready", h.ReadinessCheck)

	// API routes with OpenTelemetry instrumentation
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(otelhttp.NewMiddleware("api"))

		// Surplus endpoints
		r.Post("/surplus", h.PostSurplus)
		r.Get("/surplus/{id}", h.GetSurplus)
		r.Post("/surplus/{id}/claim", h.ClaimSurplus)

		// NGO endpoints
		r.Get("/ngos/nearby", h.GetNearbyNGOs)
	})

	return r
}

type PostSurplusRequest struct {
	ProviderID  string    `json:"provider_id"`
	Lat         float64   `json:"lat"`
	Lon         float64   `json:"lon"`
	QuantityKgs float64   `json:"quantity_kgs"`
	FoodType    string    `json:"food_type"`
	ExpiryTime  time.Time `json:"expiry_time"`
}

func (h *Handler) PostSurplus(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "PostSurplus")
	defer span.End()

	var req PostSurplusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate
	if req.QuantityKgs <= 0 || req.ExpiryTime.Before(time.Now()) {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	surplusID := uuid.New().String()

	// Start transaction
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Insert surplus
	_, err = tx.ExecContext(ctx, `
		INSERT INTO surplus (id, provider_id, location, quantity_kgs, food_type, expiry_time, status, geo_region_id, created_at)
		SELECT $1, $2, ST_SetSRID(ST_MakePoint($3, $4), 4326), $5, $6, $7, 'available',
		       (SELECT id FROM geo_regions WHERE ST_Contains(geometry, ST_SetSRID(ST_MakePoint($3, $4), 4326)) LIMIT 1),
		       NOW()
	`, surplusID, req.ProviderID, req.Lon, req.Lat, req.QuantityKgs, req.FoodType, req.ExpiryTime)

	if err != nil {
		span.RecordError(err)
		http.Error(w, "failed to create surplus", http.StatusInternalServerError)
		return
	}

	// Create outbox event
	payload, _ := json.Marshal(map[string]interface{}{
		"surplus_id":   surplusID,
		"provider_id":  req.ProviderID,
		"lat":          req.Lat,
		"lon":          req.Lon,
		"quantity_kgs": req.QuantityKgs,
		"expiry_time":  req.ExpiryTime,
	})

	event := outbox.OutboxEvent{
		ID:          uuid.New().String(),
		AggregateID: surplusID,
		EventType:   outbox.SurplusPosted,
		Payload:     payload,
	}

	err = h.outboxService.PublishWithTransaction(ctx, tx, event)
	if err != nil {
		span.RecordError(err)
		http.Error(w, "failed to publish event", http.StatusInternalServerError)
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, "failed to commit", http.StatusInternalServerError)
		return
	}

	span.SetAttributes(attribute.String("surplus.id", surplusID))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"surplus_id": surplusID,
		"status":     "posted",
	})
}

type ClaimSurplusRequest struct {
	NGOID string `json:"ngo_id"`
}

func (h *Handler) ClaimSurplus(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "ClaimSurplus")
	defer span.End()

	surplusID := chi.URLParam(r, "id")

	var req ClaimSurplusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Optimistic locking update
	result, err := h.db.ExecContext(ctx, `
		UPDATE surplus 
		SET status = 'claimed', 
		    claimed_by_ngo_id = $1, 
		    claimed_at = NOW(),
		    version = version + 1
		WHERE id = $2 
		  AND status = 'available'
		  AND expiry_time > NOW()
	`, req.NGOID, surplusID)

	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		http.Error(w, "surplus already claimed or expired", http.StatusConflict)
		return
	}

	span.SetAttributes(
		attribute.String("surplus.id", surplusID),
		attribute.String("ngo.id", req.NGOID),
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "claimed",
	})
}

func (h *Handler) GetSurplus(w http.ResponseWriter, r *http.Request) {
	// Implementation omitted for brevity
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetNearbyNGOs(w http.ResponseWriter, r *http.Request) {
	// Implementation omitted for brevity
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) LivenessCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *Handler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		http.Error(w, "database not ready", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("READY"))
}
