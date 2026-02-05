package messaging

import (
	"context"
	"fmt"
	"time"

	"github.com/albnnaardy11/pahlawan-pangan/internal/outbox"
	"github.com/nats-io/nats.go"
	"github.com/segmentio/encoding/json"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("nats-publisher")

// NATSPublisher implements MessagePublisher for NATS JetStream
type NATSPublisher struct {
	js nats.JetStreamContext
}

func NewNATSPublisher(nc *nats.Conn) (*NATSPublisher, error) {
	js, err := nc.JetStream()
	if err != nil {
		return nil, err
	}

	// Create streams if they don't exist
	streams := []string{"SURPLUS", "MATCHING", "NOTIFICATIONS"}
	for _, stream := range streams {
		_, err := js.StreamInfo(stream)
		if err != nil {
			// Create stream
			_, err = js.AddStream(&nats.StreamConfig{
				Name:      stream,
				Subjects:  []string{fmt.Sprintf("%s.*", stream)},
				Storage:   nats.FileStorage,
				Retention: nats.WorkQueuePolicy,
				MaxAge:    24 * time.Hour,
			})
			if err != nil {
				return nil, err
			}
		}
	}

	return &NATSPublisher{js: js}, nil
}

func (p *NATSPublisher) Publish(ctx context.Context, event outbox.OutboxEvent) error {
	ctx, span := tracer.Start(ctx, "NATSPublish")
	defer span.End()

	// Determine subject based on event type
	subject := p.getSubject(event.EventType)

	// Serialize event
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Create NATS message with trace propagation
	msg := nats.NewMsg(subject)
	msg.Data = data

	// Inject trace context into headers
	carrier := &natsHeaderCarrier{header: msg.Header}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	// Publish with acknowledgment
	_, err = p.js.PublishMsg(msg)
	if err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

func (p *NATSPublisher) getSubject(eventType outbox.EventType) string {
	switch eventType {
	case outbox.SurplusPosted:
		return "SURPLUS.posted"
	case outbox.SurplusClaimed:
		return "SURPLUS.claimed"
	case outbox.SurplusExpired:
		return "SURPLUS.expired"
	case outbox.RematchRequired:
		return "MATCHING.rematch"
	default:
		return "SURPLUS.unknown"
	}
}

// natsHeaderCarrier implements propagation.TextMapCarrier
type natsHeaderCarrier struct {
	header nats.Header
}

func (c *natsHeaderCarrier) Get(key string) string {
	return c.header.Get(key)
}

func (c *natsHeaderCarrier) Set(key, value string) {
	c.header.Set(key, value)
}

func (c *natsHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(c.header))
	for k := range c.header {
		keys = append(keys, k)
	}
	return keys
}
