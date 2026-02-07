package postgresql

import (
	"context"
	"database/sql"

	"github.com/albnnaardy11/pahlawan-pangan/internal/domain"
)

type surplusRepository struct {
	masterDB *sql.DB // For Write: INSERT, UPDATE, DELETE
	slaveDB  *sql.DB // For Read: SELECT
}

func NewSurplusRepository(master *sql.DB, slave *sql.DB) domain.SurplusRepository {
	return &surplusRepository{
		masterDB: master,
		slaveDB:  slave,
	}
}

func (r *surplusRepository) GetByID(ctx context.Context, id string) (*domain.SurplusItem, error) {
	var item domain.SurplusItem
	// Use slaveDB for reading
	err := r.slaveDB.QueryRowContext(ctx, "SELECT id, provider_id, quantity_kgs, status FROM surplus WHERE id = $1", id).
		Scan(&item.ID, &item.ProviderID, &item.QuantityKgs, &item.Status)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *surplusRepository) Fetch(ctx context.Context, lat, lon float64, radius int) ([]domain.SurplusItem, error) {
	// Use slaveDB for reading
	return nil, nil
}

func (r *surplusRepository) Store(ctx context.Context, item *domain.SurplusItem) error {
	// Use masterDB for writing
	_, err := r.masterDB.ExecContext(ctx, `
		INSERT INTO surplus (id, provider_id, location, quantity_kgs, food_type, expiry_time, status)
		VALUES ($1, $2, ST_SetSRID(ST_MakePoint($3, $4), 4326), $5, $6, $7, $8)
	`, item.ID, item.ProviderID, item.Longitude, item.Latitude, item.QuantityKgs, item.FoodType, item.ExpiryTime, item.Status)
	return err
}

func (r *surplusRepository) Update(ctx context.Context, item *domain.SurplusItem) error {
	// Use masterDB for writing
	return nil
}
