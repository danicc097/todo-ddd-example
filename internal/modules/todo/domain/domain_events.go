package domain

import (
	"time"

	"github.com/google/uuid"

	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

var (
	_ shared.DomainEvent = (*TodoCreatedEvent)(nil)
	_ shared.DomainEvent = (*TodoCompletedEvent)(nil)
	_ shared.DomainEvent = (*TagAddedEvent)(nil)
)

type TodoCreatedEvent struct {
	ID        TodoID
	WsID      wsDomain.WorkspaceID
	Title     TodoTitle
	Status    TodoStatus
	CreatedAt time.Time
	Occurred  time.Time
}

func (e TodoCreatedEvent) EventName() shared.EventType         { return shared.TodoCreated }
func (e TodoCreatedEvent) OccurredAt() time.Time               { return e.Occurred }
func (e TodoCreatedEvent) AggregateID() uuid.UUID              { return e.ID.UUID() }
func (e TodoCreatedEvent) AggregateType() shared.AggregateType { return shared.AggTodo }
func (e TodoCreatedEvent) WorkspaceID() uuid.UUID              { return e.WsID.UUID() }

type TodoCompletedEvent struct {
	ID        TodoID
	WsID      wsDomain.WorkspaceID
	Title     TodoTitle
	Status    TodoStatus
	CreatedAt time.Time
	Occurred  time.Time
}

func (e TodoCompletedEvent) EventName() shared.EventType         { return shared.TodoCompleted }
func (e TodoCompletedEvent) OccurredAt() time.Time               { return e.Occurred }
func (e TodoCompletedEvent) AggregateID() uuid.UUID              { return e.ID.UUID() }
func (e TodoCompletedEvent) AggregateType() shared.AggregateType { return shared.AggTodo }
func (e TodoCompletedEvent) WorkspaceID() uuid.UUID              { return e.WsID.UUID() }

type TagAddedEvent struct {
	TodoID   TodoID
	TagID    TagID
	WsID     wsDomain.WorkspaceID
	Occurred time.Time
}

func (e TagAddedEvent) EventName() shared.EventType         { return shared.TodoTagAdded }
func (e TagAddedEvent) OccurredAt() time.Time               { return e.Occurred }
func (e TagAddedEvent) AggregateID() uuid.UUID              { return e.TodoID.UUID() }
func (e TagAddedEvent) AggregateType() shared.AggregateType { return shared.AggTodo }
func (e TagAddedEvent) WorkspaceID() uuid.UUID              { return e.WsID.UUID() }
