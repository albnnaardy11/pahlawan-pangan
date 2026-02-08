package domain

import (
	"context"
	"time"
)

type UserRole string

const (
	RoleVendor  UserRole = "VENDOR"
	RoleNGO     UserRole = "NGO"
	RoleCourier UserRole = "COURIER"
	RoleAdmin   UserRole = "ADMIN"
)

type User struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	Phone          string    `json:"phone"`
	PasswordHash   string    `json:"-"`
	FullName       string    `json:"full_name"`
	Role           UserRole  `json:"role"`
	Status         string    `json:"status"` // ACTIVE, SUSPENDED, PENDING_MFA
	IsMFAEnabled   bool      `json:"is_mfa_enabled"`
	LastLogin      time.Time `json:"last_login"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// AccountEvent represents an immutable fact in user history (Audit Trail/Fraud Detection)
type AccountEvent struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Type      string    `json:"type"` // e.g., "EmailChanged", "PhoneChanged", "LoginAttempted"
	Payload   string    `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
}

type AuthResponse struct {
	Token string `json:"access_token"`
	User  User   `json:"user"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	Update(ctx context.Context, user *User) error
}

type AuthUsecase interface {
	Register(ctx context.Context, user *User, password string) error
	Login(ctx context.Context, email, password string) (*AuthResponse, error)
	ValidateToken(ctx context.Context, token string) (*User, error)
	Logout(ctx context.Context, token string) error
	
	// MFA/OTP
	RequestOTP(ctx context.Context, userID string) error
	VerifyOTP(ctx context.Context, userID, code string) (bool, error)
	
	// Account Settings (Event Sourced)
	UpdateEmail(ctx context.Context, userID, newEmail string) error
	GetAuditTrail(ctx context.Context, userID string) ([]AccountEvent, error)
}
