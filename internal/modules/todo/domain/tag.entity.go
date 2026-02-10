package domain

import (
	"errors"

	"github.com/google/uuid"
)

var ErrTagNotFound = errors.New("tag not found")

type Tag struct {
	id   uuid.UUID
	name TagName
}

func NewTag(name TagName) *Tag {
	return &Tag{
		id:   uuid.New(),
		name: name,
	}
}

func ReconstituteTag(id uuid.UUID, name TagName) *Tag {
	return &Tag{id: id, name: name}
}

func (t *Tag) ID() uuid.UUID   { return t.id }
func (t *Tag) Name() TagName { return t.name }
