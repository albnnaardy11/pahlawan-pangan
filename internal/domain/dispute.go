package domain

import (
	"context"
	"time"
)

type Dispute struct {
	ID        string    `json:"id"`
	ClaimID   string    `json:"claim_id" validate:"required"`
	UserID    string    `json:"user_id" validate:"required"`
	Reason    string    `json:"reason" validate:"required"`
	Evidence  string    `json:"evidence_url"`
	Status    string    `json:"status"` // pending, approved, rejected
	CreatedAt time.Time `json:"created_at"`
}

type DisputeUsecase interface {
	RaiseDispute(ctx context.Context, dispute *Dispute) error
	ResolveDispute(ctx context.Context, disputeID string, status string) error
	AutoRefundStaleClaims(ctx context.Context) error
}
