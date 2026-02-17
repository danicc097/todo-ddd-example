package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// DomainEvent is the generic contract all events must satisfy.
type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
	// AggregateID allows us to route events in message brokers.
	AggregateID() uuid.UUID
}

// EventsAggregate defines the contract for an aggregate root that manages events.
type EventsAggregate interface {
	Events() []DomainEvent
	ClearEvents()
}

// EventMapper defines the contract for mapping a domain event to the outbox.
type EventMapper interface {
	MapEvent(e DomainEvent) (string, []byte, error)
}

// EventPublisher is the generic output port.
type EventPublisher interface {
	Publish(ctx context.Context, events ...DomainEvent) error
}
