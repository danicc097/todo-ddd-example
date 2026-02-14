package domain

import (
	"time"

	"github.com/google/uuid"

	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

var (
	_ shared.DomainEvent = (*TodoCreatedEvent)(nil)
	_ shared.DomainEvent = (*TodoCompletedEvent)(nil)
	_ shared.DomainEvent = (*TagAddedEvent)(nil)
)

type TodoCreatedEvent struct {
	ID        TodoID
	Title     string
	Status    string
	CreatedAt time.Time
	Occurred  time.Time
}

func (e TodoCreatedEvent) EventName() string      { return "todo.created" }
func (e TodoCreatedEvent) OccurredAt() time.Time  { return e.Occurred }
func (e TodoCreatedEvent) AggregateID() uuid.UUID { return e.ID.UUID }

type TodoCompletedEvent struct {
	ID        TodoID
	Title     string
	Status    string
	CreatedAt time.Time
	Occurred  time.Time
}

func (e TodoCompletedEvent) EventName() string      { return "todo.completed" }
func (e TodoCompletedEvent) OccurredAt() time.Time  { return e.Occurred }
func (e TodoCompletedEvent) AggregateID() uuid.UUID { return e.ID.UUID }

type TagAddedEvent struct {
	TodoID   TodoID
	TagID    TagID
	Occurred time.Time
}

func (e TagAddedEvent) EventName() string      { return "todo.tag_added" }
func (e TagAddedEvent) OccurredAt() time.Time  { return e.Occurred }
func (e TagAddedEvent) AggregateID() uuid.UUID { return e.TodoID.UUID }
