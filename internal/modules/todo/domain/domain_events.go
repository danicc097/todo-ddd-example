package domain

import (
	"time"

	"github.com/google/uuid"
)

type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

type TodoCreatedEvent struct {
	ID        uuid.UUID
	Title     string
	Status    string
	CreatedAt time.Time
	Occurred  time.Time
}

func (e TodoCreatedEvent) EventName() string     { return "todo.created" }
func (e TodoCreatedEvent) OccurredAt() time.Time { return e.Occurred }

type TodoCompletedEvent struct {
	ID        uuid.UUID
	Title     string
	Status    string
	CreatedAt time.Time
	Occurred  time.Time
}

func (e TodoCompletedEvent) EventName() string     { return "todo.completed" }
func (e TodoCompletedEvent) OccurredAt() time.Time { return e.Occurred }

type TagAddedEvent struct {
	TodoID   uuid.UUID
	TagID    uuid.UUID
	Occurred time.Time
}

func (e TagAddedEvent) EventName() string     { return "todo.tag_added" }
func (e TagAddedEvent) OccurredAt() time.Time { return e.Occurred }
