package domain

import (
	"time"

	"github.com/danicc097/todo-ddd-example/internal/apperrors"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

var ErrTodoNotFound = shared.NewDomainError(apperrors.NotFound, "todo not found")

type TodoID = shared.ID[Todo]

type Todo struct {
	shared.AggregateRoot

	id              TodoID
	workspaceID     wsDomain.WorkspaceID
	title           TodoTitle
	status          TodoStatus
	dueDate         *time.Time
	recurrence      *RecurrenceRule
	lastCompletedAt *time.Time
	sessions        []FocusSession
	tags            []TagID
	createdAt       time.Time
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
		sessions:    make([]FocusSession, 0),
		createdAt:   now,
	}
	t.RecordEvent(TodoCreatedEvent{
		ID:        id,
		WsID:      workspaceID,
		Title:     title,
		Status:    StatusPending,
		CreatedAt: now,
		Occurred:  now,
	})

	return t
}

type ReconstituteTodoArgs struct {
	ID              TodoID
	WorkspaceID     wsDomain.WorkspaceID
	Title           TodoTitle
	Status          TodoStatus
	CreatedAt       time.Time
	Tags            []TagID
	DueDate         *time.Time
	Recurrence      *RecurrenceRule
	LastCompletedAt *time.Time
	Sessions        []FocusSession
}

func ReconstituteTodo(args ReconstituteTodoArgs) *Todo {
	return &Todo{
		id:              args.ID,
		workspaceID:     args.WorkspaceID,
		title:           args.Title,
		status:          args.Status,
		createdAt:       args.CreatedAt,
		tags:            args.Tags,
		dueDate:         args.DueDate,
		recurrence:      args.Recurrence,
		lastCompletedAt: args.LastCompletedAt,
		sessions:        args.Sessions,
	}
}

func (t *Todo) Complete(actorID userDomain.UserID, now time.Time) error {
	if t.status == StatusArchived {
		return ErrInvalidStatus
	}

	if t.recurrence != nil {
		if t.dueDate != nil && t.dueDate.After(now) {
			return ErrCannotCompleteFutureOccurrence
		}

		// calculate based on previous due date to retain cadence
		baseDate := now
		if t.dueDate != nil {
			baseDate = *t.dueDate
		}

		nextDate := t.recurrence.CalculateNext(baseDate)
		t.dueDate = &nextDate
		t.lastCompletedAt = &now

		t.RecordEvent(TodoRolledOverEvent{
			ID:         t.id,
			WsID:       t.workspaceID,
			NewDueDate: nextDate,
			Occurred:   now,
			ActorID:    actorID,
		})

		return nil
	}

	t.status = StatusCompleted
	t.lastCompletedAt = &now
	t.RecordEvent(TodoCompletedEvent{
		ID:        t.id,
		WsID:      t.workspaceID,
		Title:     t.title,
		Status:    t.status,
		CreatedAt: t.createdAt,
		Occurred:  now,
		ActorID:   actorID,
	})

	return nil
}

func (t *Todo) StartFocus(userID userDomain.UserID, sessionID FocusSessionID) error {
	if t.status == StatusCompleted || t.status == StatusArchived {
		return ErrCannotFocusOnCompletedTask
	}

	for _, s := range t.sessions {
		if s.IsActive() {
			return ErrFocusSessionAlreadyActive
		}
	}

	t.sessions = append(t.sessions, NewFocusSession(sessionID, userID, time.Now()))

	return nil
}

func (t *Todo) StopFocus(now time.Time) error {
	for i, s := range t.sessions {
		if s.IsActive() {
			if now.Before(s.startTime) {
				return ErrInvalidFocusStopTimeAfter
			}

			t.sessions[i].endTime = &now

			return nil
		}
	}

	return ErrNoActiveFocusSession
}

func (t *Todo) ActiveFocusSession() *FocusSession {
	for _, s := range t.sessions {
		if s.IsActive() {
			return &s
		}
	}

	return nil
}

func (t *Todo) AddTag(tagID TagID) {
	t.tags = append(t.tags, tagID)
	t.RecordEvent(TagAddedEvent{
		TodoID:   t.id,
		TagID:    tagID,
		WsID:     t.workspaceID,
		Occurred: time.Now(),
	})
}

func (t *Todo) Delete() {
	t.RecordEvent(TodoDeletedEvent{
		ID:       t.id,
		WsID:     t.workspaceID,
		Occurred: time.Now(),
	})
}

func (t *Todo) SetDueDate(d *time.Time) {
	t.dueDate = d
}

func (t *Todo) SetRecurrence(r *RecurrenceRule) {
	t.recurrence = r
}

func (t *Todo) ID() TodoID                        { return t.id }
func (t *Todo) WorkspaceID() wsDomain.WorkspaceID { return t.workspaceID }
func (t *Todo) Title() TodoTitle                  { return t.title }
func (t *Todo) Status() TodoStatus                { return t.status }
func (t *Todo) CreatedAt() time.Time              { return t.createdAt }
func (t *Todo) Tags() []TagID                     { return t.tags }
func (t *Todo) DueDate() *time.Time               { return t.dueDate }
func (t *Todo) Recurrence() *RecurrenceRule       { return t.recurrence }
func (t *Todo) LastCompletedAt() *time.Time       { return t.lastCompletedAt }
func (t *Todo) Sessions() []FocusSession          { return t.sessions }

// NOTE: entity should not know how it's serialized to the outside world (apis, messaging...)
// func (t *Todo) MarshalJSON() ([]byte, error) {
// 	...
// }
