package messaging

import (
	"context"

	"github.com/google/uuid"

	sharedDomain "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

type PublishArgs struct {
	EventType sharedDomain.EventType
	AggID     uuid.UUID
	Payload   []byte
	Headers   map[Header]string
}

// Broker defines a generic network publisher.
type Broker interface {
	Publish(ctx context.Context, args PublishArgs) error
}

type BrokerPublishFunc func(ctx context.Context, args PublishArgs) error

func (f BrokerPublishFunc) Publish(ctx context.Context, args PublishArgs) error {
	return f(ctx, args)
}

type MultiBroker struct {
	brokers []Broker
}

func NewMultiBroker(b ...Broker) *MultiBroker {
	return &MultiBroker{brokers: b}
}

// Publish delegates to all brokers. It does not implement the Broker interface.
func (m *MultiBroker) Publish(ctx context.Context, args PublishArgs) error {
	for _, b := range m.brokers {
		if err := b.Publish(ctx, args); err != nil {
			return err
		}
	}

	return nil
}
