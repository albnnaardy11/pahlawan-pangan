package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/albnnaardy11/pahlawan-pangan/internal/auth/domain"
	"github.com/albnnaardy11/pahlawan-pangan/internal/outbox"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("internal/auth/repository")

type pgUserRepository struct {
	db *sql.DB
	tx *sql.Tx
}

func NewPostgresUserRepository(db *sql.DB) domain.UserRepository {
	return &pgUserRepository{db: db}
}

// Helper to get the correct executor (DB or Tx)
func (r *pgUserRepository) executor() interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
} {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

func (r *pgUserRepository) WithTransaction(ctx context.Context, fn func(domain.UserRepository) error) error {
	ctx, span := tracer.Start(ctx, "db.transaction", trace.WithAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.isolation_level", "read_committed"),
	))
	defer span.End()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	repoTx := &pgUserRepository{
		db: r.db,
		tx: tx,
	}

	if err := fn(repoTx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "transaction rollback")
		_ = tx.Rollback() // Trace rollback?
		return err
	}

	if err := tx.Commit(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "commit failed")
		return err
	}
	
	span.SetStatus(codes.Ok, "transaction committed")
	return nil
}

func (r *pgUserRepository) SaveAudit(ctx context.Context, event *domain.AccountEvent) error {
	ctx, span := tracer.Start(ctx, "db.save_audit")
	defer span.End()

	query := `
		INSERT INTO account_audit_ledger (id, user_id, type, payload, timestamp, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.executor().ExecContext(ctx, query,
		event.ID,
		event.UserID,
		event.Type,
		event.Payload,
		event.Timestamp,
		event.IPAddress,
		event.UserAgent,
	)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	return nil
}

func (r *pgUserRepository) SaveOutbox(ctx context.Context, event *outbox.OutboxEvent) error {
	ctx, span := tracer.Start(ctx, "db.save_outbox", trace.WithAttributes(
		attribute.String("outbox.event_type", string(event.EventType)),
		attribute.String("outbox.aggregate_id", event.AggregateID),
	))
	defer span.End()

	query := `
		INSERT INTO outbox_events (id, aggregate_id, event_type, payload, created_at, published, trace_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	// Ensure defaults
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}

	// Propagate TraceID if not set
	if event.TraceID == "" {
		spanCtx := trace.SpanContextFromContext(ctx)
		if spanCtx.IsValid() {
			event.TraceID = spanCtx.TraceID().String()
		}
	}

	_, err := r.executor().ExecContext(ctx, query,
		event.ID,
		event.AggregateID,
		event.EventType,
		event.Payload,
		event.CreatedAt,
		false,         // published
		event.TraceID, // might be empty
	)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	return nil
}

func (r *pgUserRepository) Create(ctx context.Context, user *domain.User) error {
	ctx, span := tracer.Start(ctx, "db.create_user")
	defer span.End()

	// Dual-Write: Writing to both 'email' (old) and 'contact_email' (new)
	query := `
		INSERT INTO users (id, email, contact_email, phone, password_hash, full_name, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.executor().ExecContext(ctx, query,
		user.ID,
		user.Email, // Old column
		user.Email, // New column (Double write)
		user.Phone,
		user.PasswordHash,
		user.FullName,
		user.Role,
		user.Status,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	return nil
}

func (r *pgUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	ctx, span := tracer.Start(ctx, "db.get_by_email", trace.WithAttributes(
		attribute.String("user.email", email),
	))
	defer span.End()

	// Dual-Read: Select both. Application logic chooses valid one.
	query := `SELECT id, email, contact_email, phone, password_hash, full_name, role, status, created_at, updated_at FROM users WHERE email = $1 OR contact_email = $1`
	user := &domain.User{}
	
	var emailCol, contactEmailCol sql.NullString

	err := r.executor().QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&emailCol,
		&contactEmailCol,
		&user.Phone,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err != sql.ErrNoRows {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return nil, err
	}

	// Smart Switch: Prefer contact_email (new), fallback to email (old)
	if contactEmailCol.Valid && contactEmailCol.String != "" {
		user.Email = contactEmailCol.String
	} else {
		user.Email = emailCol.String
	}

	return user, nil
}

func (r *pgUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	ctx, span := tracer.Start(ctx, "db.get_by_id", trace.WithAttributes(
		attribute.String("user.id", id),
	))
	defer span.End()

	query := `SELECT id, email, contact_email, phone, password_hash, full_name, role, status, created_at, updated_at FROM users WHERE id = $1`
	user := &domain.User{}
	var emailCol, contactEmailCol sql.NullString

	err := r.executor().QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&emailCol,
		&contactEmailCol,
		&user.Phone,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err != sql.ErrNoRows {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return nil, err
	}

	if contactEmailCol.Valid && contactEmailCol.String != "" {
		user.Email = contactEmailCol.String
	} else {
		user.Email = emailCol.String
	}

	return user, nil
}

func (r *pgUserRepository) GetByIDForUpdate(ctx context.Context, id string) (*domain.User, error) {
	ctx, span := tracer.Start(ctx, "db.get_by_id_for_update", trace.WithAttributes(
		attribute.String("user.id", id),
		attribute.String("db.lock_type", "PESSIMISTIC_WRITE"), // FOR UPDATE
	))
	defer span.End()

	// Uses FOR UPDATE to lock the row during transaction
	query := `SELECT id, email, contact_email, phone, password_hash, full_name, role, status, created_at, updated_at FROM users WHERE id = $1 FOR UPDATE`
	// Attribute the statement for clarity in traces
	span.SetAttributes(attribute.String("db.statement", query))
	
	user := &domain.User{}
	var emailCol, contactEmailCol sql.NullString

	err := r.executor().QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&emailCol,
		&contactEmailCol,
		&user.Phone,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if contactEmailCol.Valid && contactEmailCol.String != "" {
		user.Email = contactEmailCol.String
	} else {
		user.Email = emailCol.String
	}

	return user, nil
}

func (r *pgUserRepository) Update(ctx context.Context, user *domain.User) error {
	ctx, span := tracer.Start(ctx, "db.update_user", trace.WithAttributes(
		attribute.String("user.id", user.ID),
	))
	defer span.End()

	// Dual-Write Update
	query := `
		UPDATE users 
		SET full_name = $1, status = $2, updated_at = $3, email = $4, contact_email = $5
		WHERE id = $6
	`
	user.UpdatedAt = time.Now()
	_, err := r.executor().ExecContext(ctx, query,
		user.FullName,
		user.Status,
		user.UpdatedAt,
		user.Email, // Update old
		user.Email, // Update new (Dual Write)
		user.ID,
	)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	return nil
}
