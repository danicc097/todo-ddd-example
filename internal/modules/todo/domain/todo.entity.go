package domain

import (
	"errors"
	"time"

	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

var ErrTodoNotFound = errors.New("todo not found")

type TodoID = shared.ID[Todo]

type Todo struct {
	AggregateRoot

	id        TodoID
	title     TodoTitle
	status    TodoStatus
	tags      []TagID
	createdAt time.Time
}

func NewTodo(title TodoTitle) *Todo {
	id := shared.NewID[Todo]()
	now := time.Now()
	t := &Todo{
		id:        id,
		title:     title,
		status:    StatusPending,
		tags:      make([]TagID, 0),
		createdAt: now,
	}
	t.RecordEvent(TodoCreatedEvent{
		ID:        id,
		Title:     title.String(),
		Status:    StatusPending.String(),
		CreatedAt: now,
		Occurred:  now,
	})

	return t
}

func ReconstituteTodo(id TodoID, title TodoTitle, status TodoStatus, createdAt time.Time, tags []TagID) *Todo {
	return &Todo{
		id:        id,
		title:     title,
		status:    status,
		createdAt: createdAt,
		tags:      tags,
	}
}

func (t *Todo) Complete() error {
	if t.status == StatusArchived {
		return ErrInvalidStatus
	}

	t.status = StatusCompleted
	t.RecordEvent(TodoCompletedEvent{
		ID:        t.id,
		Title:     t.title.String(),
		Status:    t.status.String(),
		CreatedAt: t.createdAt,
		Occurred:  time.Now(),
	})

	return nil
}

func (t *Todo) AddTag(tagID TagID) {
	t.tags = append(t.tags, tagID)
	t.RecordEvent(TagAddedEvent{
		TodoID:   t.id,
		TagID:    tagID,
		Occurred: time.Now(),
	})
}

func (t *Todo) ID() TodoID           { return t.id }
func (t *Todo) Title() TodoTitle     { return t.title }
func (t *Todo) Status() TodoStatus   { return t.status }
func (t *Todo) CreatedAt() time.Time { return t.createdAt }
func (t *Todo) Tags() []TagID        { return t.tags }

// NOTE: entity should not know how it's serialized to the outside world (apis, messaging...)
// func (t *Todo) MarshalJSON() ([]byte, error) {
// 	...
// }
