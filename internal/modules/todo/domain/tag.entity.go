package domain

import (
	"time"

	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

var ErrTagNotFound = shared.NewDomainError(apperrors.NotFound, "tag not found")

type TagID = shared.ID[Tag]

type Tag struct {
	shared.AggregateRoot

	id          TagID
	name        TagName
	workspaceID wsDomain.WorkspaceID
}

func NewTag(name TagName, workspaceID wsDomain.WorkspaceID) *Tag {
	id := shared.NewID[Tag]()
	t := &Tag{
		id:          id,
		name:        name,
		workspaceID: workspaceID,
	}

	t.RecordEvent(TagCreatedEvent{
		ID:       id,
		Name:     name,
		WsID:     workspaceID,
		Occurred: time.Now(),
	})

	return t
}

type ReconstituteTagArgs struct {
	ID          TagID
	Name        TagName
	WorkspaceID wsDomain.WorkspaceID
}

func ReconstituteTag(args ReconstituteTagArgs) *Tag {
	return &Tag{id: args.ID, name: args.Name, workspaceID: args.WorkspaceID}
}

func (t *Tag) ID() TagID                         { return t.id }
func (t *Tag) Name() TagName                     { return t.name }
func (t *Tag) WorkspaceID() wsDomain.WorkspaceID { return t.workspaceID }
