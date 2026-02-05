package notifications

import (
	"context"
	"fmt"
	"log"

	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("notification-service")

// NotificationService handles sending alerts to NGOs
type NotificationService struct {
	// e.g., Firebase Cloud Messaging or OneSignal client
}

// Dispatch sends a notification and handles delivery failure
func (s *NotificationService) Dispatch(ctx context.Context, surplusID, ngoID string) error {
	ctx, span := tracer.Start(ctx, "Dispatch")
	defer span.End()

	// Logic to send push notification
	err := s.sendPush(ngoID, fmt.Sprintf("New surplus available: %s", surplusID))
	
	if err != nil {
		// Resilience: Handle "NGO Offline" or Delivery Failure
		// Instead of just failing, we trigger the DLQ/Rematch flow
		log.Printf("[Notification] Failed to reach NGO %s. Triggering DLQ strategy.", ngoID)
		return s.handleDeliveryFailure(ctx, surplusID, ngoID, err)
	}

	return nil
}

func (s *NotificationService) sendPush(ngoID, msg string) error {
	// Simulate transient failure or "offline" status
	if ngoID == "ngo-offline-mock" {
		return fmt.Errorf("device unreachable")
	}
	return nil
}

func (s *NotificationService) handleDeliveryFailure(ctx context.Context, surplusID, ngoID string, originalErr error) error {
	// Implementation: Insert into DLQ table or emit RematchRequired event
	// This ensures the food is saved even if the primary recipient is down.
	log.Printf("[DLQ] Surplus %s must be re-routed. Original NGO %s unreachable.", surplusID, ngoID)
	
	// In production, this would be a DB insert into outbox_events with EventType = RematchRequired
	return nil
}
