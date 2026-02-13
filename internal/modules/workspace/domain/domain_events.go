package domain

import (
	"time"

	user "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
	"github.com/google/uuid"
)

var (
	_ shared.DomainEvent = (*WorkspaceCreatedEvent)(nil)
	_ shared.DomainEvent = (*MemberAddedEvent)(nil)
	_ shared.DomainEvent = (*MemberRemovedEvent)(nil)
)

type WorkspaceCreatedEvent struct {
	ID       WorkspaceID
	Name     string
	OwnerID  user.UserID
	Occurred time.Time
}

func (e WorkspaceCreatedEvent) EventName() string      { return "workspace.created" }
func (e WorkspaceCreatedEvent) OccurredAt() time.Time  { return e.Occurred }
func (e WorkspaceCreatedEvent) AggregateID() uuid.UUID { return e.ID.UUID }

type MemberAddedEvent struct {
	WorkspaceID WorkspaceID
	UserID      user.UserID
	Role        string
	Occurred    time.Time
}

func (e MemberAddedEvent) EventName() string      { return "workspace.member_added" }
func (e MemberAddedEvent) OccurredAt() time.Time  { return e.Occurred }
func (e MemberAddedEvent) AggregateID() uuid.UUID { return e.WorkspaceID.UUID }

type MemberRemovedEvent struct {
	WorkspaceID WorkspaceID
	UserID      user.UserID
	Occurred    time.Time
}

func (e MemberRemovedEvent) EventName() string      { return "workspace.member_removed" }
func (e MemberRemovedEvent) OccurredAt() time.Time  { return e.Occurred }
func (e MemberRemovedEvent) AggregateID() uuid.UUID { return e.WorkspaceID.UUID }
