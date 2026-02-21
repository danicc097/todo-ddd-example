package domain

import (
	"time"

	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

var ErrTodoNotFound = shared.NewDomainError(apperrors.NotFound, "todo not found")

type TodoID = shared.ID[Todo]

type Todo struct {
	AggregateRoot

	id          TodoID
	workspaceID wsDomain.WorkspaceID
	title       TodoTitle
	status      TodoStatus
	tags        []TagID
	createdAt   time.Time
}

func NewTodo(title TodoTitle, workspaceID wsDomain.WorkspaceID) *Todo {
	id := shared.NewID[Todo]()
	now := time.Now()
	t := &Todo{
		id:          id,
		workspaceID: workspaceID,
		title:       title,
		status:      StatusPending,
		tags:        make([]TagID, 0),
		createdAt:   now,
	}
	t.RecordEvent(TodoCreatedEvent{
		ID:          id,
		WorkspaceID: workspaceID,
		Title:       title.String(),
		Status:      StatusPending.String(),
		CreatedAt:   now,
		Occurred:    now,
	})

	return t
}

func ReconstituteTodo(id TodoID, title TodoTitle, status TodoStatus, createdAt time.Time, tags []TagID, workspaceID wsDomain.WorkspaceID) *Todo {
	return &Todo{
		id:          id,
		workspaceID: workspaceID,
		title:       title,
		status:      status,
		createdAt:   createdAt,
		tags:        tags,
	}
}

func (t *Todo) Complete() error {
	if t.status == StatusArchived {
		return ErrInvalidStatus
	}

	t.status = StatusCompleted
	t.RecordEvent(TodoCompletedEvent{
		ID:          t.id,
		WorkspaceID: t.workspaceID,
		Title:       t.title.String(),
		Status:      t.status.String(),
		CreatedAt:   t.createdAt,
		Occurred:    time.Now(),
	})

	return nil
}

func (t *Todo) AddTag(tagID TagID) {
	t.tags = append(t.tags, tagID)
	t.RecordEvent(TagAddedEvent{
		TodoID:      t.id,
		TagID:       tagID,
		WorkspaceID: t.workspaceID,
		Occurred:    time.Now(),
	})
}

func (t *Todo) ID() TodoID                        { return t.id }
func (t *Todo) WorkspaceID() wsDomain.WorkspaceID { return t.workspaceID }
func (t *Todo) Title() TodoTitle                  { return t.title }
func (t *Todo) Status() TodoStatus                { return t.status }
func (t *Todo) CreatedAt() time.Time              { return t.createdAt }
func (t *Todo) Tags() []TagID                     { return t.tags }

// NOTE: entity should not know how it's serialized to the outside world (apis, messaging...)
// func (t *Todo) MarshalJSON() ([]byte, error) {
// 	...
// }
