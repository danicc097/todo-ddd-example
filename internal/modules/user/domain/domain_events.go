package domain

import (
	"time"

	"github.com/google/uuid"

	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

type UserCreatedEvent struct {
	ID       UserID
	Email    UserEmail
	Name     UserName
	Occurred time.Time
}

func (e UserCreatedEvent) EventName() shared.EventType         { return shared.UserCreated }
func (e UserCreatedEvent) OccurredAt() time.Time               { return e.Occurred }
func (e UserCreatedEvent) AggregateID() uuid.UUID              { return e.ID.UUID() }
func (e UserCreatedEvent) AggregateType() shared.AggregateType { return shared.AggUser }

type UserDeletedEvent struct {
	ID       UserID
	Occurred time.Time
}

func (e UserDeletedEvent) EventName() shared.EventType         { return shared.UserDeleted }
func (e UserDeletedEvent) OccurredAt() time.Time               { return e.Occurred }
func (e UserDeletedEvent) AggregateID() uuid.UUID              { return e.ID.UUID() }
func (e UserDeletedEvent) AggregateType() shared.AggregateType { return shared.AggUser }
