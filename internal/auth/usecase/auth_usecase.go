package usecase

import (
	"context"
	"crypto/rsa"
	"crypto/subtle"
	"errors"
	"fmt"
	"os"
	"time"
	"unsafe"

	"github.com/albnnaardy11/pahlawan-pangan/internal/auth/domain"
	"github.com/albnnaardy11/pahlawan-pangan/internal/messaging"
	"github.com/albnnaardy11/pahlawan-pangan/internal/outbox"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/encoding/json"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var tracer = otel.Tracer("internal/auth/usecase")

// Zero-allocation string to byte slice conversion
// WARNING: The returned slice must not be modified.
func stringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

type authUsecase struct {
	repo          domain.UserRepository
	redis         *redis.Client
	natsPublisher *messaging.NATSPublisher
	privateKey    *rsa.PrivateKey
	publicKey     *rsa.PublicKey
	timeout       time.Duration
}

func NewAuthUsecase(repo domain.UserRepository, rb *redis.Client, np *messaging.NATSPublisher, timeout time.Duration) domain.AuthUsecase {
	// Load RSA Keys for RS256 (Asymmetric)
	privKeyData, err := os.ReadFile("private.pem")
	if err != nil {
		panic("Failed to load private.pem for RS256")
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privKeyData)
	if err != nil {
		panic("Invalid private key")
	}

	pubKeyData, err := os.ReadFile("public.pem")
	if err != nil {
		panic("Failed to load public.pem")
	}
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(pubKeyData)
	if err != nil {
		panic("Invalid public key")
	}

	return &authUsecase{
		repo:          repo,
		redis:         rb,
		natsPublisher: np, // Kept for backward compatibility, but we prefer Outbox Pattern
		privateKey:    privateKey,
		publicKey:     publicKey,
		timeout:       timeout,
	}
}

func (u *authUsecase) Register(ctx context.Context, user *domain.User, password string) error {
	ctx, span := tracer.Start(ctx, "usecase.register")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()

	hashed, err := HashPassword(password) // Using Argon2id
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "hashing failed")
		return err
	}
	
	user.ID = uuid.New().String()
	user.PasswordHash = hashed
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	if err := u.repo.Create(ctx, user); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "db create failed")
		return err
	}

	return nil
}

func (u *authUsecase) Login(ctx context.Context, email, password string) (*domain.AuthResponse, error) {
	ctx, span := tracer.Start(ctx, "usecase.login", 
		// trace.WithAttributes(attribute.String("email", email)), // PII caution
	)
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()

	user, err := u.repo.GetByEmail(ctx, email)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "user not found or db error")
		return nil, errors.New("invalid credentials")
	}

	match, err := ComparePassword(password, user.PasswordHash)
	if err != nil || !match {
		span.RecordError(errors.New("password mismatch"))
		span.SetStatus(codes.Error, "invalid credentials")
		return nil, errors.New("invalid credentials")
	}

	token, err := u.generateToken(user)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return &domain.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (u *authUsecase) Logout(ctx context.Context, tokenString string) error {
	ctx, span := tracer.Start(ctx, "usecase.logout")
	defer span.End()

	// 1. Parse Token to get Expiry (Prevent Redis bloat)
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return u.publicKey, nil
	})

	if err != nil {
		// Validating token validity before blacklisting is safer
		span.RecordError(err)
		return err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("invalid token claims")
	}

	expFloat, ok := claims["exp"].(float64)
	if !ok {
		return errors.New("invalid exp claim")
	}

	expTime := time.Unix(int64(expFloat), 0)
	ttl := time.Until(expTime)

	if ttl <= 0 {
		return nil // Already expired
	}

	// 2. Set Redis Expiry == Token Remaining TTL
	if err := u.redis.Set(ctx, "blacklist:"+tokenString, "true", ttl).Err(); err != nil {
		span.RecordError(err)
		return err
	}
	return nil
}

func (u *authUsecase) RequestOTP(ctx context.Context, userID string) error {
	ctx, span := tracer.Start(ctx, "usecase.request_otp", 
		// trace.WithAttributes(attribute.String("user_id", userID)),
	)
	defer span.End()
	span.SetAttributes(attribute.String("user.id", userID))

	// Secure Random 6-digit OTP
	code := "123456" // Placeholder: In prod use crypto/rand

	// Store in Redis with 5 min expiry
	// Note: We use a pipeline for slight perf bump if we were doing multiple ops, 
	// but here single RTT is fine.
	if err := u.redis.Set(ctx, "otp:"+userID, code, 5*time.Minute).Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "redis set failed")
		return err
	}

	// Outbox Pattern: Save event to SQL DB instead of direct NATS publish
	// This ensures Atomicity (if we were inside a transaction) or at least Persistence.
	// A background worker will read 'outbox_events' and push to NATS.
	payload, _ := json.Marshal(struct {
		UserID string `json:"user_id"`
		Code   string `json:"code"`
		Action string `json:"action"`
	}{userID, code, "MFA_LOGIN"})

	event := outbox.OutboxEvent{
		ID:          uuid.New().String(),
		AggregateID: userID,
		EventType:   "auth.otp_requested",
		Payload:     payload, // json.RawMessage
		CreatedAt:   time.Now(),
	}

	if err := u.repo.SaveOutbox(ctx, &event); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "save outbox failed")
		return err
	}

	return nil
}

func (u *authUsecase) VerifyOTP(ctx context.Context, userID, code string) (bool, error) {
	ctx, span := tracer.Start(ctx, "usecase.verify_otp")
	defer span.End()
	span.SetAttributes(attribute.String("user.id", userID))

	stored, err := u.redis.Get(ctx, "otp:"+userID).Result()
	if err != nil {
		// Could be key not found (expired/invalid) covers most cases
		return false, nil
	}

	// 1. Prevent Brute Force: Delete OTP after one use
	u.redis.Del(ctx, "otp:"+userID)

	// 2. Constant Time Compare (Timing Attack Protection)
	// Zero-Allocation conversion
	if len(stored) != len(code) {
		return false, nil
	}
	
	match := subtle.ConstantTimeCompare(stringToBytes(stored), stringToBytes(code)) == 1
	if !match {
		span.SetAttributes(attribute.Bool("otp.match", false))
	}
	return match, nil
}

// UpdateEmail with Transactional Sourcing (Atomic & Locked)
func (u *authUsecase) UpdateEmail(ctx context.Context, userID, newEmail string) error {
	ctx, span := tracer.Start(ctx, "usecase.update_email")
	defer span.End()
	span.SetAttributes(attribute.String("user.id", userID))

	return u.repo.WithTransaction(ctx, func(txRepo domain.UserRepository) error {
		// CRITICAL: Use GetByIDForUpdate (SELECT ... FOR UPDATE)
		// This locks the user row, preventing race conditions if two updates happen simultaneously.
		user, err := txRepo.GetByIDForUpdate(ctx, userID)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "get user for update failed")
			return err
		}

		oldEmail := user.Email
		
		// 1. Save Audit Log (Atomic with update)
		auditEvent := &domain.AccountEvent{
			ID:        uuid.New().String(),
			UserID:    userID,
			Type:      "EMAIL_CHANGED",
			Payload:   fmt.Sprintf("Old: %s, New: %s", oldEmail, newEmail),
			Timestamp: time.Now(),
		}
		
		if err := txRepo.SaveAudit(ctx, auditEvent); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "save audit failed")
			return err
		}

		// 2. Update User
		user.Email = newEmail
		if err := txRepo.Update(ctx, user); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "update user failed")
			return err
		}
		return nil
	})
}

func (u *authUsecase) ValidateToken(ctx context.Context, tokenString string) (*domain.User, error) {
	ctx, span := tracer.Start(ctx, "usecase.validate_token")
	defer span.End()

	// 1. Check Redis Blacklist (Force Logout Feature)
	isBlacklisted, _ := u.redis.Get(ctx, "blacklist:"+tokenString).Result()
	if isBlacklisted != "" {
		span.RecordError(errors.New("token revoked"))
		return nil, errors.New("token is revoked")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Strict Algorithm Check: RS256 ONLY
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return u.publicKey, nil
	})

	if err != nil || !token.Valid {
		span.RecordError(err)
		return nil, errors.New("invalid token")
	}

	claims := token.Claims.(jwt.MapClaims)
	userID := claims["sub"].(string)
	span.SetAttributes(attribute.String("user.id", userID))

	user, err := u.repo.GetByID(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "get user by ID failed")
		return nil, err
	}
	return user, nil
}

// ...

func (u *authUsecase) generateToken(user *domain.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	}
	// Sign with RS256 (Private Key)
	return jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(u.privateKey)
}

func (u *authUsecase) GetAuditTrail(ctx context.Context, userID string) ([]domain.AccountEvent, error) {
	// Mock implementation
	return nil, nil
}
