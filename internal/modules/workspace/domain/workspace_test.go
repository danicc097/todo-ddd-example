package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
)

func TestNewWorkspace(t *testing.T) {
	t.Parallel()

	creatorID := userDomain.UserID(uuid.New())
	ws := NewWorkspace("Test Workspace", "Description", creatorID)

	assert.NotNil(t, ws)
	assert.Equal(t, "Test Workspace", ws.Name())
	assert.Contains(t, ws.Members(), creatorID)
	assert.Equal(t, RoleOwner, ws.Members()[creatorID])

	events := ws.Events()
	assert.Len(t, events, 2)
	assert.IsType(t, WorkspaceCreatedEvent{}, events[0])
	assert.IsType(t, MemberAddedEvent{}, events[1])
}

func TestWorkspace_AddMember(t *testing.T) {
	t.Parallel()

	creatorID := userDomain.UserID(uuid.New())
	ws := NewWorkspace("WS", "Desc", creatorID)
	ws.ClearEvents()

	newMemberID := userDomain.UserID(uuid.New())
	err := ws.AddMember(newMemberID, RoleMember)

	assert.NoError(t, err)
	assert.Contains(t, ws.Members(), newMemberID)
	assert.Equal(t, RoleMember, ws.Members()[newMemberID])

	events := ws.Events()
	assert.Len(t, events, 1)
	assert.IsType(t, MemberAddedEvent{}, events[0])
	evt := events[0].(MemberAddedEvent)
	assert.Equal(t, newMemberID, evt.UserID)
}

func TestWorkspace_AddMember_AlreadyExists(t *testing.T) {
	t.Parallel()

	creatorID := userDomain.UserID(uuid.New())
	ws := NewWorkspace("WS", "Desc", creatorID)

	err := ws.AddMember(creatorID, RoleMember)
	assert.ErrorIs(t, err, ErrUserAlreadyMember)
}

func TestWorkspace_RemoveMember(t *testing.T) {
	t.Parallel()

	creatorID := userDomain.UserID(uuid.New())
	ws := NewWorkspace("WS", "Desc", creatorID)

	memberID := userDomain.UserID(uuid.New())
	_ = ws.AddMember(memberID, RoleMember)
	ws.ClearEvents()

	err := ws.RemoveMember(memberID)
	assert.NoError(t, err)
	assert.NotContains(t, ws.Members(), memberID)

	events := ws.Events()
	assert.Len(t, events, 1)
	assert.IsType(t, MemberRemovedEvent{}, events[0])
}

func TestWorkspace_RemoveMember_NotFound(t *testing.T) {
	t.Parallel()

	creatorID := userDomain.UserID(uuid.New())
	ws := NewWorkspace("WS", "Desc", creatorID)

	err := ws.RemoveMember(userDomain.UserID(uuid.New()))
	assert.ErrorIs(t, err, ErrMemberNotFound)
}

func TestWorkspace_RemoveMember_LastOwner(t *testing.T) {
	t.Parallel()

	creatorID := userDomain.UserID(uuid.New())
	ws := NewWorkspace("WS", "Desc", creatorID)

	err := ws.RemoveMember(creatorID)
	assert.ErrorIs(t, err, ErrAtLeastOneOwner)
}
