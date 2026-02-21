package domain

import (
	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

var ErrTagNotFound = shared.NewDomainError(apperrors.NotFound, "tag not found")

type TagID = shared.ID[Tag]

type Tag struct {
	id          TagID
	name        TagName
	workspaceID wsDomain.WorkspaceID
}

func NewTag(name TagName, workspaceID wsDomain.WorkspaceID) *Tag {
	return &Tag{
		id:          shared.NewID[Tag](),
		name:        name,
		workspaceID: workspaceID,
	}
}

func ReconstituteTag(id TagID, name TagName, workspaceID wsDomain.WorkspaceID) *Tag {
	return &Tag{id: id, name: name, workspaceID: workspaceID}
}

func (t *Tag) ID() TagID                         { return t.id }
func (t *Tag) Name() TagName                     { return t.name }
func (t *Tag) WorkspaceID() wsDomain.WorkspaceID { return t.workspaceID }
