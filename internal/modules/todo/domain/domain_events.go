package domain

import (
	"time"

	"github.com/google/uuid"

	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

var (
	_ shared.DomainEvent = (*TodoCreatedEvent)(nil)
	_ shared.DomainEvent = (*TodoCompletedEvent)(nil)
	_ shared.DomainEvent = (*TagAddedEvent)(nil)
	_ shared.DomainEvent = (*TagCreatedEvent)(nil)
	_ shared.DomainEvent = (*TodoRolledOverEvent)(nil)
	_ shared.DomainEvent = (*TodoDeletedEvent)(nil)
)

type TagCreatedEvent struct {
	ID       TagID
	Name     TagName
	WsID     wsDomain.WorkspaceID
	Occurred time.Time
}

func (e TagCreatedEvent) EventName() shared.EventType         { return shared.TodoTagCreated }
func (e TagCreatedEvent) OccurredAt() time.Time               { return e.Occurred }
func (e TagCreatedEvent) AggregateID() uuid.UUID              { return e.ID.UUID() }
func (e TagCreatedEvent) AggregateType() shared.AggregateType { return shared.AggTag }
func (e TagCreatedEvent) WorkspaceID() uuid.UUID              { return e.WsID.UUID() }

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
	ActorID   userDomain.UserID
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

type TodoRolledOverEvent struct {
	ID         TodoID
	WsID       wsDomain.WorkspaceID
	NewDueDate time.Time
	Occurred   time.Time
	ActorID    userDomain.UserID
}

func (e TodoRolledOverEvent) EventName() shared.EventType         { return shared.TodoRolledOver }
func (e TodoRolledOverEvent) OccurredAt() time.Time               { return e.Occurred }
func (e TodoRolledOverEvent) AggregateID() uuid.UUID              { return e.ID.UUID() }
func (e TodoRolledOverEvent) AggregateType() shared.AggregateType { return shared.AggTodo }
func (e TodoRolledOverEvent) WorkspaceID() uuid.UUID              { return e.WsID.UUID() }

type TodoDeletedEvent struct {
	ID       TodoID
	WsID     wsDomain.WorkspaceID
	Occurred time.Time
}

func (e TodoDeletedEvent) EventName() shared.EventType         { return shared.TodoDeleted }
func (e TodoDeletedEvent) OccurredAt() time.Time               { return e.Occurred }
func (e TodoDeletedEvent) AggregateID() uuid.UUID              { return e.ID.UUID() }
func (e TodoDeletedEvent) AggregateType() shared.AggregateType { return shared.AggTodo }
func (e TodoDeletedEvent) WorkspaceID() uuid.UUID              { return e.WsID.UUID() }
