package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"

	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

var (
	ErrAtLeastOneOwner   = errors.New("workspace must have at least one owner")
	ErrUserAlreadyMember = errors.New("user is already a member")
	ErrMemberNotFound    = errors.New("member not found")
	ErrWorkspaceNotFound = errors.New("workspace not found")
)

type Workspace struct {
	id          uuid.UUID
	name        string
	description string
	members     map[uuid.UUID]WorkspaceRole
	createdAt   time.Time

	events []shared.DomainEvent
}

func NewWorkspace(name, description string, creatorID uuid.UUID) *Workspace {
	w := &Workspace{
		id:          uuid.New(),
		name:        name,
		description: description,
		members:     make(map[uuid.UUID]WorkspaceRole),
		createdAt:   time.Now(),
	}

	w.members[creatorID] = RoleOwner

	return w
}

func ReconstituteWorkspace(
	id uuid.UUID,
	name string,
	description string,
	createdAt time.Time,
	members map[uuid.UUID]WorkspaceRole,
) *Workspace {
	return &Workspace{
		id:          id,
		name:        name,
		description: description,
		members:     members,
		createdAt:   createdAt,
	}
}

func (w *Workspace) AddMember(userID uuid.UUID, role WorkspaceRole) error {
	if _, exists := w.members[userID]; exists {
		return ErrUserAlreadyMember
	}

	w.members[userID] = role

	return nil
}

func (w *Workspace) RemoveMember(userID uuid.UUID) error {
	role, exists := w.members[userID]
	if !exists {
		return ErrMemberNotFound
	}

	if role == RoleOwner && w.countOwners() == 1 {
		return ErrAtLeastOneOwner
	}

	delete(w.members, userID)

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

func (w *Workspace) ID() uuid.UUID                        { return w.id }
func (w *Workspace) Name() string                         { return w.name }
func (w *Workspace) Description() string                  { return w.description }
func (w *Workspace) CreatedAt() time.Time                 { return w.createdAt }
func (w *Workspace) Members() map[uuid.UUID]WorkspaceRole { return w.members }
func (w *Workspace) Events() []shared.DomainEvent         { return w.events }
func (w *Workspace) ClearEvents()                         { w.events = nil }
