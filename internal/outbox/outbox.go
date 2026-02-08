package outbox

import (
	"context"
	"database/sql"
	"time"

	"github.com/segmentio/encoding/json"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("outbox-service")

// EventType represents the type of domain event
type EventType string

const (
	SurplusPosted   EventType = "surplus.posted"
	SurplusClaimed  EventType = "surplus.claimed"
	SurplusExpired  EventType = "surplus.expired"
	RematchRequired EventType = "surplus.rematch_required"
	FoodDelivered   EventType = "delivery.completed"
	FundsReleased   EventType = "escrow.funds_released"
)

// Event represents an event to be published
type Event struct {
	ID          string          `json:"id"`
	AggregateID string          `json:"aggregate_id"` // surplus_id or ngo_id
	EventType   EventType       `json:"event_type"`
	Payload     json.RawMessage `json:"payload"`
	CreatedAt   time.Time       `json:"created_at"`
	Published   bool            `json:"published"`
	PublishedAt *time.Time      `json:"published_at,omitempty"`
	TraceID     string          `json:"trace_id"`
}

// Service handles transactional outbox pattern
type Service struct {
	db *sql.DB
}

// NewOutboxService creates a new instance of the outbox service
func NewOutboxService(db *sql.DB) *Service {
	return &Service{db: db}
}

// PublishWithTransaction atomically writes to DB and outbox
func (s *Service) PublishWithTransaction(ctx context.Context, tx *sql.Tx, event Event) error {
	ctx, span := tracer.Start(ctx, "PublishWithTransaction")
	defer span.End()

	// Extract trace context
	spanCtx := trace.SpanContextFromContext(ctx)
	event.TraceID = spanCtx.TraceID().String()
	event.CreatedAt = time.Now()

	// Insert into outbox table within the same transaction
	query := `
		INSERT INTO outbox_events (id, aggregate_id, event_type, payload, created_at, published, trace_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := tx.ExecContext(ctx, query,
		event.ID,
		event.AggregateID,
		event.EventType,
		event.Payload,
		event.CreatedAt,
		false,
		event.TraceID,
	)

	if err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

// PollAndPublish reads unpublished events and sends to message broker
func (s *Service) PollAndPublish(ctx context.Context, publisher Publisher, batchSize int) error {
	ctx, span := tracer.Start(ctx, "PollAndPublish")
	defer span.End()

	// Select unpublished events with FOR UPDATE SKIP LOCKED for concurrency
	query := `
		SELECT id, aggregate_id, event_type, payload, created_at, trace_id
		FROM outbox_events
		WHERE published = false
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`

	rows, err := s.db.QueryContext(ctx, query, batchSize)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			span.RecordError(closeErr)
		}
	}()

	// Pre-allocate slice to avoid resizing overhead (Mechanical Sympathy)
	events := make([]Event, 0, batchSize)
	for rows.Next() {
		var event Event
		err := rows.Scan(
			&event.ID,
			&event.AggregateID,
			&event.EventType,
			&event.Payload,
			&event.CreatedAt,
			&event.TraceID,
		)
		if err != nil {
			return err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// Publish each event
	for _, event := range events {
		// 1. Reliability Hook: "The Stale Event Dropper" (Chaos Monkey Resistance)
		// If an event is older than 5 minutes, it's irrelevant (e.g., OTPs). Drop it.
		if time.Since(event.CreatedAt) > 5*time.Minute {
			span.AddEvent("dropping_stale_event", trace.WithAttributes(
				attribute.String("event.id", event.ID),
				attribute.String("event.type", string(event.EventType)),
				attribute.String("event.age", time.Since(event.CreatedAt).String()),
			))

			// Mark as "published" (effectively ignored) to prevent reprocessing loop
			_, err := s.db.ExecContext(ctx, `
				UPDATE outbox_events 
				SET published = true, published_at = $1 
				WHERE id = $2
			`, time.Now(), event.ID)

			if err != nil {
				span.RecordError(err)
			}
			continue
		}

		// Restore trace context
		err := publisher.Publish(ctx, event)
		if err != nil {
			// Log error but continue with other events
			span.RecordError(err)
			continue
		}

		// Mark as published
		_, err = s.db.ExecContext(ctx, `
			UPDATE outbox_events 
			SET published = true, published_at = $1 
			WHERE id = $2
		`, time.Now(), event.ID)

		if err != nil {
			span.RecordError(err)
		}
	}

	return nil
}

// Publisher interface for NATS/Kafka abstraction
type Publisher interface {
	Publish(ctx context.Context, event Event) error
}
