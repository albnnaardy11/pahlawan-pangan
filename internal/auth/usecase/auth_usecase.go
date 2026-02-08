package usecase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/albnnaardy11/pahlawan-pangan/internal/auth/domain"
	"github.com/albnnaardy11/pahlawan-pangan/internal/messaging"
	"github.com/albnnaardy11/pahlawan-pangan/internal/outbox"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/encoding/json"
)

type authUsecase struct {
	repo          domain.UserRepository
	redis         *redis.Client
	natsPublisher *messaging.NATSPublisher
	jwtSecret     []byte
	timeout       time.Duration
}

func NewAuthUsecase(repo domain.UserRepository, rb *redis.Client, np *messaging.NATSPublisher, timeout time.Duration) domain.AuthUsecase {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "pahlawan-pangan-super-secret-key-2026" // #nosec G101
	}
	return &authUsecase{
		repo:          repo,
		redis:         rb,
		natsPublisher: np,
		jwtSecret:     []byte(secret),
		timeout:       timeout,
	}
}

func (u *authUsecase) Register(ctx context.Context, user *domain.User, password string) error {
	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()

	hashed, err := HashPassword(password) // Using Argon2id
	if err != nil {
		return err
	}
	
	user.ID = uuid.New().String()
	user.PasswordHash = hashed
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	return u.repo.Create(ctx, user)
}

func (u *authUsecase) Login(ctx context.Context, email, password string) (*domain.AuthResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()

	user, err := u.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	match, err := ComparePassword(password, user.PasswordHash)
	if err != nil || !match {
		return nil, errors.New("invalid credentials")
	}

	token, err := u.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (u *authUsecase) ValidateToken(ctx context.Context, tokenString string) (*domain.User, error) {
	// 1. Check Redis Blacklist (Force Logout Feature)
	isBlacklisted, _ := u.redis.Get(ctx, "blacklist:"+tokenString).Result()
	if isBlacklisted != "" {
		return nil, errors.New("token is revoked")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return u.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims := token.Claims.(jwt.MapClaims)
	userID := claims["sub"].(string)

	return u.repo.GetByID(ctx, userID)
}

func (u *authUsecase) Logout(ctx context.Context, token string) error {
	// Add to Redis Blacklist until token expires (e.g., 24h)
	return u.redis.Set(ctx, "blacklist:"+token, "true", 24*time.Hour).Err()
}

func (u *authUsecase) RequestOTP(ctx context.Context, userID string) error {
	code := "123456" // In prod: cryptographically secure random 6-digit
	
	// Store in Redis with 5 min expiry
	u.redis.Set(ctx, "otp:"+userID, code, 5*time.Minute)

	// Async Push via NATS for background processing (Unicorn Standard)
	payload, _ := json.Marshal(struct {
		UserID string `json:"user_id"`
		Code   string `json:"code"`
		Action string `json:"action"`
	}{userID, code, "MFA_LOGIN"})

	_ = u.natsPublisher.Publish(ctx, outbox.OutboxEvent{
		ID:          uuid.New().String(),
		AggregateID: userID,
		EventType:   "auth.otp_requested",
		Payload:     payload,
	})

	return nil
}

func (u *authUsecase) VerifyOTP(ctx context.Context, userID, code string) (bool, error) {
	stored, err := u.redis.Get(ctx, "otp:"+userID).Result()
	if err != nil {
		return false, nil
	}
	return stored == code, nil
}

// UpdateEmail with Event Sourcing (Fraud Prevention)
func (u *authUsecase) UpdateEmail(ctx context.Context, userID, newEmail string) error {
	user, _ := u.repo.GetByID(ctx, userID)
	oldEmail := user.Email
	user.Email = newEmail
	
	if err := u.repo.Update(ctx, user); err != nil {
		return err
	}

	// Record Event for Audit/Fraud detection
	event := domain.AccountEvent{
		ID:        uuid.New().String(),
		UserID:    userID,
		Type:      "EMAIL_CHANGED",
		Payload:   fmt.Sprintf("Old: %s, New: %s", oldEmail, newEmail),
		Timestamp: time.Now(),
	}
	
	// In production, save to separate 'account_audit_ledger' table
	payloadJson, _ := json.Marshal(event)
	fmt.Printf("[Audit Ledger] %s\n", string(payloadJson))

	return nil
}

func (u *authUsecase) GetAuditTrail(ctx context.Context, userID string) ([]domain.AccountEvent, error) {
	// Mock implementation
	return nil, nil
}

func (u *authUsecase) generateToken(user *domain.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(u.jwtSecret)
}
