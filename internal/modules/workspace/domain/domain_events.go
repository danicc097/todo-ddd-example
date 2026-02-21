package domain

import (
	"time"

	"github.com/google/uuid"

	user "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

var (
	_ shared.DomainEvent = (*WorkspaceCreatedEvent)(nil)
	_ shared.DomainEvent = (*WorkspaceDeletedEvent)(nil)
	_ shared.DomainEvent = (*MemberAddedEvent)(nil)
	_ shared.DomainEvent = (*MemberRemovedEvent)(nil)
)

type WorkspaceCreatedEvent struct {
	ID       WorkspaceID
	Name     string
	OwnerID  user.UserID
	Occurred time.Time
}

func (e WorkspaceCreatedEvent) EventName() shared.EventType {
	return shared.WorkspaceCreated
}
func (e WorkspaceCreatedEvent) OccurredAt() time.Time               { return e.Occurred }
func (e WorkspaceCreatedEvent) AggregateID() uuid.UUID              { return e.ID.UUID() }
func (e WorkspaceCreatedEvent) AggregateType() shared.AggregateType { return shared.AggWorkspace }

type WorkspaceDeletedEvent struct {
	ID       WorkspaceID
	Occurred time.Time
}

func (e WorkspaceDeletedEvent) EventName() shared.EventType {
	return shared.WorkspaceDeleted
}
func (e WorkspaceDeletedEvent) OccurredAt() time.Time               { return e.Occurred }
func (e WorkspaceDeletedEvent) AggregateID() uuid.UUID              { return e.ID.UUID() }
func (e WorkspaceDeletedEvent) AggregateType() shared.AggregateType { return shared.AggWorkspace }

type MemberAddedEvent struct {
	WorkspaceID WorkspaceID
	UserID      user.UserID
	Role        string
	Occurred    time.Time
}

func (e MemberAddedEvent) EventName() shared.EventType {
	return shared.WorkspaceMemberAdded
}
func (e MemberAddedEvent) OccurredAt() time.Time               { return e.Occurred }
func (e MemberAddedEvent) AggregateID() uuid.UUID              { return e.WorkspaceID.UUID() }
func (e MemberAddedEvent) AggregateType() shared.AggregateType { return shared.AggWorkspace }

type MemberRemovedEvent struct {
	WorkspaceID WorkspaceID
	UserID      user.UserID
	Occurred    time.Time
}

func (e MemberRemovedEvent) EventName() shared.EventType {
	return shared.WorkspaceMemberRemoved
}
func (e MemberRemovedEvent) OccurredAt() time.Time               { return e.Occurred }
func (e MemberRemovedEvent) AggregateID() uuid.UUID              { return e.WorkspaceID.UUID() }
func (e MemberRemovedEvent) AggregateType() shared.AggregateType { return shared.AggWorkspace }
