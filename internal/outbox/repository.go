package outbox

import (
	"context"
	"database/sql"
)

type Repository interface {
	Save(ctx context.Context, tx *sql.Tx, event Event) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Save(ctx context.Context, tx *sql.Tx, event Event) error {
	query := `
		INSERT INTO outbox_events (id, aggregate_id, event_type, payload, created_at, published, trace_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query,
			event.ID, event.AggregateID, event.EventType, event.Payload, event.CreatedAt, event.Published, event.TraceID)
	} else {
		_, err = r.db.ExecContext(ctx, query,
			event.ID, event.AggregateID, event.EventType, event.Payload, event.CreatedAt, event.Published, event.TraceID)
	}
	return err
}
