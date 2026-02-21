package postgres

import (
	"time"

	"github.com/danicc097/todo-ddd-example/internal/generated/db"
	"github.com/danicc097/todo-ddd-example/internal/modules/todo/domain"
	wsDomain "github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
	shared "github.com/danicc097/todo-ddd-example/internal/shared/domain"
)

/**
 * Todo
 */

type TodoMapper struct{}

func (m *TodoMapper) ToDomain(row db.GetTodoByIDRow) *domain.Todo {
	title, _ := domain.NewTodoTitle(row.Title)

	tagIDs := make([]domain.TagID, len(row.Tags))
	for i, id := range row.Tags {
		tagIDs[i] = domain.TagID(id)
	}

	return domain.ReconstituteTodo(
		row.ID,
		title,
		domain.TodoStatus(row.Status),
		row.CreatedAt,
		tagIDs,
		row.WorkspaceID,
	)
}

func (m *TodoMapper) ListRowToDomain(row db.ListTodosByWorkspaceIDRow) *domain.Todo {
	title, _ := domain.NewTodoTitle(row.Title)

	tagIDs := make([]domain.TagID, len(row.Tags))
	for i, id := range row.Tags {
		tagIDs[i] = domain.TagID(id)
	}

	return domain.ReconstituteTodo(
		row.ID,
		title,
		domain.TodoStatus(row.Status),
		row.CreatedAt,
		tagIDs,
		row.WorkspaceID,
	)
}

// ToPersistence maps Domain to the primary table struct.
func (m *TodoMapper) ToPersistence(t *domain.Todo) db.Todos {
	return db.Todos{
		ID:          t.ID(),
		Title:       t.Title().String(),
		Status:      t.Status().String(),
		CreatedAt:   t.CreatedAt(),
		WorkspaceID: t.WorkspaceID(),
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

type TagAddedOutboxDTO struct {
	TodoID       domain.TodoID        `json:"todo_id"`
	TagID        domain.TagID         `json:"tag_id"`
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
