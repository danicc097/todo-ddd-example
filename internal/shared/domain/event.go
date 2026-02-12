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

// EventPublisher is the generic output port.
//
//counterfeiter:generate . EventPublisher
type EventPublisher interface {
	Publish(ctx context.Context, events ...DomainEvent) error
}
