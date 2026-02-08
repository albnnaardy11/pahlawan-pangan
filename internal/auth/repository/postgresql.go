package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/albnnaardy11/pahlawan-pangan/internal/auth/domain"
)

type pgUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) domain.UserRepository {
	return &pgUserRepository{db: db}
}

func (r *pgUserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, phone, password_hash, full_name, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Phone,
		user.PasswordHash,
		user.FullName,
		user.Role,
		user.Status,
		user.CreatedAt,
		user.UpdatedAt,
	)
	return err
}

func (r *pgUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, email, phone, password_hash, full_name, role, status, created_at, updated_at FROM users WHERE email = $1`
	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *pgUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `SELECT id, email, phone, password_hash, full_name, role, status, created_at, updated_at FROM users WHERE id = $1`
	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *pgUserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users 
		SET full_name = $1, status = $2, updated_at = $3 
		WHERE id = $4
	`
	user.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, query,
		user.FullName,
		user.Status,
		user.UpdatedAt,
		user.ID,
	)
	return err
}
