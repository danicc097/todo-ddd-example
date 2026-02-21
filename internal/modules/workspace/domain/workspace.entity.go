package domain

import (
	"time"

	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

var (
	ErrAtLeastOneOwner   = shared.NewDomainError(apperrors.InvalidInput, "workspace must have at least one owner")
	ErrUserAlreadyMember = shared.NewDomainError(apperrors.InvalidInput, "user is already a member")
	ErrMemberNotFound    = shared.NewDomainError(apperrors.NotFound, "member not found")
	ErrWorkspaceNotFound = shared.NewDomainError(apperrors.NotFound, "workspace not found")
	ErrNotOwner          = shared.NewDomainError(apperrors.Unauthorized, "only the owner can perform this action")
)

type WorkspaceID = shared.ID[Workspace]

type Workspace struct {
	id          WorkspaceID
	name        string
	description string
	members     map[userDomain.UserID]WorkspaceRole
	createdAt   time.Time

	events []shared.DomainEvent
}

func NewWorkspace(name, description string, creatorID userDomain.UserID) *Workspace {
	id := shared.NewID[Workspace]()
	now := time.Now()

	ws := &Workspace{
		id:          id,
		name:        name,
		description: description,
		members:     make(map[userDomain.UserID]WorkspaceRole),
		createdAt:   now,
	}

	ws.members[creatorID] = RoleOwner

	ws.RecordEvent(WorkspaceCreatedEvent{
		ID:       id,
		Name:     name,
		OwnerID:  creatorID,
		Occurred: now,
	})

	ws.RecordEvent(MemberAddedEvent{
		WsID:     id,
		UserID:   creatorID,
		Role:     RoleOwner,
		Occurred: now,
	})

	return ws
}

func ReconstituteWorkspace(
	id WorkspaceID,
	name string,
	description string,
	createdAt time.Time,
	members map[userDomain.UserID]WorkspaceRole,
) *Workspace {
	return &Workspace{
		id:          id,
		name:        name,
		description: description,
		members:     members,
		createdAt:   createdAt,
	}
}

func (w *Workspace) AddMember(userID userDomain.UserID, role WorkspaceRole) error {
	if _, exists := w.members[userID]; exists {
		return ErrUserAlreadyMember
	}

	w.members[userID] = role

	w.RecordEvent(MemberAddedEvent{
		WsID:     w.id,
		UserID:   userID,
		Role:     role,
		Occurred: time.Now(),
	})

	return nil
}

func (w *Workspace) IsOwner(userID userDomain.UserID) bool {
	role, exists := w.members[userID]
	return exists && role == RoleOwner
}

func (w *Workspace) RemoveMember(userID userDomain.UserID) error {
	role, exists := w.members[userID]
	if !exists {
		return ErrMemberNotFound
	}

	if role == RoleOwner && w.countOwners() == 1 {
		return ErrAtLeastOneOwner
	}

	delete(w.members, userID)

	w.RecordEvent(MemberRemovedEvent{
		WsID:     w.id,
		UserID:   userID,
		Occurred: time.Now(),
	})

	return nil
}

func (w *Workspace) countOwners() int {
	count := 0

	for _, role := range w.members {
		if role == RoleOwner {
			count++
		}
	}

	return count
}

func (w *Workspace) Delete() {
	w.RecordEvent(WorkspaceDeletedEvent{
		ID:       w.id,
		Occurred: time.Now(),
	})
}

func (w *Workspace) ID() WorkspaceID                              { return w.id }
func (w *Workspace) Name() string                                 { return w.name }
func (w *Workspace) Description() string                          { return w.description }
func (w *Workspace) CreatedAt() time.Time                         { return w.createdAt }
func (w *Workspace) Members() map[userDomain.UserID]WorkspaceRole { return w.members }
func (w *Workspace) Events() []shared.DomainEvent                 { return w.events }
func (w *Workspace) ClearEvents()                                 { w.events = nil }

func (w *Workspace) RecordEvent(e shared.DomainEvent) {
	w.events = append(w.events, e)
}
