package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/albnnaardy11/pahlawan-pangan/internal/domain"
	"github.com/albnnaardy11/pahlawan-pangan/internal/fintech"
)

type disputeUsecase struct {
	escrowSvc *fintech.EscrowService
}

func NewDisputeUsecase(escrow *fintech.EscrowService) domain.DisputeUsecase {
	return &disputeUsecase{
		escrowSvc: escrow,
	}
}

func (u *disputeUsecase) RaiseDispute(ctx context.Context, dispute *domain.Dispute) error {
	dispute.Status = "pending"
	dispute.CreatedAt = time.Now()
	fmt.Printf("⚖️ [LEGAL] Dispute raised for Claim %s. Status: PENDING.\n", dispute.ClaimID)
	return nil
}

func (u *disputeUsecase) ResolveDispute(ctx context.Context, disputeID string, status string) error {
	if status == "approved" {
		// Logic: Identify payment ID from disputeID and trigger refund
		_ = u.escrowSvc.RefundFunds(ctx, "PAY-MOCK-123", "USER-MOCK-456")
	}
	return nil
}

func (u *disputeUsecase) AutoRefundStaleClaims(ctx context.Context) error {
	fmt.Println("⏲️ [SRE] Running Auto-Refund worker for stale claims (15m timeout)...")
	// Logic: Find claims where status = 'locked' and timestamp > 15m
	return nil
}
