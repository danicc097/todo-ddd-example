package postgres

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/application"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

/**
 * Todo
 */

type TodoMapper struct{}

func (m *TodoMapper) ToDomain(row db.GetTodoAggregateByIDRow) *domain.Todo {
	title, _ := domain.NewTodoTitle(row.Title)

	tagIDs := make([]domain.TagID, len(row.Tags))
	for i, id := range row.Tags {
		tagIDs[i] = domain.TagID(id)
	}

	recurrence := m.mapRecurrence(row.RecurrenceInterval, row.RecurrenceAmount)
	sessions := m.mapFocusSessionsDomain(row.FocusSessions)

	return domain.ReconstituteTodo(domain.ReconstituteTodoArgs{
		ID:              row.ID,
		Title:           title,
		Status:          domain.TodoStatus(row.Status),
		CreatedAt:       row.CreatedAt,
		Tags:            tagIDs,
		WorkspaceID:     row.WorkspaceID,
		DueDate:         row.DueDate,
		Recurrence:      recurrence,
		LastCompletedAt: row.LastCompletedAt,
		Sessions:        sessions,
	})
}

func (m *TodoMapper) ListRowToDomain(row db.ListTodosByWorkspaceIDRow) *domain.Todo {
	title, _ := domain.NewTodoTitle(row.Title)

	tagIDs := make([]domain.TagID, len(row.Tags))
	for i, id := range row.Tags {
		tagIDs[i] = domain.TagID(id)
	}

	recurrence := m.mapRecurrence(row.RecurrenceInterval, row.RecurrenceAmount)
	sessions := m.mapFocusSessionsDomain(row.FocusSessions)

	return domain.ReconstituteTodo(domain.ReconstituteTodoArgs{
		ID:              row.ID,
		Title:           title,
		Status:          domain.TodoStatus(row.Status),
		CreatedAt:       row.CreatedAt,
		Tags:            tagIDs,
		WorkspaceID:     row.WorkspaceID,
		DueDate:         row.DueDate,
		Recurrence:      recurrence,
		LastCompletedAt: row.LastCompletedAt,
		Sessions:        sessions,
	})
}

func (m *TodoMapper) mapRecurrence(interval *string, amount *int32) *domain.RecurrenceRule {
	if interval != nil && amount != nil {
		r, _ := domain.NewRecurrenceRule(*interval, int(*amount))
		return &r
	}

	return nil
}

func (m *TodoMapper) mapFocusSessions(raw any) []application.FocusSessionReadModel {
	var sessions []application.FocusSessionReadModel
	if raw == nil {
		return sessions
	}

	var (
		b   []byte
		err error
	)

	switch v := raw.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		b, err = json.Marshal(v)
	}

	if err == nil && len(b) > 0 {
		var rawSessions []struct {
			ID        uuid.UUID  `json:"id"`
			UserID    uuid.UUID  `json:"user_id"`
			StartTime time.Time  `json:"start_time"`
			EndTime   *time.Time `json:"end_time"`
		}
		if err := json.Unmarshal(b, &rawSessions); err == nil {
			for _, s := range rawSessions {
				sessions = append(sessions, application.FocusSessionReadModel{
					ID:        s.ID,
					StartTime: s.StartTime,
					EndTime:   s.EndTime,
				})
			}
		}
	}

	return sessions
}

func (m *TodoMapper) mapFocusSessionsDomain(raw any) []domain.FocusSession {
	var sessions []domain.FocusSession
	if raw == nil {
		return sessions
	}

	var (
		b   []byte
		err error
	)

	switch v := raw.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		b, err = json.Marshal(v)
	}

	if err == nil && len(b) > 0 {
		var rawSessions []struct {
			ID        uuid.UUID  `json:"id"`
			UserID    uuid.UUID  `json:"user_id"`
			StartTime time.Time  `json:"start_time"`
			EndTime   *time.Time `json:"end_time"`
		}
		if err := json.Unmarshal(b, &rawSessions); err == nil {
			for _, s := range rawSessions {
				sessions = append(sessions, domain.ReconstituteFocusSession(domain.ReconstituteFocusSessionArgs{
					ID:        domain.FocusSessionID(s.ID),
					UserID:    userDomain.UserID(s.UserID),
					StartTime: s.StartTime,
					EndTime:   s.EndTime,
				}))
			}
		}
	}

	return sessions
}

// ToPersistence maps Domain to the primary table struct.
func (m *TodoMapper) ToPersistence(t *domain.Todo) db.Todos {
	var (
		rInterval *string
		rAmount   *int32
	)

	if t.Recurrence() != nil {
		interval := t.Recurrence().Interval()
		amount := int32(t.Recurrence().Amount())
		rInterval = &interval
		rAmount = &amount
	}

	return db.Todos{
		ID:                 t.ID(),
		Title:              t.Title().String(),
		Status:             t.Status().String(),
		CreatedAt:          t.CreatedAt(),
		WorkspaceID:        t.WorkspaceID(),
		DueDate:            t.DueDate(),
		RecurrenceInterval: rInterval,
		RecurrenceAmount:   rAmount,
		LastCompletedAt:    t.LastCompletedAt(),
	}
}

/**
 * Events
 */

type TodoOutboxDTO struct {
	ID           domain.TodoID        `json:"id"`
	WorkspaceID  wsDomain.WorkspaceID `json:"workspace_id"`
	Title        string               `json:"title"`
	Status       string               `json:"status"`
	CreatedAt    time.Time            `json:"created_at"`
	EventVersion int                  `json:"event_version"`
}

type TodoRolledOverOutboxDTO struct {
	ID           domain.TodoID        `json:"id"`
	WorkspaceID  wsDomain.WorkspaceID `json:"workspace_id"`
	NewDueDate   time.Time            `json:"new_due_date"`
	ActorID      userDomain.UserID    `json:"actor_id"`
	EventVersion int                  `json:"event_version"`
}

type TagAddedOutboxDTO struct {
	TodoID       domain.TodoID        `json:"todo_id"`
	TagID        domain.TagID         `json:"tag_id"`
	WorkspaceID  wsDomain.WorkspaceID `json:"workspace_id"`
	EventVersion int                  `json:"event_version"`
}

type TodoDeletedOutboxDTO struct {
	ID           domain.TodoID        `json:"id"`
	WorkspaceID  wsDomain.WorkspaceID `json:"workspace_id"`
	EventVersion int                  `json:"event_version"`
}

func (m *TodoMapper) MapEvent(e shared.DomainEvent) (shared.EventType, any, error) {
	var payload any

	switch evt := e.(type) {
	case domain.TodoCreatedEvent:
		payload = TodoOutboxDTO{
			ID:           evt.ID,
			WorkspaceID:  evt.WsID,
			Title:        evt.Title.String(),
			Status:       evt.Status.String(),
			CreatedAt:    evt.CreatedAt,
			EventVersion: 1,
		}
	case domain.TodoCompletedEvent:
		payload = TodoOutboxDTO{
			ID:           evt.ID,
			WorkspaceID:  evt.WsID,
			Title:        evt.Title.String(),
			Status:       evt.Status.String(),
			CreatedAt:    evt.CreatedAt,
			EventVersion: 1,
		}
	case domain.TodoRolledOverEvent:
		payload = TodoRolledOverOutboxDTO{
			ID:           evt.ID,
			WorkspaceID:  evt.WsID,
			NewDueDate:   evt.NewDueDate,
			ActorID:      evt.ActorID,
			EventVersion: 1,
		}
	case domain.TodoDeletedEvent:
		payload = TodoDeletedOutboxDTO{
			ID:           evt.ID,
			WorkspaceID:  evt.WsID,
			EventVersion: 1,
		}
	case domain.TagAddedEvent:
		payload = TagAddedOutboxDTO{
			TodoID:       evt.TodoID,
			TagID:        evt.TagID,
			WorkspaceID:  evt.WsID,
			EventVersion: 1,
		}
	default:
		return "", nil, nil
	}

	return e.EventName(), payload, nil
}
