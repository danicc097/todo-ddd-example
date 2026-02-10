package domain

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrTodoNotFound = errors.New("todo not found")

type Todo struct {
	AggregateRoot
	id        uuid.UUID
	title     TodoTitle
	status    TodoStatus
	tags      []uuid.UUID
	createdAt time.Time
}

func NewTodo(title TodoTitle) *Todo {
	id := uuid.New()
	now := time.Now()
	t := &Todo{
		id:        id,
		title:     title,
		status:    StatusPending,
		tags:      make([]uuid.UUID, 0),
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

func ReconstituteTodo(id uuid.UUID, title TodoTitle, status TodoStatus, createdAt time.Time, tags []uuid.UUID) *Todo {
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

func (t *Todo) AddTag(tagID uuid.UUID) {
	t.tags = append(t.tags, tagID)
	t.RecordEvent(TagAddedEvent{
		TodoID:   t.id,
		TagID:    tagID,
		Occurred: time.Now(),
	})
}

func (t *Todo) ID() uuid.UUID        { return t.id }
func (t *Todo) Title() TodoTitle     { return t.title }
func (t *Todo) Status() TodoStatus   { return t.status }
func (t *Todo) CreatedAt() time.Time { return t.createdAt }
func (t *Todo) Tags() []uuid.UUID    { return t.tags }

func (t *Todo) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID        uuid.UUID   `json:"id"`
		Title     string      `json:"title"`
		Status    TodoStatus  `json:"status"`
		Tags      []uuid.UUID `json:"tags"`
		CreatedAt time.Time   `json:"created_at"`
	}{
		ID:        t.id,
		Title:     t.title.String(),
		Status:    t.status,
		Tags:      t.tags,
		CreatedAt: t.createdAt,
	})
}
