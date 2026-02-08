package repository

import (
	"context"
	"database/sql"

	"github.com/albnnaardy11/pahlawan-pangan/internal/carbon/domain"
)

type carbonRepository struct {
	db *sql.DB
}

func NewCarbonRepository(db *sql.DB) *carbonRepository {
	return &carbonRepository{db: db}
}

func (r *carbonRepository) Save(ctx context.Context, entry domain.CarbonEntry) error {
	query := `
		INSERT INTO carbon_ledger (id, vendor_id, order_id, category, weight_kg, carbon_saved_kg, timestamp, prev_hash, hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.ExecContext(ctx, query,
		entry.ID,
		entry.VendorID,
		entry.OrderID,
		entry.FoodCategory,
		entry.WeightKg,
		entry.CarbonSavedKg,
		entry.Timestamp,
		entry.PreviousHash,
		entry.Hash,
	)
	return err
}

func (r *carbonRepository) GetByVendorPeriod(ctx context.Context, vendorID string, start, end string) ([]domain.CarbonEntry, error) {
	query := `
		SELECT id, vendor_id, order_id, category, weight_kg, carbon_saved_kg, timestamp, prev_hash, hash
		FROM carbon_ledger
		WHERE vendor_id = $1 AND timestamp BETWEEN $2 AND $3
		ORDER BY timestamp ASC
	`
	rows, err := r.db.QueryContext(ctx, query, vendorID, start, end)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var entries []domain.CarbonEntry
	for rows.Next() {
		var e domain.CarbonEntry
		err := rows.Scan(
			&e.ID, &e.VendorID, &e.OrderID, &e.FoodCategory,
			&e.WeightKg, &e.CarbonSavedKg, &e.Timestamp,
			&e.PreviousHash, &e.Hash,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func (r *carbonRepository) GetLastHash(ctx context.Context) (string, error) {
	var hash string
	err := r.db.QueryRowContext(ctx, "SELECT hash FROM carbon_ledger ORDER BY timestamp DESC LIMIT 1").Scan(&hash)
	if err == sql.ErrNoRows {
		return "0000000000000000", nil
	}
	return hash, err
}
