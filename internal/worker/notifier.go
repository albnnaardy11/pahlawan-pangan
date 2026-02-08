package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/albnnaardy11/pahlawan-pangan/internal/geo"
	"github.com/albnnaardy11/pahlawan-pangan/internal/notifications"
	"github.com/albnnaardy11/pahlawan-pangan/internal/outbox"
)

type SurplusNotifier struct {
	geoSvc   *geo.GeoService
	notifSvc *notifications.NotificationService
	logger   *zap.Logger
}

func NewSurplusNotifier(geoSvc *geo.GeoService, notifSvc *notifications.NotificationService) *SurplusNotifier {
	return &SurplusNotifier{
		geoSvc:   geoSvc,
		notifSvc: notifSvc,
		logger:   zap.NewExample(),
	}
}

// Start consuming events from NATS
func (n *SurplusNotifier) Run(ctx context.Context, nc *nats.Conn) {
	_, err := nc.Subscribe("surplus.posted", func(msg *nats.Msg) {
		// Process message
		var event outbox.Event
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			n.logger.Error("Failed to unmarshal event", zap.Error(err))
			return
		}

		// Handle SurplusPosted events
		if event.EventType == outbox.SurplusPosted {
			n.handleSurplusPosted(ctx, event.Payload)
		}
	})

	if err != nil {
		n.logger.Error("Failed to subscribe to NATS", zap.Error(err))
	} else {
		n.logger.Info("Started SurplusNotifier worker")
	}

	<-ctx.Done()
}

func (n *SurplusNotifier) handleSurplusPosted(ctx context.Context, payload json.RawMessage) {
	// 1. Parse payload to get Surplus Location
	var data struct {
		SurplusID   string  `json:"surplus_id"`
		Lat         float64 `json:"lat"`
		Lon         float64 `json:"lon"`
		QuantityKgs float64 `json:"quantity_kgs"`
	}
	if err := json.Unmarshal(payload, &data); err != nil {
		n.logger.Error("Invalid payload", zap.Error(err))
		return
	}

	// 2. Find Users within 500m (Real-time Geo Query)
	// This is the "Unicorn Logic" - querying millions of users in ms
	userIDs, err := n.geoSvc.FindUsersNearby(ctx, data.Lat, data.Lon, 500)
	if err != nil {
		n.logger.Error("Geo query failed", zap.Error(err))
		return
	}

	if len(userIDs) == 0 {
		n.logger.Info("No users found nearby", zap.String("surplus_id", data.SurplusID))
		return
	}

	n.logger.Info("Found nearby users",
		zap.Int("count", len(userIDs)),
		zap.String("surplus_id", data.SurplusID))

	// 3. Batch Notify (Fan-out)
	title := "Free Food Nearby! 🍱"
	message := fmt.Sprintf("%.1f kg available within 500m of you!", data.QuantityKgs)

	err = n.notifSvc.NotifyBatch(ctx, userIDs, title, message)
	if err != nil {
		n.logger.Error("Failed to send notifications", zap.Error(err))
	}
}
