package messaging

import (
	"context"

	"github.com/google/uuid"
)

// Broker defines a generic network publisher.
type Broker interface {
	Publish(ctx context.Context, eventType string, aggID uuid.UUID, payload []byte, headers map[string]string) error
}

type MultiBroker struct {
	brokers []Broker
}

func NewMultiBroker(b ...Broker) *MultiBroker {
	return &MultiBroker{brokers: b}
}

func (m *MultiBroker) Publish(ctx context.Context, eventType string, aggID uuid.UUID, payload []byte, headers map[string]string) error {
	for _, b := range m.brokers {
		if err := b.Publish(ctx, eventType, aggID, payload, headers); err != nil {
			return err
		}
	}

	return nil
}
