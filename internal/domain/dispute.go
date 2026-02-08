// Package domain contains core business entities and interfaces for the dispute resolution system.
package domain

import (
	"context"
	"time"
)

// Dispute represents a user-raised dispute for a claim transaction.
type Dispute struct {
	ID        string    `json:"id"`
	ClaimID   string    `json:"claim_id" validate:"required"`
	UserID    string    `json:"user_id" validate:"required"`
	Reason    string    `json:"reason" validate:"required"`
	Evidence  string    `json:"evidence_url"`
	Status    string    `json:"status"` // pending, approved, rejected
	CreatedAt time.Time `json:"created_at"`
}

// DisputeUsecase defines the business logic interface for dispute management.
type DisputeUsecase interface {
	RaiseDispute(ctx context.Context, dispute *Dispute) error
	ResolveDispute(ctx context.Context, disputeID string, status string) error
	AutoRefundStaleClaims(ctx context.Context) error
}
