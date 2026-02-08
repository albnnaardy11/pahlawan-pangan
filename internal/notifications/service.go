package notifications

import (
	"context"
	"fmt"
	"log"
	"time"

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

// NotifyBatch sends notifications to multiple users concurrently (Fan-out)
func (s *NotificationService) NotifyBatch(ctx context.Context, userIDs []string, title, message string) error {
	const batchSize = 100 // Process 100 users per worker task

	// Parallel processing using a semaphore to limit concurrent HTTP requests
	// Unicorn scale: we might hit 10k users. We don't want 10k goroutines instantly.
	sem := make(chan struct{}, 50) // Limit to 50 concurrent request batches
	
	for i := 0; i < len(userIDs); i += batchSize {
		end := i + batchSize
		if end > len(userIDs) {
			end = len(userIDs)
		}
		
		batch := userIDs[i:end]
		
		sem <- struct{}{} // Acquire token
		go func(users []string) {
			defer func() { <-sem }() // Release token
			
			// In production: sending a "multicast" request to FCM/OneSignal is better than loop
			// For simulation: loop
			for _, uid := range users {
				// Fire and forget individual push
				// Use a new context with timeout for each push
				_, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				_ = s.sendPush(uid, fmt.Sprintf("%s: %s", title, message))
				cancel()
			}
		}(batch)
	}

	return nil
}

func (s *NotificationService) sendPush(ngoID string, _ string) error {
	// Simulate transient failure or "offline" status
	if ngoID == "ngo-offline-mock" {
		return fmt.Errorf("device unreachable")
	}
	// Integration point: Firebase Cloud Messaging (FCM) / OneSignal
	// client.Send(msg, ngoID)
	return nil
}

func (s *NotificationService) handleDeliveryFailure(_ context.Context, surplusID, ngoID string, _ error) error {
	// Implementation: Insert into DLQ table or emit RematchRequired event
	// This ensures the food is saved even if the primary recipient is down.
	log.Printf("[DLQ] Surplus %s must be re-routed. Original NGO %s unreachable.", surplusID, ngoID)

	// In production, this would be a DB insert into outbox_events with EventType = RematchRequired
	return nil
}
