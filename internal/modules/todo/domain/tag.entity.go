package domain

import (
	"errors"

	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

var ErrTagNotFound = errors.New("tag not found")

type TagID = shared.ID[Tag]

type Tag struct {
	id   TagID
	name TagName
}

func NewTag(name TagName) *Tag {
	return &Tag{
		id:   shared.NewID[Tag](),
		name: name,
	}
}

func ReconstituteTag(id TagID, name TagName) *Tag {
	return &Tag{id: id, name: name}
}

func (t *Tag) ID() TagID   { return t.id }
func (t *Tag) Name() TagName { return t.name }
