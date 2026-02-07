package outbox

import (
	"context"
	"database/sql"
	"time"

	"github.com/segmentio/encoding/json"
	"go.opentelemetry.io/otel"
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
)

// OutboxEvent represents an event to be published
type OutboxEvent struct {
	ID          string          `json:"id"`
	AggregateID string          `json:"aggregate_id"` // surplus_id or ngo_id
	EventType   EventType       `json:"event_type"`
	Payload     json.RawMessage `json:"payload"`
	CreatedAt   time.Time       `json:"created_at"`
	Published   bool            `json:"published"`
	PublishedAt *time.Time      `json:"published_at,omitempty"`
	TraceID     string          `json:"trace_id"`
}

// OutboxService handles transactional outbox pattern
type OutboxService struct {
	db *sql.DB
}

func NewOutboxService(db *sql.DB) *OutboxService {
	return &OutboxService{db: db}
}

// PublishWithTransaction atomically writes to DB and outbox
func (s *OutboxService) PublishWithTransaction(ctx context.Context, tx *sql.Tx, event OutboxEvent) error {
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
func (s *OutboxService) PollAndPublish(ctx context.Context, publisher MessagePublisher, batchSize int) error {
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
	defer rows.Close()

	var events []OutboxEvent
	for rows.Next() {
		var event OutboxEvent
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

// MessagePublisher interface for NATS/Kafka abstraction
type MessagePublisher interface {
	Publish(ctx context.Context, event OutboxEvent) error
}
